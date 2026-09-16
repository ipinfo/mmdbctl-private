package compress

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	dataSectionSeparatorSize = 16
)

// TODO
type Options struct {
	DisableCompact     bool
	LogMemorySnapshots bool
}

// Compress reads the mmdb file at inputPath, deduplicates identical subtrees of
// its search trie, and writes the smaller result to outputPath. The input file
// is left untouched.
func Compress(logger *slog.Logger, inputPath, outputPath string, opts Options) (CompressResult, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	result, err := optimize(logger, inputPath, outputPath, opts.LogMemorySnapshots, opts.DisableCompact)
	if err != nil {
		return result, err
	}
	return result, nil
}

// TODO
type CompressResult struct {
	InputBytes      uint64
	OutputBytes     uint64
	InputNodeCount  uint32
	OutputNodeCount uint32
	InputTreeBytes  uint64
	OutputTreeBytes uint64
	DataBytes       uint64
	OutputDataBytes uint64
	MetadataBytes   uint64
	Elapsed         time.Duration
}

func optimize(logger *slog.Logger,
	inputPath, outputPath string,
	logMemorySnapshots bool,
	disableCompact bool,
) (CompressResult, error) {
	startTime := time.Now()
	result := CompressResult{}

	logger.Debug(fmt.Sprintf("read: %s", inputPath))
	timeRead := time.Now()
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return result, fmt.Errorf("Failure reading input file: %w", err)
	}
	result.InputBytes = uint64(len(input))
	logger.Debug(fmt.Sprintf("read: %s loaded in %s",
		HumanBytes(result.InputBytes),
		time.Since(timeRead).Round(time.Millisecond),
	))

	if logMemorySnapshots {
		logMemory(logger, "after-read")
	}

	metaStart, metaBytes, err := findMetadata(input)
	if err != nil {
		return result, fmt.Errorf("Failure finding metadata: %w", err)
	}

	meta, err := decodeMetadata(metaBytes)
	if err != nil {
		return result, fmt.Errorf("Failure decoding metadata: %w", err)
	}

	nodeCount, ok := meta["node_count"].(uint32)
	if !ok {
		return result, fmt.Errorf("metadata.node_count missing or wrong type (got %T)", meta["node_count"])
	}

	recordSize, ok := meta["record_size"].(uint16)
	if !ok {
		return result, fmt.Errorf("metadata.record_size missing or wrong type (got %T)", meta["record_size"])
	}
	if recordSize != 24 && recordSize != 28 && recordSize != 32 {
		return result, fmt.Errorf("unsupported record_size %d", recordSize)
	}
	result.InputNodeCount = nodeCount
	nodeBytes := uint64(recordSize) / 4
	treeBytes := uint64(nodeCount) * nodeBytes
	result.InputTreeBytes = treeBytes

	if metaStart < treeBytes+dataSectionSeparatorSize {
		return result, fmt.Errorf("metadata starts before end of tree+separator")
	}
	dataStart := treeBytes + dataSectionSeparatorSize
	dataEnd := metaStart
	result.DataBytes = dataEnd - dataStart
	result.MetadataBytes = uint64(result.InputBytes - metaStart)
	logger.Debug(fmt.Sprintf("layout: nodes=%d record_size=%d tree=%s data=%s meta=%s",
		nodeCount,
		recordSize,
		HumanBytes(result.InputTreeBytes),
		HumanBytes(result.DataBytes),
		HumanBytes(result.MetadataBytes),
	))

	tree := input[:treeBytes]
	dataSection := input[dataStart:dataEnd]

	// Phase 1: bottom-up canonical-ID assignment.
	logger.Debug(fmt.Sprintf("Phase 1: canonicalize starting (%d input nodes)", nodeCount))
	canonicalizeStartTime := time.Now()
	canon, err := canonicalize(logger, tree, nodeCount, uint64(recordSize))
	if err != nil {
		return result, fmt.Errorf("Failure during canonicalization: %w", err)
	}
	result.OutputNodeCount = uint32(len(canon.outNodes))
	result.OutputTreeBytes = uint64(result.OutputNodeCount) * nodeBytes
	logger.Debug(fmt.Sprintf("phase1: canonicalize done in %s — %d -> %d nodes (%+.2f%%)",
		time.Since(canonicalizeStartTime).Round(time.Millisecond),
		nodeCount,
		result.OutputNodeCount,
		100.0*float64(int64(result.OutputNodeCount)-int64(nodeCount))/float64(nodeCount),
	))
	if logMemorySnapshots {
		logMemory(logger, "after-canon")
	}

	// Phase 1.5: data-section compaction. Find every record reachable from
	// the canonical tree's leaf pointers (transitively, including through
	// MMDB pointers/kind 1), emit them contiguously into a new data section,
	// rewrite tree leaf pointers with the new offsets.
	emittedData := dataSection
	if !disableCompact {
		compactBytes, err := compact(logger, &canon, dataSection)
		if err != nil {
			// We don't format error as the error we receive is already formatted
			return result, err
		}
		emittedData = compactBytes
		if logMemorySnapshots {
			logMemory(logger, "after-compact")
		}
	}

	// Phase 2: emit the new tree with rewritten pointers. Output node indices
	// are assigned in canonicalization order, with the root pinned at index 0.
	logger.Debug(fmt.Sprintf("phase2: emit tree (%s)", HumanBytes(result.OutputTreeBytes)))
	emitStartTime := time.Now()
	newTree, err := emitTree(canon, uint64(recordSize), result.OutputNodeCount)
	if uint64(len(newTree)) != result.OutputTreeBytes {
		return result, fmt.Errorf("tree size mismatch: got %d want %d", len(newTree), result.OutputTreeBytes)
	}
	logger.Debug(fmt.Sprintf("phase2: emit tree done in %s", time.Since(emitStartTime).Round(time.Millisecond)))

	// Phase 3: rewrite metadata's node_count. Other fields are preserved.
	logger.Debug("phase3: rewrite metadata")
	newMeta := make(map[string]any, len(meta))
	maps.Copy(newMeta, meta)

	newMeta["node_count"] = result.OutputNodeCount
	newMetaBytes, err := encodeMetadata(newMeta)
	if err != nil {
		return result, fmt.Errorf("encode metadata: %w", err)
	}

	// Phase 4: assemble the output file.
	logger.Debug(fmt.Sprintf("phase4: write %s", outputPath))
	writeStartTime := time.Now()
	var out bytes.Buffer
	out.Grow(len(newTree) + dataSectionSeparatorSize + len(emittedData) + len(metadataStartMarker) + len(newMetaBytes))
	out.Write(newTree)
	out.Write(make([]byte, dataSectionSeparatorSize))
	out.Write(emittedData)
	out.Write(metadataStartMarker)
	out.Write(newMetaBytes)
	result.OutputBytes = uint64(out.Len())
	result.OutputDataBytes = uint64(len(emittedData))

	if dir := filepath.Dir(outputPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return result, fmt.Errorf("create output dir: %w", err)
		}
	}
	if err := os.WriteFile(outputPath, out.Bytes(), 0o644); err != nil {
		return result, fmt.Errorf("write output: %w", err)
	}
	logger.Debug(fmt.Sprintf("phase4: wrote %s in %s",
		HumanBytes(uint64(result.OutputBytes)),
		time.Since(writeStartTime).Round(time.Millisecond),
	))
	result.Elapsed = time.Since(startTime)

	return result, nil
}

// TODO: Maybe we can move this in compact.go
func compact(logger *slog.Logger, canon *canonResult, dataSection []byte) ([]byte, error) {
	rootSet := make(map[uint32]struct{}, len(canon.outNodes))
	for _, n := range canon.outNodes {
		if n.left.kind == dataKind {
			rootSet[n.left.id] = struct{}{}
		}
		if n.right.kind == dataKind {
			rootSet[n.right.id] = struct{}{}
		}
	}
	roots := make([]uint32, 0, len(rootSet))
	for o := range rootSet {
		roots = append(roots, o)
	}
	logger.Debug(fmt.Sprintf("phase1.5: data-section compaction (%s -> ?)", HumanBytes(uint64(len(dataSection)))))

	cmp, err := compactDataSection(logger, dataSection, roots)
	if err != nil {
		return nil, fmt.Errorf("compact: %w", err)
	}
	// Rewrite tree-leaf pointer ids to reference new offsets.
	for i := range canon.outNodes {
		if canon.outNodes[i].left.kind == dataKind {
			newOff, ok := cmp.offsetMap[canon.outNodes[i].left.id]
			if !ok {
				return nil, fmt.Errorf("missing offsetMap entry for tree leaf left id %d", canon.outNodes[i].left.id)
			}
			canon.outNodes[i].left.id = newOff
		}
		if canon.outNodes[i].right.kind == dataKind {
			newOff, ok := cmp.offsetMap[canon.outNodes[i].right.id]
			if !ok {
				return nil, fmt.Errorf("missing offsetMap entry for tree leaf right id %d", canon.outNodes[i].right.id)
			}
			canon.outNodes[i].right.id = newOff
		}
	}
	return cmp.bytes, nil
}

func logMemory(logger *slog.Logger, label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Debug(fmt.Sprintf("mem[%s]: alloc=%s sys=%s heap_in_use=%s",
		label, HumanBytes(m.Alloc), HumanBytes(m.Sys), HumanBytes(m.HeapInuse)))
}

func emitTree(c canonResult, recordSize uint64, outNodeCount uint32) ([]byte, error) {
	nodeBytes := recordSize / 4
	out := make([]byte, uint64(outNodeCount)*nodeBytes)
	for i := range outNodeCount {
		left := pointerToTreeValue(c.outNodes[i].left, outNodeCount)
		right := pointerToTreeValue(c.outNodes[i].right, outNodeCount)
		if err := writeNodePair(out, i, recordSize, left, right); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func pointerToTreeValue(p pointer, outNodeCount uint32) uint32 {
	switch p.kind {
	case nullKind:
		return outNodeCount
	case nodeKind:
		return p.id
	case dataKind:
		return p.id + outNodeCount + dataSectionSeparatorSize
	}
	return outNodeCount
}

func writeNodePair(buf []byte, idx uint32, recordSize uint64, left, right uint32) error {
	nodeBytes := recordSize / 4
	off := uint64(idx) * nodeBytes
	if off+nodeBytes > uint64(len(buf)) {
		return fmt.Errorf("write node %d offset %d > buf %d", idx, off, len(buf))
	}
	b := buf[off:]
	switch recordSize {
	case 24:
		if left>>24 != 0 || right>>24 != 0 {
			return fmt.Errorf("node %d: 24-bit record overflow (l=%d r=%d)", idx, left, right)
		}
		b[0] = byte(left >> 16)
		b[1] = byte(left >> 8)
		b[2] = byte(left)
		b[3] = byte(right >> 16)
		b[4] = byte(right >> 8)
		b[5] = byte(right)
	case 28:
		if left>>28 != 0 || right>>28 != 0 {
			return fmt.Errorf("node %d: 28-bit record overflow (l=%d r=%d)", idx, left, right)
		}
		b[0] = byte(left >> 16)
		b[1] = byte(left >> 8)
		b[2] = byte(left)
		b[3] = byte((left>>20)&0xf0) | byte((right>>24)&0x0f)
		b[4] = byte(right >> 16)
		b[5] = byte(right >> 8)
		b[6] = byte(right)
	case 32:
		binary.BigEndian.PutUint32(b[0:4], left)
		binary.BigEndian.PutUint32(b[4:8], right)
	default:
		return fmt.Errorf("unsupported record_size %d", recordSize)
	}
	return nil
}
