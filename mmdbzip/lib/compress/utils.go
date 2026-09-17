package compress

import (
	"encoding/binary"
	"fmt"
)

// HumanBytes returns an easy to read representation of bytes
func HumanBytes(n uint64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)
	switch {
	case n >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(n)/float64(GiB))
	case n >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(n)/float64(MiB))
	case n >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(n)/float64(KiB))
	}
	return fmt.Sprintf("%d B", n)
}

// readSizeAt decodes the count/length prefix of a length-prefixed kind
// following the MMDB spec:
//
//	size 0..28: literal value
//	size 29:    next 1 byte is (length - 29)
//	size 30:    next 2 bytes are (length - 285) big-endian
//	size 31:    next 3 bytes are (length - 65821) big-endian
func readSizeAt(buf []byte, sizeBits int, cur uint32) (uint32, uint32, error) {
	switch {
	case sizeBits <= 28:
		return uint32(sizeBits), cur, nil
	case sizeBits == 29:
		if int(cur)+1 > len(buf) {
			return 0, 0, fmt.Errorf("size29 ext")
		}
		return 29 + uint32(buf[cur]), cur + 1, nil
	case sizeBits == 30:
		if int(cur)+2 > len(buf) {
			return 0, 0, fmt.Errorf("size30 ext")
		}
		return 285 + uint32(binary.BigEndian.Uint16(buf[cur:cur+2])), cur + 2, nil
	case sizeBits == 31:
		if int(cur)+3 > len(buf) {
			return 0, 0, fmt.Errorf("size31 ext")
		}
		v := uint32(buf[cur])<<16 | uint32(buf[cur+1])<<8 | uint32(buf[cur+2])
		return 65821 + v, cur + 3, nil
	}
	return 0, 0, fmt.Errorf("invalid sizeBits %d", sizeBits)
}
