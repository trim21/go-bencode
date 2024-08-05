package bencode

import (
	"io"

	"github.com/trim21/go-bencode/internal/encoder"
)

// Marshaler allow users to implement its own encoder.
type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

// IsZeroValue add support type implements Marshaler and omitempty
//
//	var s struct {
//		Field T `bencode:"field,omitempty"`
//	}
//
// if `T` implements  [Marshaler], it can implement [IsZeroValue] so bencode know if it's
type IsZeroValue interface {
	// IsZeroBencodeValue enable support for omitempty feature.
	// if it's being used as struct field, and it returns true, this field will be skipped.
	IsZeroBencodeValue() bool
}

func Marshal(v any) ([]byte, error) {
	ctx := encoder.NewCtx()
	defer encoder.FreeCtx(ctx)

	err := encoder.MarshalCtx(ctx, v)
	if err != nil {
		return nil, err
	}

	return append([]byte(nil), ctx.Buf...), nil
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v any) error {
	ctx := encoder.NewCtx()
	defer encoder.FreeCtx(ctx)

	err := encoder.MarshalCtx(ctx, v)
	if err != nil {
		return err
	}

	_, err = e.w.Write(ctx.Buf)
	return err
}

func AppendInt(b []byte, i int64) []byte {
	return encoder.AppendInt(b, i)
}

func AppendStr(b []byte, s string) []byte {
	return encoder.AppendStr(b, s)
}

func AppendBytes(b []byte, s []byte) []byte {
	return encoder.AppendBytes(b, s)
}
