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

// MarshalTo appends the bencode encoding of v to dst and returns the extended
// buffer. It doesn't copy the result, so the returned slice may share the array
// of dst, and dst must not be written to while the result is still in use.
//
// MarshalTo is for a caller which keeps a buffer of its own and encodes many
// values into it. A nil dst has no buffer to append to, and is the same as
// Marshal.
func MarshalTo(dst []byte, v any) ([]byte, error) {
	if dst == nil {
		return Marshal(v)
	}

	ctx := encoder.NewCtx()

	// The context goes back to the pool with the buffer it came with: neither dst
	// nor the array the encoder grows into belongs to the pool, and handing one of
	// them to the next call would overwrite the result.
	pooled := ctx.Buf
	ctx.Buf = dst

	err := encoder.MarshalCtx(ctx, v)
	out := ctx.Buf

	ctx.Buf = pooled
	encoder.FreeCtx(ctx)

	if err != nil {
		return dst, err
	}

	return out, nil
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

	n, err := e.w.Write(ctx.Buf)
	if err == nil && n != len(ctx.Buf) {
		return io.ErrShortWrite
	}
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
