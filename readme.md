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

If you want to encode customize type as struct field with `omitempty`,
do implement both `bencode.Marshaler` and `bencode.IsZeroValue`,
so encoder could know if it's a empty value and skip fields.

Bencode doesn't have null type, so all struct field with pointer type(`*T`) will get `omitempty` by default.

### Unmarshal

go `any` type will be decoded as `map[string]any`, `[]any`, `int64` or `string`.

`[]uint8`(`[]byte`) and `[N]uint8`(`[N]byte`) will be decoded as bencode string.

Decode Go string may not be valid utf8 string.

Go Array will be decoded with size check.
Only bencode string/list with same length are valid, otherwise it will return a error.

## Note

go `reflect` package allow you to create dynamic struct
with [reflect.StructOf](https://pkg.go.dev/reflect#StructOf),
but please use it with caution.

For performance, this package will try to "compile" input type to a static encoder/decoder
at first time and cache it for future use.

So a dynamic struct may cause memory leak.

## License

MIT License
