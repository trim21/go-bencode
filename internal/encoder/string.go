package encoder

import (
	"reflect"
	"strconv"
)

// encodeString encode string "result" to `s:6:"result";`
// encode UTF-8 string "叛逆的鲁鲁修" `s:18:"叛逆的鲁鲁修";`
// str length is underling bytes length, not len(str)
func encodeString(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	return AppendStr(b, rv.String()), nil
}

func AppendStr(b []byte, s string) []byte {
	b = strconv.AppendInt(b, int64(len(s)), 10)
	b = append(b, ':')
	return append(b, s...)
}
