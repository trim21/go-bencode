package bencode_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
)

// maxDecodeNestingDepth (internal/decoder/type.go). The interface (any) path
// must honour the same limit the typed slice/map/struct paths already enforce.
const maxDepth = 10000

func nestedLists(n int) []byte {
	return []byte(strings.Repeat("l", n) + strings.Repeat("e", n))
}

func nestedDicts(n int) []byte {
	return []byte(strings.Repeat("d1:a", n) + "i0e" + strings.Repeat("e", n))
}

func TestUnmarshalDepthGuardInterfacePath(t *testing.T) {
	t.Run("list at limit accepted", func(t *testing.T) {
		var out any
		require.NoError(t, bencode.Unmarshal(nestedLists(maxDepth), &out))
	})
	t.Run("list over limit rejected", func(t *testing.T) {
		var out any
		require.Error(t, bencode.Unmarshal(nestedLists(maxDepth+1), &out))
	})
	t.Run("dict at limit accepted", func(t *testing.T) {
		var out any
		require.NoError(t, bencode.Unmarshal(nestedDicts(maxDepth), &out))
	})
	t.Run("dict over limit rejected", func(t *testing.T) {
		var out any
		require.Error(t, bencode.Unmarshal(nestedDicts(maxDepth+1), &out))
	})
	t.Run("typed slice path agrees at the same limit", func(t *testing.T) {
		var ok []any
		require.NoError(t, bencode.Unmarshal(nestedLists(maxDepth), &ok))
		var bad []any
		require.Error(t, bencode.Unmarshal(nestedLists(maxDepth+1), &bad))
	})
	t.Run("unbounded nesting is rejected, not a fatal stack overflow", func(t *testing.T) {
		// Before the guard this recursed into an unrecoverable
		// "fatal error: stack overflow" that recover() cannot catch.
		var out any
		require.Error(t, bencode.Unmarshal([]byte(strings.Repeat("l", 1_000_000)), &out))
	})
}

func TestUnmarshalStringLengthOverflow(t *testing.T) {
	// A maxint64 length prefix made cursor+colon+size overflow int and wrap
	// negative, slipping past the bound check and panicking in the slice
	// expression. readString shares every string/key path, so cover them all.
	const maxIntPrefix = "9223372036854775807:x"

	t.Run("string value", func(t *testing.T) {
		var out any
		require.Error(t, bencode.Unmarshal([]byte(maxIntPrefix), &out))
	})
	t.Run("dictionary key (any)", func(t *testing.T) {
		var out any
		require.Error(t, bencode.Unmarshal([]byte("d9223372036854775807:xe"), &out))
	})
	t.Run("dictionary key (typed map)", func(t *testing.T) {
		var out map[string]string
		require.Error(t, bencode.Unmarshal([]byte("d9223372036854775807:xe"), &out))
	})
	t.Run("typed string field", func(t *testing.T) {
		var out struct {
			A string `bencode:"a"`
		}
		require.Error(t, bencode.Unmarshal([]byte("d1:a9223372036854775807:xe"), &out))
	})
}

// FuzzUnmarshal decodes arbitrary input into any: a malformed blob may error,
// but parsing untrusted bencode must never panic or crash the process.
func FuzzUnmarshal(f *testing.F) {
	seeds := [][]byte{
		[]byte("i42e"),
		[]byte("4:spam"),
		[]byte("l4:spami1ee"),
		[]byte("d3:cow3:moo4:spam4:eggse"),
		[]byte("9223372036854775807:x"),
		[]byte("d9223372036854775807:xe"),
		[]byte(strings.Repeat("l", 100000)),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var out any
		_ = bencode.Unmarshal(data, &out)
	})
}
