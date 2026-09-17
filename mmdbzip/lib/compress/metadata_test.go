package compress

import (
	"strings"
	"testing"
)

// TestMetadataSizeEncodingBoundaries is the regression test for the writeCtrl
// off-by-one at the size-29/size-30 boundary. The size-29 encoding covers
// payload lengths 29..284 (29 + one byte 0..255). The boundary constant was
// `28+255` (=283), which mis-routed a length of exactly 284 into the size-30
// branch where it encoded as uint16(284-285) = 65535 — declaring a 65820-byte
// value that every reader rejects. The fix is the boundary constant `29+255`
// (=284).
//
// We round-trip a map of string values whose lengths straddle every size-class
// boundary and assert exact equality. Length 284 fails to round-trip on the
// unfixed bound (decode reads a 65820-byte string and overruns the buffer) and
// passes after the fix.
func TestMetadataSizeEncodingBoundaries(t *testing.T) {
	lengths := []int{
		28,    // last literal (size 0..28)
		29,    // first size-29
		283,   // size-29 (the old, wrong boundary)
		284,   // size-29 boundary (the bug: must NOT spill into size-30)
		285,   // first size-30
		65820, // last size-30 (285 + 65535)
		65821, // first size-31
	}

	m := make(map[string]any, len(lengths))
	for _, n := range lengths {
		key := "k" + itoa(n)
		m[key] = strings.Repeat("a", n)
	}

	enc, err := encodeMetadata(m)
	if err != nil {
		t.Fatalf("encodeMetadata: %v", err)
	}
	dec, err := decodeMetadata(enc)
	if err != nil {
		t.Fatalf("decodeMetadata: %v", err)
	}

	if len(dec) != len(m) {
		t.Fatalf("decoded map has %d entries, want %d", len(dec), len(m))
	}
	for _, n := range lengths {
		key := "k" + itoa(n)
		got, ok := dec[key].(string)
		if !ok {
			t.Errorf("key %q missing or wrong type (%T)", key, dec[key])
			continue
		}
		if len(got) != n {
			t.Errorf("value for %q: decoded length %d, want %d", key, len(got), n)
			continue
		}
		if got != m[key].(string) {
			t.Errorf("value for %q did not round-trip exactly", key)
		}
	}
}

// TestMetadataMapAndArrayCountBoundary round-trips a map and an array whose
// ENTRY/ELEMENT counts sit on the size-29 boundary (284), since writeCtrl is
// shared by string length and map/array count encoding.
func TestMetadataMapAndArrayCountBoundary(t *testing.T) {
	for _, count := range []int{283, 284, 285} {
		inner := make(map[string]any, count)
		arr := make([]any, count)
		for i := 0; i < count; i++ {
			k := "e" + itoa(i)
			inner[k] = uint16(i % 65536)
			arr[i] = "x" + itoa(i)
		}
		top := map[string]any{
			"the_map":   inner,
			"the_array": arr,
		}
		enc, err := encodeMetadata(top)
		if err != nil {
			t.Fatalf("count=%d encodeMetadata: %v", count, err)
		}
		dec, err := decodeMetadata(enc)
		if err != nil {
			t.Fatalf("count=%d decodeMetadata: %v", count, err)
		}
		gotMap, ok := dec["the_map"].(map[string]any)
		if !ok {
			t.Fatalf("count=%d: the_map missing or wrong type (%T)", count, dec["the_map"])
		}
		if len(gotMap) != count {
			t.Errorf("count=%d: decoded map has %d entries, want %d", count, len(gotMap), count)
		}
		gotArr, ok := dec["the_array"].([]any)
		if !ok {
			t.Fatalf("count=%d: the_array missing or wrong type (%T)", count, dec["the_array"])
		}
		if len(gotArr) != count {
			t.Errorf("count=%d: decoded array has %d elements, want %d", count, len(gotArr), count)
		}
		for i := 0; i < count; i++ {
			if got := gotArr[i].(string); got != "x"+itoa(i) {
				t.Errorf("count=%d: array[%d] = %q, want %q", count, i, got, "x"+itoa(i))
			}
		}
	}
}

// itoa is a tiny base-10 formatter to avoid importing strconv just for tests.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
