# go-bencode

![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/trim21/go-bencode?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/trim21/go-bencode#section-readme.svg)](https://pkg.go.dev/github.com/trim21/go-bencode#section-readme)

Decoding and encoding bencode.

Support All go type including `map`/`slice`/`struct`/`array`, and simple type like `bool`/`int`/`uint`/`string`/....

`float32` and `float64` are not supported, bencode doesn't have this type.

Encoding and decoding some type from standard library like `time.Time`, `net.IP` are not supported.
If you have any thought about how to support these types, please create an issue.

Or you can wrap these types and implement `bencode.Marshaler` or `bencode.Unmarshaler`

## Install

```console
go get github.com/trim21/go-bencode
```

## Usage

See [examples](./example_test.go)

### Marshal

Bencode doesn't have null type, so all pointer fields (`*T`) in structs get `omitempty` by default — a nil pointer is skipped during encoding.

If you want to apply `omitempty` to custom types, implement both `bencode.Marshaler` and `bencode.IsZeroValue`, so the encoder can determine whether a value is empty.

See [bencode.RawBytes](https://pkg.go.dev/github.com/trim21/go-bencode#RawBytes) for an example.

### Unmarshal

#### Basic Types

```go
var b bool
bencode.Unmarshal([]byte("i1e"), &b) // true
bencode.Unmarshal([]byte("i0e"), &b) // false

var i int
bencode.Unmarshal([]byte("i100e"), &i)  // 100
bencode.Unmarshal([]byte("i-100e"), &i) // -100

var u uint
bencode.Unmarshal([]byte("i100e"), &u)  // 100

var s string
bencode.Unmarshal([]byte("5:hello"), &s) // "hello"
bencode.Unmarshal([]byte("0:"), &s)      // ""
```

Supported integer types: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `*big.Int`.

#### Raw Bytes

`[]byte` and `[N]byte` are decoded as bencode string (raw bytes, not list of integers):

```go
var b []byte
bencode.Unmarshal([]byte("5:hello"), &b) // []byte("hello")

var arr [20]byte
bencode.Unmarshal([]byte("20:aaaaaaaaaaaaaaaaaaaa"), &arr)
```

Go arrays have a strict length check: the bencode string/list must have exactly the same length, otherwise an error is returned.

#### `any` Type

When decoding into `any`, the target type is inferred:

```go
var v any
bencode.Unmarshal([]byte("i1e"), &v)                          // int64(1)
bencode.Unmarshal([]byte("5:hello"), &v)                      // "hello"
bencode.Unmarshal([]byte("le"), &v)                           // []any{}
bencode.Unmarshal([]byte("li1ei2ei3ee"), &v)                  // []any{int64(1), int64(2), int64(3)}
bencode.Unmarshal([]byte("de"), &v)                           // map[string]any{}
bencode.Unmarshal([]byte("d1:ai1e1:bssee"), &v)               // map[string]any{"a": int64(1), "b": "ss"}
```

#### Structs

Struct fields are mapped by the `bencode` tag. Use `-` to skip a field, or `bencode:",omitempty"` to omit zero values.

```go
type Container struct {
    F string `bencode:"f1q"`
    V int64  `bencode:"1a9"`
    Skip bool `bencode:"-"`
}

var c Container
bencode.Unmarshal([]byte(`d3:f1q10:0147852369e`), &c)
// c.F == "0147852369"
```

Missing fields are left at their zero values; extra unknown keys are skipped silently.

Pointer fields are set to `nil` when the key is absent, and allocated when present:

```go
type Container struct {
    F *string `bencode:"f1q"`
}
var c Container
bencode.Unmarshal([]byte(`d3:f1q10:0147852369e`), &c) // c.F != nil, *c.F == "0147852369"
bencode.Unmarshal([]byte(`de`), &c)                    // c.F == nil
```

Nested pointer (`**T`) is not supported.

Anonymous (embedded) struct fields are flattened into the parent. If an anonymous field has its own `bencode` tag, it is treated as a named sub-dict instead.

#### Slices & Maps

```go
// slices
type S struct {
    Value []string `bencode:"value"`
}
var c S
bencode.Unmarshal([]byte(`d5:valuel3:one3:two1:qee`), &c)
// c.Value == []string{"one", "two", "q"}

// maps (keys must be string or [N]byte)
type M struct {
    Value map[string]string `bencode:"value"`
}
var c M
bencode.Unmarshal([]byte(`d5:valued4:five1:54:four1:4ee`), &c)
// c.Value == map[string]string{"five": "5", "four": "4"}
```

Map keys can also be `[N]byte` (fixed-size byte array).

#### Custom Unmarshaler

Implement `bencode.Unmarshaler` for custom decoding logic:

```go
type Unmarshaler interface {
    UnmarshalBencode([]byte) error
}
```

`bencode.RawBytes` is a built-in type that implements this interface, capturing raw bencode bytes:

```go
var c struct {
    Value bencode.RawBytes `bencode:"value"`
}
bencode.Unmarshal([]byte(`d5:valued3:keyli1ei2eee`), &c)
// string(c.Value) == `d3:keyli1ei2eee`
```

#### Strict vs Relaxed Parsing

By default, `Unmarshal` enforces strict bencode rules:
- Dictionary keys must be in lexicographic order
- Duplicate dictionary keys are rejected

Use `UnmarshalRelaxed` to allow out-of-order keys and duplicate keys (last value wins):

```go
// strict: fails because keys are out of order
bencode.Unmarshal([]byte(`d1:bi2e1:ai1ee`), &m)

// relaxed: succeeds
bencode.UnmarshalRelaxed([]byte(`d1:bi2e1:ai1ee`), &m)
// m == map[string]int64{"a": 1, "b": 2}

// strict: fails on duplicate keys
bencode.Unmarshal([]byte(`d1:ai1e1:ai2ee`), &m)

// relaxed: last value wins
bencode.UnmarshalRelaxed([]byte(`d1:ai1e1:ai2ee`), &m)
// m == map[string]int64{"a": 2}
```

#### Notes

- Input must not be empty (`""`), regardless of target type.
- Decoded Go string may not be valid UTF-8; validate yourself if needed.
- Go arrays (not slices) are decoded with a length check — the bencode list/string must have exactly the same number of elements.

## Note

go `reflect` package allow you to create dynamic struct
with [reflect.StructOf](https://pkg.go.dev/reflect#StructOf),
but please use it with caution.

For performance, this package will try to "compile" input type to a static encoder/decoder
at first time and cache it for future use.

So a dynamic struct may cause memory leak.

## License

MIT License
