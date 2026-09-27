package decoder

import (
	"bytes"
	"fmt"
	"math"

	"github.com/trim21/go-bencode/internal/errors"
)

func skipString(buf []byte, cursor int) (int, error) {
	_, end, err := readString(buf, cursor)
	return end, err
}

func skipInteger(buf []byte, cursor int) (int, error) {
	_, _, _, _, end, err := decodeIntegerBytes(buf, cursor)
	return end, err
}

func skipList(buf []byte, cursor int, depth int64, relaxed bool) (int, error) {
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	cursor++

	bufSize := len(buf)

	for {
		if cursor >= bufSize {
			return 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			return cursor + 1, nil
		}

		c, err := skipValue(buf, cursor, depth, relaxed)
		if err != nil {
			return 0, err
		}

		cursor = c
	}
}

func skipDictionary(buf []byte, cursor int, depth int64, relaxed bool) (int, error) {
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	bufSize := len(buf)

	if cursor+2 > bufSize {
		return 0, errors.DataTooShort()
	}

	if buf[cursor] != 'd' {
		return 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
	}
	cursor++

	var lastKey []byte

	for {
		if cursor >= bufSize {
			return 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		currentKey, c, err := readString(buf, cursor)
		if err != nil {
			return 0, err
		}

		if !relaxed && lastKey != nil {
			switch bytes.Compare(lastKey, currentKey) {
			case 0:
				return cursor, fmt.Errorf("dictionary conrains duplicated keys %s. index %d", currentKey, cursor)
			case 1:
				return cursor, fmt.Errorf("dictionary conrains unordered keys %s, %s. index %d", lastKey, currentKey, cursor)
			}
		}
		lastKey = currentKey

		cursor = c

		if cursor >= bufSize {
			return 0, errors.ErrExpecting("object value after colon", buf, cursor)
		}

		c, err = skipValue(buf, cursor, depth, relaxed)
		if err != nil {
			return 0, err
		}
		cursor = c
	}
}

// skip value with index also check syntax
func skipValue(buf []byte, cursor int, depth int64, relaxed bool) (int, error) {
	switch buf[cursor] {
	case 'l':
		return skipList(buf, cursor, depth, relaxed)
	case 'd':
		return skipDictionary(buf, cursor, depth, relaxed)
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

	size, leadingZero, ok := parseLength(buf[cursor : cursor+colon])
	if !ok {
		return nil, 0, fmt.Errorf("invalid bytes, length is not valid int. index %d", cursor)
	}

	if leadingZero {
		return nil, 0, fmt.Errorf("invalid bytes, leading 0 in length. index %d", cursor)
	}

	// size is attacker-controlled up to maxint; subtract instead of adding so
	// cursor+colon+size can't overflow int and wrap negative past the bound check.
	if size > len(buf)-(cursor+colon+1) {
		return nil, 0, errors.ErrSyntax("invalid bytes, size overflow buffer. index %d", cursor)
	}

	end := cursor + colon + size + 1

	return buf[cursor+colon+1 : end], end, nil
}

// parseLength parses the digits of a bencode string length prefix in one pass: it
// validates them and accumulates them, and it reports a leading zero of its own
// because bencode gives it its own error. ok is false for a prefix which is not a
// number, or which wouldn't fit in an int.
func parseLength(b []byte) (size int, leadingZero bool, ok bool) {
	if b[0] < '0' || b[0] > '9' {
		return 0, false, false
	}

	size = int(b[0] - '0')
	leadingZero = b[0] == '0' && len(b) > 1

	for i := 1; i < len(b); i++ {
		c := b[i]
		if c < '0' || c > '9' {
			return 0, false, false
		}

		d := int(c - '0')
		if size > (math.MaxInt-d)/10 {
			return 0, false, false
		}

		size = size*10 + d
	}

	return size, leadingZero, true
}
