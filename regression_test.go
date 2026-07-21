package bencode_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	return len(p) - 1, nil
}

func TestMarshalRejectsNil(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "top-level", value: nil},
		{name: "map value", value: map[string]any{"key": nil}},
		{name: "struct field", value: struct{ Value any }{}},
		{name: "slice element", value: []any{nil}},
		{name: "typed nil in interface", value: map[string]any{"key": (*int)(nil)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			require.NotPanics(t, func() {
				_, err = bencode.Marshal(tt.value)
			})
			require.Error(t, err)
		})
	}
}

func TestUnmarshalRejectsNilDestination(t *testing.T) {
	var dst *int
	var err error
	require.NotPanics(t, func() {
		err = bencode.Unmarshal([]byte("i1e"), dst)
	})
	require.Error(t, err)
}

func TestUnmarshalRejectsTruncatedDictionary(t *testing.T) {
	tests := []struct {
		name        string
		destination func() any
	}{
		{name: "typed map", destination: func() any { return &map[string]int{} }},
		{name: "known struct field", destination: func() any {
			return &struct {
				Value int `bencode:"key"`
			}{}
		}},
		{name: "unknown struct field", destination: func() any { return &struct{}{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			require.NotPanics(t, func() {
				err = bencode.Unmarshal([]byte("d3:key"), tt.destination())
			})
			require.Error(t, err)
		})
	}
}

func TestUnmarshalSliceRejectsNonList(t *testing.T) {
	for _, input := range []string{"de", "ie"} {
		t.Run(input, func(t *testing.T) {
			var dst []int
			require.Error(t, bencode.Unmarshal([]byte(input), &dst))
		})
	}
}

func TestEncoderReportsShortWrite(t *testing.T) {
	err := bencode.NewEncoder(shortWriter{}).Encode(42)
	require.ErrorIs(t, err, io.ErrShortWrite)
}
