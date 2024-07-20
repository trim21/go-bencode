package bencode_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
	"github.com/trim21/go-bencode/internal/test"
)

func TestMarshalBytes(t *testing.T) {
	var b = bencode.RawMessage("i1e")

	var S = struct {
		V bencode.RawMessage `bencode:"v"`
	}{V: b}

	actual, err := bencode.Marshal(S)
	require.NoError(t, err)
	test.StringEqual(t, "d1:vi1ee", string(actual))
}
