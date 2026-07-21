package decoder

import (
	stderrors "errors"
	"reflect"
	"testing"

	berrors "github.com/trim21/go-bencode/internal/errors"
)

type typeErrorValue struct{}

func (*typeErrorValue) UnmarshalBencode([]byte) error {
	return &berrors.UnmarshalTypeError{Value: "string", Type: reflect.TypeFor[int](), Offset: 99}
}

type syntaxErrorValue struct{}

func (*syntaxErrorValue) UnmarshalBencode([]byte) error {
	return berrors.ErrSyntax("custom syntax error", 99)
}

type nonEmptyInterface interface {
	Method()
}

type recursiveValue struct {
	Next *recursiveValue
}

func TestUnmarshalerErrorContext(t *testing.T) {
	t.Run("type error gets field context", func(t *testing.T) {
		var dst struct {
			Value typeErrorValue `bencode:"value"`
		}
		err := Unmarshal([]byte("d5:value1:xe"), &dst)

		var typeErr *berrors.UnmarshalTypeError
		if !stderrors.As(err, &typeErr) {
			t.Fatalf("Unmarshal() error = %v, want UnmarshalTypeError", err)
		}
		if typeErr.Struct != "" || typeErr.Field != "value" {
			t.Fatalf("error context = %q.%q, want %q.%q", typeErr.Struct, typeErr.Field, "", "value")
		}
	})

	t.Run("syntax error gets source offset", func(t *testing.T) {
		var dst syntaxErrorValue
		err := Unmarshal([]byte("1:x"), &dst)

		var syntaxErr *berrors.SyntaxError
		if !stderrors.As(err, &syntaxErr) {
			t.Fatalf("Unmarshal() error = %v, want SyntaxError", err)
		}
		if syntaxErr.Offset != 0 {
			t.Fatalf("error offset = %d, want 0", syntaxErr.Offset)
		}
	})

	t.Run("malformed source is rejected before callback", func(t *testing.T) {
		var dst syntaxErrorValue
		if err := Unmarshal([]byte("x"), &dst); err == nil {
			t.Fatal("Unmarshal() unexpectedly succeeded")
		}
	})
}

func TestDecoderCompilationErrors(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
	}{
		{name: "map key", typ: reflect.TypeFor[map[int]string]()},
		{name: "map value", typ: reflect.TypeFor[map[string]nonEmptyInterface]()},
		{name: "slice element", typ: reflect.TypeFor[[]nonEmptyInterface]()},
		{name: "array element", typ: reflect.TypeFor[[1]nonEmptyInterface]()},
		{name: "interface", typ: reflect.TypeFor[nonEmptyInterface]()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := compile(tt.typ, "", "", map[reflect.Type]Decoder{}); err == nil {
				t.Fatalf("compile(%s) unexpectedly succeeded", tt.typ)
			}
		})
	}
}

func TestStructDecoderBoundaries(t *testing.T) {
	if _, err := compile(reflect.TypeFor[recursiveValue](), "", "", map[reflect.Type]Decoder{}); err != nil {
		t.Fatalf("compile recursive struct: %v", err)
	}

	type ignoredValue struct {
		Ignored int `bencode:"-"`
	}
	if _, err := compile(reflect.TypeFor[ignoredValue](), "", "", map[reflect.Type]Decoder{}); err != nil {
		t.Fatalf("compile ignored field: %v", err)
	}

	type invalidValue struct {
		Value nonEmptyInterface
	}
	if _, err := compile(reflect.TypeFor[invalidValue](), "", "", map[reflect.Type]Decoder{}); err == nil {
		t.Fatal("struct with unsupported field unexpectedly compiled")
	}

	type emptyValue struct{}
	dec, err := compile(reflect.TypeFor[emptyValue](), "", "", map[reflect.Type]Decoder{})
	if err != nil {
		t.Fatalf("compile empty struct: %v", err)
	}
	structDec := dec.(*structDecoder)
	rv := reflect.ValueOf(&emptyValue{}).Elem()

	tests := []struct {
		name  string
		ctx   *Context
		depth int64
	}{
		{name: "short input", ctx: &Context{Buf: []byte("d")}},
		{name: "wrong type", ctx: &Context{Buf: []byte("le")}},
		{name: "max depth", ctx: &Context{Buf: []byte("de")}, depth: maxDecodeNestingDepth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := structDec.Decode(tt.ctx, 0, tt.depth, rv); err == nil {
				t.Fatal("Decode() unexpectedly succeeded")
			}
		})
	}
}

func TestContainerDecoderBoundaries(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		dec, err := compile(reflect.TypeFor[map[string]int](), "", "", map[reflect.Type]Decoder{})
		if err != nil {
			t.Fatalf("compile map: %v", err)
		}
		mapDec := dec.(*mapDecoder)
		var dst map[string]int
		rv := reflect.ValueOf(&dst).Elem()

		for _, tc := range []struct {
			name  string
			buf   string
			depth int64
		}{
			{name: "empty", buf: ""},
			{name: "short", buf: "d"},
			{name: "max depth", buf: "de", depth: maxDecodeNestingDepth},
			{name: "missing terminator", buf: "d1:ai1e"},
			{name: "invalid value", buf: "d1:a1:xe"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := mapDec.Decode(&Context{Buf: []byte(tc.buf)}, 0, tc.depth, rv); err == nil {
					t.Fatal("Decode() unexpectedly succeeded")
				}
			})
		}
	})

	t.Run("slice", func(t *testing.T) {
		dec, err := compile(reflect.TypeFor[[]int](), "", "", map[reflect.Type]Decoder{})
		if err != nil {
			t.Fatalf("compile slice: %v", err)
		}
		sliceDec := dec.(*sliceDecoder)
		var dst []int
		rv := reflect.ValueOf(&dst).Elem()

		for _, tc := range []struct {
			name  string
			buf   string
			depth int64
		}{
			{name: "empty", buf: ""},
			{name: "max depth", buf: "le", depth: maxDecodeNestingDepth},
			{name: "missing terminator", buf: "l"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := sliceDec.Decode(&Context{Buf: []byte(tc.buf)}, 0, tc.depth, rv); err == nil {
					t.Fatal("Decode() unexpectedly succeeded")
				}
			})
		}
	})
}

func TestSyntaxHelpersRejectMalformedValues(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{name: "list ends early", call: func() error { _, err := skipList([]byte("l"), 0, 0, false); return err }},
		{name: "dictionary ends early", call: func() error { _, err := skipDictionary([]byte("d"), 0, 0, false); return err }},
		{name: "not a dictionary", call: func() error { _, err := skipDictionary([]byte("le"), 0, 0, false); return err }},
		{name: "missing string length", call: func() error { _, _, err := readString([]byte(":x"), 0); return err }},
		{name: "missing colon", call: func() error { _, _, err := readString([]byte("1x"), 0); return err }},
		{name: "invalid string length", call: func() error { _, _, err := readString([]byte("x:y"), 0); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("malformed value unexpectedly succeeded")
			}
		})
	}
}

func TestDirectDecoderBounds(t *testing.T) {
	t.Run("interface cursor", func(t *testing.T) {
		dec := newInterfaceDecoder(reflect.TypeFor[any](), "", "")
		ctx := &Context{Buf: nil}
		var dst any
		if _, err := dec.Decode(ctx, 0, 0, reflect.ValueOf(&dst).Elem()); err == nil {
			t.Fatal("Decode() unexpectedly succeeded")
		}
		if _, _, err := dec.decodeAny(ctx, 0, 0); err == nil {
			t.Fatal("decodeAny() unexpectedly succeeded")
		}
	})

	t.Run("array cursor and size", func(t *testing.T) {
		dec := newArrayDecoder(newIntDecoder(reflect.TypeFor[int](), "", ""), reflect.TypeFor[int](), 1, "", "")
		var dst [1]int
		rv := reflect.ValueOf(&dst).Elem()

		if _, err := dec.Decode(&Context{}, 0, 0, rv); err == nil {
			t.Fatal("empty input unexpectedly succeeded")
		}
		if _, err := dec.Decode(&Context{Buf: []byte("de")}, 0, 0, rv); err == nil {
			t.Fatal("non-list input unexpectedly succeeded")
		}
		if _, err := dec.Decode(&Context{Buf: []byte("le")}, 0, 0, rv); err == nil {
			t.Fatal("short array input unexpectedly succeeded")
		}
		if _, err := dec.Decode(&Context{Buf: []byte("li1ei2ee")}, 0, 0, rv); err == nil {
			t.Fatal("oversized array input unexpectedly succeeded")
		}
	})
}
