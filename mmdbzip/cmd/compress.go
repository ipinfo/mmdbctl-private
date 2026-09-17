package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ipinfo/cli/lib/complete"
	"github.com/ipinfo/cli/lib/complete/predict"
	"github.com/ipinfo/mmdbctl/mmdbzip/lib/compress"
)

var completionsCompress = &complete.Command{
	Flags: map[string]complete.Predictor{
		"-h":     predict.Nothing,
		"--help": predict.Nothing,
	},
}

func printHelpCompress(prog string) {
	fmt.Printf(
		`Usage: %s compress [<opts>] <input_mmdb_file> <output_mmdb_file>

Description:
  Deduplicate identical subtrees of the search trie and write a smaller mmdb
  file with identical lookup semantics. The input file is left untouched.

Options:
  General:
    --help, -h
      show help.
`, prog)
}

func cmdCompress(prog string, args []string) error {
	var help bool
	var dryRun bool
	var verbose bool
	var disableCompact bool
	var overwrite bool

	fs := newFlagSet("compress")
	fs.BoolVar(&dryRun, "dry-run", false, "skip writing output; just sanity-check the canonicalization")
	fs.BoolVarP(&verbose, "verbose", "v", false, "verbose: log phase boundaries, memory snapshots, periodic finalize counts")
	fs.BoolVar(&disableCompact, "no-compact", false, "skip data-section compaction; copy data section verbatim (debug)")
	fs.BoolVarP(&overwrite, "overwrite", "o", false, "overwrite output file if it already exists, defaults to false")
	fs.BoolVarP(&help, "help", "h", false, "show help.")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if help || len(rest) == 0 {
		printHelpCompress(prog)
		return nil
	}
	if len(rest) < 2 && !dryRun {
		return errors.New("Missing output file path")
	}

	inputPath := rest[0]
	var outputPath string
	if dryRun {
		outputPath = os.DevNull
	} else {
		outputPath = rest[1]
	}

	if _, err := os.Stat(outputPath); err == nil && !overwrite {
		return errors.New("Output file exists, use --overwrite to overwrite it")
	}

	logger := newLogger(verbose)
	opts := compress.Options{
		DisableCompact:     disableCompact,
		LogMemorySnapshots: verbose,
	}

	result, err := compress.Compress(logger, inputPath, outputPath, opts)
	if err != nil {
		return err
	}
	printSummary(inputPath, outputPath, result)
	return nil
}

func printSummary(inputPath, outputPath string, res compress.CompressResult) {
	fmt.Println()
	fmt.Printf("input:   %s\n", inputPath)
	fmt.Printf("output:  %s\n", outputPath)
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  size:\t%s\t->\t%s\t%s\n",
		compress.HumanBytes(res.InputBytes),
		compress.HumanBytes(res.OutputBytes),
		formatSizeDelta(int64(res.InputBytes), int64(res.OutputBytes)),
	)
	nodesDiff := int64(res.OutputNodeCount) - int64(res.InputNodeCount)
	fmt.Fprintf(w, "  nodes:\t%s\t->\t%s\t%s\n",
		commaInt(uint64(res.InputNodeCount)),
		commaInt(uint64(res.OutputNodeCount)),
		formatPctDelta(nodesDiff, int64(res.InputNodeCount)),
	)
	fmt.Fprintf(w, "  tree:\t%s\t->\t%s\t\n",
		compress.HumanBytes(res.InputTreeBytes),
		compress.HumanBytes(res.OutputTreeBytes),
	)
	bytesDiff := int64(res.OutputDataBytes) - int64(res.InputDataBytes)
	fmt.Fprintf(w, "  data:\t%s\t->\t%s\t%s\n",
		compress.HumanBytes(res.InputDataBytes),
		compress.HumanBytes(res.OutputDataBytes),
		formatPctDelta(bytesDiff, int64(res.InputDataBytes)),
	)
	w.Flush()

	fmt.Println()
	fmt.Printf("elapsed: %s\n", res.Elapsed.Round(time.Millisecond))
}

// formatPctDelta returns "(-65.78%)", "(+5.40%)", or "(unchanged)".
func formatPctDelta(diff, base int64) string {
	if diff == 0 || base == 0 {
		return "(unchanged)"
	}
	pct := 100.0 * float64(diff) / float64(base)
	if diff < 0 {
		return fmt.Sprintf("(%.2f%%)", pct)
	}
	return fmt.Sprintf("(+%.2f%%)", pct)
}

// formatSizeDelta returns "(saved 19.12 MiB, -65.76%)" when output is
// smaller, "(grew 2.13 MiB, +5.40%)" when larger, or "(unchanged)".
func formatSizeDelta(in, out int64) string {
	diff := out - in
	if diff == 0 {
		return "(unchanged)"
	}
	pct := 100.0 * float64(diff) / float64(in)
	if diff < 0 {
		return fmt.Sprintf("(saved %s, %.2f%%)", compress.HumanBytes(uint64(-diff)), pct)
	}
	return fmt.Sprintf("(grew %s, +%.2f%%)", compress.HumanBytes(uint64(diff)), pct)
}

// commaInt formats an integer with comma thousands separators.
func commaInt(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		b.WriteByte(',')
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
