package bencode_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
	"github.com/trim21/go-bencode/internal/test"
)

func TestMarshalBytes(t *testing.T) {
	var b = bencode.RawBytes("i1e")

	var S = struct {
		V bencode.RawBytes `bencode:"v"`
	}{V: b}

	actual, err := bencode.Marshal(S)
	require.NoError(t, err)
	test.StringEqual(t, "d1:vi1ee", actual)
}

func TestRawBytes_UnmarshalBencode_nil(t *testing.T) {
	var b *bencode.RawBytes
	err := b.UnmarshalBencode([]byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil pointer")
}

func TestRawBytes_UnmarshalBencode(t *testing.T) {
	var b bencode.RawBytes
	err := b.UnmarshalBencode([]byte("i42e"))
	require.NoError(t, err)
	require.Equal(t, bencode.RawBytes("i42e"), b)
}

func TestRawBytes_IsZeroBencodeValue(t *testing.T) {
	var b bencode.RawBytes
	require.True(t, b.IsZeroBencodeValue())

	b = bencode.RawBytes("x")
	require.False(t, b.IsZeroBencodeValue())
}
