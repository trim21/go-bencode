package decoder

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/trim21/go-bencode/internal/errors"
)

func skipString(buf []byte, cursor int) (int, error) {
	_, end, err := readString(buf, cursor)
	return end, err
}

func skipInteger(buf []byte, cursor int) (int, error) {
	_, end, err := decodeIntegerBytes(buf, cursor)
	return end, err
}

func skipList(buf []byte, cursor int, depth int64) (int, error) {
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	cursor++

	bufSize := len(buf)

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			return cursor + 1, nil
		}

		c, err := skipValue(buf, cursor, depth)
		if err != nil {
			return 0, err
		}

		cursor = c
	}
}

func skipDictionary(buf []byte, cursor int, depth int64) (int, error) {
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	bufSize := len(buf)

	if cursor+2 > bufSize {
		return 0, errors.ErrSyntax("buffer overflow when parsing directory", cursor)
	}

	if buf[cursor] == 'd' {
		cursor++
	} else {
		return 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
	}

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		_, c, err := readString(buf, cursor)
		if err != nil {
			return 0, err
		}

		cursor = c

		if cursor >= bufSize {
			return 0, errors.ErrExpected("object value after colon", cursor)
		}

		c, err = skipValue(buf, cursor, depth)
		if err != nil {
			return 0, err
		}
		cursor = c
	}
}

// skip value with index also check syntax
func skipValue(buf []byte, cursor int, depth int64) (int, error) {
	switch buf[cursor] {
	case 'l':
		return skipList(buf, cursor, depth)
	case 'd':
		return skipDictionary(buf, cursor, depth)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return skipString(buf, cursor)
	case 'i':
		return skipInteger(buf, cursor)
	default:
		return cursor, errors.ErrUnexpectedEnd("null", cursor)
	}

}

// parse `${length}:${content}` and return "${content}" as slice of buf.
// return cursor set to start of next value
func readString(buf []byte, cursor int) ([]byte, int, error) {
	colon := bytes.IndexByte(buf[cursor:], ':')

	if colon == -1 {
		return nil, 0, fmt.Errorf("invalid bytes, failed find expected char ':'. index %d", cursor)
	}

	if colon == 0 {
		return nil, 0, fmt.Errorf("invalid bytes, missing leading length. index %d", cursor)
	}

	sizeBuf := buf[cursor : cursor+colon]

	if !validIntBytes(sizeBuf) {
		return nil, 0, fmt.Errorf("invalid bytes, length is not valid int. index %d", cursor)
	}

	if colon > 1 {
		if sizeBuf[0] == '0' {
			return nil, 0, fmt.Errorf("invalid bytes, leading 0 in length. index %d", cursor)
		}
	}

	size, err := strconv.Atoi(string(sizeBuf))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid bytes, length is not valid int. index %d", cursor)
	}

	if len(buf) <= cursor+colon+size {
		return nil, 0, errors.ErrSyntax("invalid bytes, size overflow buffer. index %d", cursor)
	}

	end := cursor + colon + size + 1

	return buf[cursor+colon+1 : end], end, nil
}

func parseUint64(b []byte) (uint64, error) {
	// fast path, input should be already validated.
	if len(b) < 20 {
		var r uint64
		for _, c := range b {
			r = r*10 + uint64(c-'0')
		}

		return r, nil
	}

	return strconv.ParseUint(string(b), 10, 64)
}
