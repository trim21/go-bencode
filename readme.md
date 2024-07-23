# go-bencode

![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/trim21/go-bencode?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/trim21/go-bencode#section-readme.svg)](https://pkg.go.dev/github.com/trim21/go-bencode#section-readme)

Decoding and encoding bencode.

Support All go type including `map`, `slice`, `struct`, `array`, and simple type like `int`, `uint`
...etc.

Encoding and decoding some type from standard library like `time.Time`, `net.IP` are not supported.
If you have any thought about how to support these types, please create an issue.

Or you can wrap these types and implement `bencode.Marshaler` or `bencode.Unmarshaler`

## Supported and tested go version

- 1.22

## Install

```console
go get github.com/trim21/go-bencode
```

## Usage

See [examples](./example_test.go)

### Unmarshal

go `any` type will be decoded as `map[string]any`, `[]any`, `int64` or `string`.

`[]uint8` and `[...]uint8` will be decoded as bencode string.

Decode Go string may not be valid utf8 string.

GO Array will be decoded as bencode list with size check.
Only bencode list with same length are valid, otherwise it will return a error.

## Note

go `reflect` package allow you to create dynamic struct
with [reflect.StructOf](https://pkg.go.dev/reflect#StructOf),
but please use it with caution.

For performance, this package will try to "compile" input type to a static encoder/decoder
at first time and cache it for future use.

So a dynamic struct may cause memory leak.

## License

MIT License
