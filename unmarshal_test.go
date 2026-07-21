package bencode_test

import (
	"fmt"
	"io"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
	"github.com/trim21/go-bencode/internal/decoder"
)

// make sure they are equal
var _ bencode.Unmarshaler = decoder.Unmarshaler(nil)
var _ decoder.Unmarshaler = bencode.Unmarshaler(nil)

func TestUnmarshal(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		var b1 bool
		require.NoError(t, bencode.Unmarshal([]byte("i1e"), &b1))
		require.True(t, b1)
		var b2 bool
		require.NoError(t, bencode.Unmarshal([]byte("i0e"), &b2))
		require.False(t, b2)

		var b bool
		require.Error(t, bencode.Unmarshal([]byte("de"), &b))
		require.Error(t, bencode.Unmarshal([]byte("le"), &b))
		require.Error(t, bencode.Unmarshal([]byte("1:a"), &b))
	})

	t.Run("int", func(t *testing.T) {
		var i int
		require.NoError(t, bencode.Unmarshal([]byte("i100e"), &i))
		require.Equal(t, 100, i)

		require.NoError(t, bencode.Unmarshal([]byte("i-100e"), &i))
		require.Equal(t, -100, i)

		require.Error(t, bencode.Unmarshal([]byte("ie"), &i))
		require.Error(t, bencode.Unmarshal([]byte("i-0e"), &i))
		require.Error(t, bencode.Unmarshal([]byte("i100000000000000000000000000000000000000000e"), &i))
		require.Error(t, bencode.Unmarshal([]byte("1:q"), &i))
	})

	t.Run("uint", func(t *testing.T) {
		var i uint
		require.NoError(t, bencode.Unmarshal([]byte("i100e"), &i))
		require.EqualValues(t, 100, i)

		require.Error(t, bencode.Unmarshal([]byte("i-100e"), &i))

		require.Error(t, bencode.Unmarshal([]byte("ie"), &i))
		require.Error(t, bencode.Unmarshal([]byte("i-0e"), &i))
		require.Error(t, bencode.Unmarshal([]byte("i100000000000000000000000000000000000000000e"), &i))
		require.Error(t, bencode.Unmarshal([]byte("1:q"), &i))
	})

	t.Run("str", func(t *testing.T) {
		var s string
		require.NoError(t, bencode.Unmarshal([]byte("1:e"), &s))
		require.EqualValues(t, "e", s)

		require.Error(t, bencode.Unmarshal([]byte("1:"), &s))

		require.Error(t, bencode.Unmarshal([]byte("1:aq"), &s))

		require.Error(t, bencode.Unmarshal([]byte("ie"), &s))
		require.Error(t, bencode.Unmarshal([]byte("i-0e"), &s))
		require.Error(t, bencode.Unmarshal([]byte("i100000000000000000000000000000000000000000e"), &s))
		require.Error(t, bencode.Unmarshal([]byte("de"), &s))
		require.Error(t, bencode.Unmarshal([]byte("le"), &s))
	})

	t.Run("[]byte", func(t *testing.T) {
		var s []byte
		require.NoError(t, bencode.Unmarshal([]byte("1:e"), &s))
		require.EqualValues(t, "e", s)

		require.Error(t, bencode.Unmarshal([]byte("1:"), &s))

		require.Error(t, bencode.Unmarshal([]byte("1:aq"), &s))

		require.Error(t, bencode.Unmarshal([]byte("ie"), &s))
		require.Error(t, bencode.Unmarshal([]byte("i-0e"), &s))
		require.Error(t, bencode.Unmarshal([]byte("i100000000000000000000000000000000000000000e"), &s))
		require.Error(t, bencode.Unmarshal([]byte("de"), &s))
		require.Error(t, bencode.Unmarshal([]byte("le"), &s))
	})

	t.Run("any", func(t *testing.T) {
		var s any
		require.NoError(t, bencode.Unmarshal([]byte("1:e"), &s))
		require.Equal(t, "e", s)
		s = nil

		require.Error(t, bencode.Unmarshal([]byte("1:"), &s))
		require.Error(t, bencode.Unmarshal([]byte("1:aq"), &s))
		require.Error(t, bencode.Unmarshal([]byte("ie"), &s))
		require.Error(t, bencode.Unmarshal([]byte("i-0e"), &s))
		//require.Error(t, bencode.Unmarshal([]byte("i100000000000000000000000000000000000000000e"), &s))

		s = nil
		require.NoError(t, bencode.Unmarshal([]byte("i1e"), &s))
		require.Equal(t, int64(1), s)

		s = nil
		require.NoError(t, bencode.Unmarshal([]byte("de"), &s))
		require.Equal(t, map[string]any{}, s)

		s = nil
		require.NoError(t, bencode.Unmarshal([]byte("d1:ai1e1:b1:se"), &s))
		require.Equal(t, map[string]any{"a": int64(1), "b": "s"}, s)

		s = nil
		require.NoError(t, bencode.Unmarshal([]byte("le"), &s))
		require.Equal(t, []any{}, s)

		s = nil
		require.NoError(t, bencode.Unmarshal([]byte("li1ei2ei3ei4ee"), &s))
		require.Equal(t, []any{int64(1), int64(2), int64(3), int64(4)}, s)
	})
}

func TestUnmarshal_struct(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		type Container struct {
			F string `bencode:"f1q"`
			V bool   `bencode:"1a9"`
		}

		var c Container
		raw := `d3:f1q10:0147852369e`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, "0147852369", c.F)
	})

	t.Run("empty", func(t *testing.T) {
		type Container struct {
			F string `bencode:"f"`
		}

		var c Container
		raw := `de`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, "", c.F)
	})
}

func TestUnmarshal_struct_bytes(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		type Container struct {
			F []byte `bencode:"v"`
		}

		var c Container
		raw := `d1:v10:0147852369e`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, []byte("0147852369"), c.F)
	})

	t.Run("empty", func(t *testing.T) {
		type Container struct {
			F []byte `bencode:"f"`
		}

		var c Container
		raw := `de`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Nil(t, c.F)
	})
}

func TestUnmarshal_struct_uint(t *testing.T) {

	t.Run("uint", func(t *testing.T) {
		type Container struct {
			F uint `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi147852369ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, uint(147852369), c.F)
	})

	t.Run("uint8", func(t *testing.T) {
		type Container struct {
			F uint8 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi255ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, uint8(255), c.F)
	})

	t.Run("uint16", func(t *testing.T) {
		type Container struct {
			F uint16 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi574ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, uint16(574), c.F)
	})

	t.Run("uint32", func(t *testing.T) {
		type Container struct {
			F uint32 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi57400ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, uint32(57400), c.F)
	})

	t.Run("uint64", func(t *testing.T) {
		type Container struct {
			F uint64 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi5740000ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, uint64(5740000), c.F)
	})
}

func TestUnmarshal_struct_int(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		type Container struct {
			F int `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi147852369ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, int(147852369), c.F)
	})

	t.Run("int8", func(t *testing.T) {
		type Container struct {
			F int8 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi65ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, int8(65), c.F)
	})

	t.Run("int16", func(t *testing.T) {
		type Container struct {
			F int16 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi574ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, int16(574), c.F)
	})

	t.Run("int32", func(t *testing.T) {
		type Container struct {
			F int32 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi57400ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, int32(57400), c.F)
	})

	t.Run("int64", func(t *testing.T) {
		type Container struct {
			F int64 `bencode:"f1q"`
		}

		var c Container
		raw := `d3:f1qi5740000ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, int64(5740000), c.F)
	})
}

func TestUnmarshal_slice(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		type Container struct {
			Value []string `bencode:"value"`
		}

		var c Container
		raw := `d5:valuelee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Len(t, c.Value, 0)
	})

	t.Run("string", func(t *testing.T) {
		type Container struct {
			Value []string `bencode:"value"`
		}
		var c Container
		raw := `d5:valuel3:one3:two1:qee`
		require.NoError(t, bencode.Unmarshal([]byte(raw), &c))
		require.Equal(t, []string{"one", "two", "q"}, c.Value)
	})

	t.Run("string more length", func(t *testing.T) {
		type Container struct {
			Value []string `bencode:"value"`
		}
		var c Container
		raw := `d5:valuel1:01:11:21:31:41:51:61:71:81:9ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}, c.Value)
	})
}

func TestUnmarshal_array(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		type Container struct {
			Value [5]string `bencode:"value"`
		}

		var c Container
		raw := `de`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, [5]string{}, c.Value)
	})

	t.Run("string less length", func(t *testing.T) {
		type Container struct {
			Value [5]string `bencode:"value"`
		}
		var c Container
		raw := `d5:valuel1:01:11:21:3ee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.ErrorContains(t, err, "failed to decode list into GO array, bencode list length")
	})

	t.Run("string more length", func(t *testing.T) {
		type Container struct {
			Value [5]string `bencode:"value"`
		}
		var c Container
		raw := `d5:valuel3:one3:two1:q1:a2:zxee`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, [5]string{"one", "two", "q", "a", "zx"}, c.Value)
	})
}

func TestUnmarshal_bigInt(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		var v = big.Int{}
		require.NoError(t, bencode.Unmarshal([]byte("i0e"), &v))
		require.EqualValues(t, 0, v.Int64())

		v = big.Int{}
		require.NoError(t, bencode.Unmarshal([]byte("i1e"), &v))
		require.EqualValues(t, 1, v.Int64())
	})

	t.Run("struct", func(t *testing.T) {
		vv := Generic[big.Int]{}
		require.NoError(t, bencode.Unmarshal([]byte("d5:Valuei1ee"), &vv))
		require.EqualValues(t, 1, vv.Value.Int64())

		v := Generic[*big.Int]{}
		require.NoError(t, bencode.Unmarshal([]byte("d5:Valuei1ee"), &v))
		require.EqualValues(t, 1, vv.Value.Int64())
	})
}

func TestUnmarshal_skip_value(t *testing.T) {
	type Container struct {
		Value []string `bencode:"value1"`
	}

	var c Container
	raw := `d6:value0de6:value1l3:one3:two1:q1:a2:zxe6:value2lee`
	err := bencode.Unmarshal([]byte(raw), &c)
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two", "q", "a", "zx"}, c.Value)
}

func TestUnmarshal_unmarshaler(t *testing.T) {
	type Container struct {
		Value bencode.RawBytes `bencode:"value"`
	}

	var c Container
	raw := `d5:valued3:keyl3:one3:two1:q1:a2:zxe1:vd1:ai1e1:bi2eeee`
	err := bencode.Unmarshal([]byte(raw), &c)
	require.NoError(t, err)
	require.Equal(t, `d3:keyl3:one3:two1:q1:a2:zxe1:vd1:ai1e1:bi2eee`, string(c.Value))
}

func TestUnmarshal_map(t *testing.T) {
	t.Run("map[string]string", func(t *testing.T) {
		raw := `d5:valued4:five1:54:four1:43:one1:15:three1:33:two1:2ee`
		var c struct {
			Value map[string]string `bencode:"value"`
		}

		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, map[string]string{
			"one":   "1",
			"two":   "2",
			"three": "3",
			"four":  "4",
			"five":  "5",
		}, c.Value)
	})

	t.Run("map[any]string", func(t *testing.T) {
		raw := `d5:valued4:five1:54:four1:43:one1:15:three1:33:two1:2ee`
		var c struct {
			Value map[string]any `bencode:"value"`
		}

		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"one":   "1",
			"two":   "2",
			"three": "3",
			"four":  "4",
			"five":  "5",
		}, c.Value)
	})

	t.Run("any", func(t *testing.T) {
		raw := `d5:valued4:five1:54:four1:43:one1:15:three1:33:two1:2ee`
		var c struct {
			Value any `bencode:"value"`
		}

		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"one":   "1",
			"two":   "2",
			"three": "3",
			"four":  "4",
			"five":  "5",
		}, c.Value)
	})
}

func TestUnmarshal_ptr_string(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		var c struct {
			F *string `bencode:"f1q"`
		}

		raw := `d3:f1q10:0147852369e`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.NotNil(t, c.F)
		require.Equal(t, "0147852369", *c.F)
	})

	t.Run("empty", func(t *testing.T) {
		var c struct {
			F *string `bencode:"f"`
		}

		raw := `de`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.NoError(t, err)
		require.Nil(t, c.F)
	})

	t.Run("nested", func(t *testing.T) {
		var c struct {
			F **string `bencode:"f"`
		}

		raw := `de`
		err := bencode.Unmarshal([]byte(raw), &c)
		require.Error(t, err)
	})
}

func TestUnmarshal_anonymous_field(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		type N struct {
			A int
			B int
		}

		type M struct {
			N
			C int
		}

		var m M
		require.NoError(t, bencode.Unmarshal([]byte("d1:Ai3e1:Bi2e1:Ci1ee"), &m))
		require.Equal(t, M{N: N{
			A: 3,
			B: 2,
		}, C: 1}, m)
	})

	t.Run("named", func(t *testing.T) {
		type N struct {
			A int
			B int
		}

		type M struct {
			N `bencode:"n"`
			C int
		}

		var m M
		require.NoError(t, bencode.Unmarshal([]byte("d1:Ci1e1:nd1:Ai3e1:Bi2eee"), &m))
		require.Equal(t, m, M{N: N{
			A: 3,
			B: 2,
		}, C: 1})
	})

	t.Run("duplicated-name", func(t *testing.T) {
		type N struct {
			C int
		}

		type M struct {
			N
			C int
		}

		var m M
		err := bencode.Unmarshal([]byte("de"), &m)
		require.Error(t, err)
	})
}

func TestUnmarshal_empty_input(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		var data []int
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
	t.Run("array", func(t *testing.T) {
		var data [5]int
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
	t.Run("map", func(t *testing.T) {
		var data map[uint]int
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
	t.Run("interface", func(t *testing.T) {
		var data any
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
	t.Run("string", func(t *testing.T) {
		var data string
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
	t.Run("int", func(t *testing.T) {
		var data int
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})

	t.Run("uint", func(t *testing.T) {
		var data uint
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})

	t.Run("bool", func(t *testing.T) {
		var data bool
		require.Error(t, bencode.Unmarshal([]byte(""), &data))
	})
}

func TestUnmarshal_null_array_1(t *testing.T) {
	raw := `le`

	type Tag struct {
		Name  *string `bencode:"tag_name"`
		Count int     `bencode:"result"`
	}

	var tags []Tag

	err := bencode.Unmarshal([]byte(raw), &tags)
	require.NoError(t, err)
}

func TestUnmarshal_null_array_2(t *testing.T) {
	raw := `d4:Testde1:ai2e1:bde1:oi1ee`

	var data any

	err := bencode.Unmarshal([]byte(raw), &data)
	require.NoError(t, err)

	require.Equal(t, data, map[string]any{
		"a":    int64(2),
		"o":    int64(1),
		"Test": map[string]any{},
		"b":    map[string]any{},
	})
}

func TestUnmarshal_nestedPtr(t *testing.T) {
	type T *int
	type Data struct {
		Field *T
	}

	var data Data
	require.Error(t, bencode.Unmarshal([]byte("de"), &data))
}

func TestUnmarshal_arrayBytes(t *testing.T) {
	var data [20]byte

	err := bencode.Unmarshal([]byte(`20:aaaaaaaaaaaaaaaaaaaa`), &data)
	require.NoError(t, err)

	require.Equal(t, [20]byte([]byte("aaaaaaaaaaaaaaaaaaaa")), data)

	var m map[[20]byte]int
	require.NoError(t, bencode.Unmarshal([]byte(`d20:aaaaaaaaaaaaaaaaaaaai1ee`), &m))
	require.Equal(t, map[[20]byte]int{[20]byte([]byte("aaaaaaaaaaaaaaaaaaaa")): 1}, m)

	require.Error(t, bencode.Unmarshal([]byte(`d19:aaaaaaaaaaaaaaaaaaai1ee`), &m))

	var v struct {
		S     map[[20]byte]struct{} `bencode:"s"`
		Value map[string]string     `bencode:"value"`
	}

	raw := `d5:valued4:five1:5ee`

	require.NoError(t, bencode.Unmarshal([]byte(raw), &v))
}

func TestUnmarshal_unorderedKey(t *testing.T) {
	var m map[string]string
	require.Error(t, bencode.Unmarshal([]byte(`d1:01:01:11:10:0:e`), &m))

	var s struct{}
	require.Error(t, bencode.Unmarshal([]byte(`d1:01:01:11:10:0:e`), &s))

	var a any
	require.Error(t, bencode.Unmarshal([]byte(`d1:01:01:11:10:0:e`), &a))
}

type userType struct {
	t time.Time
}

func (u userType) MarshalBencode() ([]byte, error) {
	return bencode.Marshal(u.t.Format(time.RFC3339))
}

func (u *userType) UnmarshalBencode(bytes []byte) error {
	var s string
	err := bencode.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	u.t = t

	return nil
}

var _ bencode.Unmarshaler = (*userType)(nil)
var _ bencode.Marshaler = userType{}

func TestUnmarshalRelaxed(t *testing.T) {
	t.Run("unordered dict keys - any", func(t *testing.T) {
		// Keys not in sorted order: "min interval" > "complete" but appears first
		raw := `d8:intervali3636e12:min intervali300e8:completei1e10:incompletei0e5:peerslee`

		// Strict should fail
		var v any
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err, "strict parsing should reject unordered keys")

		// Relaxed should succeed
		var v2 any
		err = bencode.UnmarshalRelaxed([]byte(raw), &v2)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"interval":     int64(3636),
			"min interval": int64(300),
			"complete":     int64(1),
			"incomplete":   int64(0),
			"peers":        []any{},
		}, v2)
	})

	t.Run("unordered dict keys - map", func(t *testing.T) {
		raw := `d1:bi2e1:ai1ee`

		var m map[string]int64
		err := bencode.Unmarshal([]byte(raw), &m)
		require.Error(t, err, "strict parsing should reject unordered keys")

		err = bencode.UnmarshalRelaxed([]byte(raw), &m)
		require.NoError(t, err)
		require.Equal(t, map[string]int64{"a": 1, "b": 2}, m)
	})

	t.Run("unordered dict keys - struct", func(t *testing.T) {
		raw := `d1:bi2e1:ai1ee`

		var v struct {
			A int64 `bencode:"a"`
			B int64 `bencode:"b"`
		}
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err, "strict parsing should reject unordered keys")

		err = bencode.UnmarshalRelaxed([]byte(raw), &v)
		require.NoError(t, err)
		require.Equal(t, int64(1), v.A)
		require.Equal(t, int64(2), v.B)
	})

	t.Run("duplicate dict keys", func(t *testing.T) {
		raw := `d1:ai1e1:ai2ee`

		var m map[string]int64
		err := bencode.Unmarshal([]byte(raw), &m)
		require.Error(t, err, "strict parsing should reject duplicate keys")

		err = bencode.UnmarshalRelaxed([]byte(raw), &m)
		require.NoError(t, err)
		require.Equal(t, map[string]int64{"a": 2}, m, "last value wins")
	})

	t.Run("duplicate dict keys - any", func(t *testing.T) {
		raw := `d1:ai1e1:ai2ee`

		var v any
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err, "strict parsing should reject duplicate keys")

		err = bencode.UnmarshalRelaxed([]byte(raw), &v)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": int64(2)}, v, "last value wins")
	})

	t.Run("duplicate dict keys - struct", func(t *testing.T) {
		raw := `d1:ai1e1:ai2ee`

		var v struct {
			A int64 `bencode:"a"`
		}
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err, "strict parsing should reject duplicate keys")

		err = bencode.UnmarshalRelaxed([]byte(raw), &v)
		require.NoError(t, err)
		require.Equal(t, int64(2), v.A, "last value wins")
	})

	t.Run("trailing data", func(t *testing.T) {
		raw := `i1ei2e`

		var v int64
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err, "strict parsing should reject trailing data")

		var v2 int64
		err = bencode.UnmarshalRelaxed([]byte(raw), &v2)
		require.Error(t, err, "relaxed parsing also rejects trailing data")
	})

	t.Run("skipped dict with unordered keys - strict fails", func(t *testing.T) {
		// struct has field "d" but not "a", skipped dict under "a" has unordered keys
		raw := `d1:ad1:ci1e1:bi2ee1:di3ee`

		var v struct {
			D int64 `bencode:"d"`
		}
		// strict should fail because skipped dict has unordered keys
		err := bencode.Unmarshal([]byte(raw), &v)
		require.Error(t, err)
	})

	t.Run("skipped dict with unordered keys - relaxed", func(t *testing.T) {
		raw := `d1:ad1:ci1e1:bi2ee1:di3ee`

		var v struct {
			D int64 `bencode:"d"`
		}
		err := bencode.UnmarshalRelaxed([]byte(raw), &v)
		require.NoError(t, err)
		require.Equal(t, int64(3), v.D)
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	type S struct {
		Name   string
		Length int
	}

	type Data struct {
		I8   int8
		Int  int
		U8   uint8
		Uint uint
		Raw  bencode.RawBytes

		Marshaler userType
		M         map[string]string

		Slice []S

		Str       string
		ByteSlice []byte
		ByteArray [20]byte
	}

	encoded, err := bencode.Marshal(Data{
		I8:        1,
		Int:       2,
		U8:        3,
		Uint:      4,
		M:         map[string]string{"1": "a"},
		Raw:       bencode.RawBytes("i10e"),
		Marshaler: userType{t: time.Now()},
		Str:       "ss",
		ByteSlice: []byte("hello world"),
		Slice: []S{{
			Name:   "index.html",
			Length: 100,
		}, {
			Name:   "index.js",
			Length: 2000,
		}},
		ByteArray: [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	})

	require.NoError(b, err)

	b.Run("type", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var v Data
			err := bencode.Unmarshal(encoded, &v)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("any", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var v any
			err := bencode.Unmarshal(encoded, &v)
			if err != nil {
				panic(err)
			}
		}
	})
}

func TestUnmarshalRelaxed_empty(t *testing.T) {
	var v any
	err := bencode.UnmarshalRelaxed([]byte(""), &v)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty data")

	var s string
	err = bencode.UnmarshalRelaxed([]byte(""), &s)
	require.Error(t, err)

	err = bencode.UnmarshalRelaxed(nil, &v)
	require.Error(t, err)
}

// --- invalid type decoder (chan, func, float, complex) ---

func TestUnmarshal_invalid_types(t *testing.T) {
	t.Run("chan", func(t *testing.T) {
		type S struct{ C chan int `bencode:"c"` }
		var s S
		err := bencode.Unmarshal([]byte("d1:ci1ee"), &s)
		require.Error(t, err)
	})

	t.Run("func", func(t *testing.T) {
		type S struct{ F func() `bencode:"f"` }
		var s S
		err := bencode.Unmarshal([]byte("d1:fi1ee"), &s)
		require.Error(t, err)
	})

	t.Run("float64", func(t *testing.T) {
		type S struct{ F float64 `bencode:"f"` }
		var s S
		err := bencode.Unmarshal([]byte("d1:fi1ee"), &s)
		require.Error(t, err)
	})

	t.Run("complex128", func(t *testing.T) {
		type S struct{ C complex128 `bencode:"c"` }
		var s S
		err := bencode.Unmarshal([]byte("d1:ci1ee"), &s)
		require.Error(t, err)
	})

	t.Run("non-empty interface", func(t *testing.T) {
		var v io.Reader
		err := bencode.Unmarshal([]byte("de"), &v)
		require.Error(t, err)
	})

	t.Run("map with int key", func(t *testing.T) {
		var m map[int]string
		err := bencode.Unmarshal([]byte("de"), &m)
		require.Error(t, err)
	})
}

// --- unmarshaler returning error (annotateError path) ---

type errUnmarshaler struct{}

func (e *errUnmarshaler) UnmarshalBencode([]byte) error {
	return fmt.Errorf("custom unmarshal error")
}

var _ bencode.Unmarshaler = (*errUnmarshaler)(nil)

func TestUnmarshal_unmarshaler_error(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		var v errUnmarshaler
		err := bencode.Unmarshal([]byte("4:test"), &v)
		require.Error(t, err)
		require.Contains(t, err.Error(), "custom unmarshal error")
	})

	t.Run("struct field", func(t *testing.T) {
		type S struct {
			E errUnmarshaler `bencode:"e"`
		}
		var s S
		err := bencode.Unmarshal([]byte("d1:e4:teste"), &s)
		require.Error(t, err)
	})
}

// --- decodeAny default case (invalid first byte) ---

func TestUnmarshal_decodeAny_invalid(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte("e"), &v)
	require.Error(t, err)

	err = bencode.Unmarshal([]byte("x"), &v)
	require.Error(t, err)
}

// --- array overflow (more items than array length) ---

func TestUnmarshal_array_overflow(t *testing.T) {
	type S struct {
		Arr [2]int `bencode:"arr"`
	}
	var s S
	err := bencode.Unmarshal([]byte("d3:arrli1ei2ei3eee"), &s)
	require.Error(t, err)
	require.Contains(t, err.Error(), "array overflow")
}

// --- slice grow (more than 8 elements) ---

func TestUnmarshal_slice_grow(t *testing.T) {
	var s []int
	err := bencode.Unmarshal([]byte("li0ei1ei2ei3ei4ei5ei6ei7ei8ei9ee"), &s)
	require.NoError(t, err)
	require.Len(t, s, 10)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, s)
}

// --- bool invalid values ---

func TestUnmarshal_bool_invalid(t *testing.T) {
	var b bool
	err := bencode.Unmarshal([]byte("i2e"), &b)
	require.Error(t, err)

	err = bencode.Unmarshal([]byte("i"), &b)
	require.Error(t, err)
}

// --- edge-case integer parsing (leading zero, missing e, etc.) ---

func TestUnmarshal_int_edge_cases(t *testing.T) {
	var i int

	// leading zero
	err := bencode.Unmarshal([]byte("i01e"), &i)
	require.Error(t, err)

	// negative with leading zero
	err = bencode.Unmarshal([]byte("i-01e"), &i)
	require.Error(t, err)

	// non-digit
	err = bencode.Unmarshal([]byte("iabe"), &i)
	require.Error(t, err)

	// missing closing 'e' (we can't test exactly i alone since that also fails, 
	// but "i100" will fail with missing 'e')
	err = bencode.Unmarshal([]byte("i100"), &i)
	require.Error(t, err)

	// empty integer body
	err = bencode.Unmarshal([]byte("ie"), &i)
	require.Error(t, err)
}

// --- readString edge cases ---

func TestUnmarshal_readString_edge_cases(t *testing.T) {
	var s string

	// no colon
	err := bencode.Unmarshal([]byte("5"), &s)
	require.Error(t, err)

	// leading zero in length
	err = bencode.Unmarshal([]byte("05:hello"), &s)
	require.Error(t, err)

	// non-digit length
	err = bencode.Unmarshal([]byte("ab:hello"), &s)
	require.Error(t, err)

	// length overflow
	err = bencode.Unmarshal([]byte("10:x"), &s)
	require.Error(t, err)
}

// --- skip syntax errors ---

func TestUnmarshal_skipValue_errors(t *testing.T) {
	// invalid start byte in skipped value via struct
	type S struct {
		V string `bencode:"v"`
	}
	var s S
	// key 'v' followed by something that's not a valid bencode start 
	err := bencode.Unmarshal([]byte("d1:vxe"), &s)
	require.Error(t, err)
}

// --- encode map with [N]byte keys (decoding already tested in TestUnmarshal_arrayBytes)
// NOTE: encoding map with [N]byte keys panics due to unaddressable byte array in arrayByteKeyCompare.
// This is a known bug in the library. We test via decode path only.
func TestMarshal_map_byte_array_key_decode(t *testing.T) {
	var m map[[2]byte]int
	err := bencode.Unmarshal([]byte("d2:\x00\x01i42ee"), &m)
	require.NoError(t, err)
	require.Equal(t, map[[2]byte]int{{0, 1}: 42}, m)
}

// --- decode byte array error paths ---

func TestUnmarshal_byte_array_error(t *testing.T) {
	t.Run("wrong length", func(t *testing.T) {
		var arr [5]byte
		err := bencode.Unmarshal([]byte("3:ab"), &arr)
		require.Error(t, err)
	})

	t.Run("not a string", func(t *testing.T) {
		var arr [5]byte
		err := bencode.Unmarshal([]byte("i42e"), &arr)
		require.Error(t, err)
	})
}

// --- decode integer into uint overflow ---

func TestUnmarshal_uint_overflow(t *testing.T) {
	t.Run("negative into uint", func(t *testing.T) {
		var u uint
		err := bencode.Unmarshal([]byte("i-1e"), &u)
		require.Error(t, err)
	})

	t.Run("negative into uint8", func(t *testing.T) {
		var u uint8
		err := bencode.Unmarshal([]byte("i-1e"), &u)
		require.Error(t, err)
	})
}

// --- int overflow ---

func TestUnmarshal_int_overflow(t *testing.T) {
	t.Run("int8", func(t *testing.T) {
		var i int8
		err := bencode.Unmarshal([]byte("i99999e"), &i)
		require.Error(t, err)
	})

	t.Run("int16", func(t *testing.T) {
		var i int16
		err := bencode.Unmarshal([]byte("i999999e"), &i)
		require.Error(t, err)
	})
}

// --- syntax error in skipDictionary / skipList ---

func TestUnmarshal_syntax_skipDict_error(t *testing.T) {
	// dict with invalid key start
	var v struct {
		A string `bencode:"a"`
	}
	err := bencode.Unmarshal([]byte("d1:xi1e1:bi2ee"), &v)
	require.Error(t, err)
}

func TestUnmarshal_syntax_truncated(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte("d1:ai1e"), &v)
	require.Error(t, err)

	err = bencode.Unmarshal([]byte("li1e"), &v)
	require.Error(t, err)
}

// --- validateType: non-pointer value ---

func TestUnmarshal_validateType(t *testing.T) {
	t.Run("non-pointer", func(t *testing.T) {
		type S struct{ F int }
		var s S
		err := bencode.Unmarshal([]byte("de"), s)
		require.Error(t, err)
	})
}

// --- slice of unsupported element type ---

func TestUnmarshal_slice_of_unsupported(t *testing.T) {
	t.Run("func", func(t *testing.T) {
		var s []func()
		err := bencode.Unmarshal([]byte("li1ee"), &s)
		require.Error(t, err)
	})

	t.Run("chan", func(t *testing.T) {
		var s []chan int
		err := bencode.Unmarshal([]byte("li1ee"), &s)
		require.Error(t, err)
	})
}

// --- array of unsupported element type ---

func TestUnmarshal_array_of_unsupported(t *testing.T) {
	t.Run("func", func(t *testing.T) {
		var a [3]func()
		err := bencode.Unmarshal([]byte("de"), &a)
		require.Error(t, err)
	})
}

// --- triple pointer (triggers newPtrDecoder error) ---

func TestUnmarshal_triple_ptr(t *testing.T) {
	var p ***int
	err := bencode.Unmarshal([]byte("de"), &p)
	require.Error(t, err)
}

// --- decodeKey with invalid key (readString error in struct key) ---

func TestUnmarshal_struct_invalid_key(t *testing.T) {
	type S struct {
		F string `bencode:"f"`
	}
	var s S
	// key length 100 but only 3 bytes available
	err := bencode.Unmarshal([]byte("d100:fi1ee"), &s)
	require.Error(t, err)
}

// --- ptr_to_ptr (compileStruct nested ptr) already covered by TestUnmarshal_nestedPtr ---

// --- reflectInterfaceValue nil interface ---

func TestMarshal_nil_interface(t *testing.T) {
	type S struct {
		Value any `bencode:"value"`
	}
	s := S{Value: nil}
	b, err := bencode.Marshal(s)
	require.Error(t, err)
	require.Nil(t, b)
}

// --- Marshaler returning empty bytes ---

type emptyMarshaler struct{}

func (e emptyMarshaler) MarshalBencode() ([]byte, error) {
	return []byte{}, nil
}

var _ bencode.Marshaler = emptyMarshaler{}

func TestMarshal_empty_marshaler(t *testing.T) {
	_, err := bencode.Marshal(emptyMarshaler{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bytes")
}

// --- big.Int pointer nil (the nil path in encodeBigIntPtr) ---

func TestMarshal_bigIntPtr_nil(t *testing.T) {
	type S struct {
		V *big.Int `bencode:"v"`
	}
	s := S{V: nil}
	b, err := bencode.Marshal(s)
	require.NoError(t, err)
	require.Equal(t, "de", string(b))
}

// --- struct field decode error wrapping ---

func TestUnmarshal_struct_field_decode_error(t *testing.T) {
	type S struct {
		F int `bencode:"f"`
	}
	var s S
	// string "hello" cannot be decoded into int
	err := bencode.Unmarshal([]byte("d1:f5:helloe"), &s)
	require.Error(t, err)
}

// --- Marshal with unsupported type (hits compile default case) ---

func TestMarshal_unsupported_type(t *testing.T) {
	_, err := bencode.Marshal(float64(1.0))
	require.Error(t, err)

	_, err = bencode.Marshal(complex128(0))
	require.Error(t, err)

	_, err = bencode.Marshal(make(chan int))
	require.Error(t, err)
}

// --- Encode with writer error ---

type errWriter struct{}

func (e errWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

func TestEncoder_Encode_write_error(t *testing.T) {
	enc := bencode.NewEncoder(errWriter{})
	err := enc.Encode(42)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write error")
}

// --- struct with anonymous non-struct field (triggers compile error) ---

func TestMarshal_anonymous_non_struct(t *testing.T) {
	type MyInt int
	type S struct {
		MyInt
	}
	_, err := bencode.Marshal(S{})
	require.Error(t, err)
}

func TestUnmarshal_anonymous_non_struct(t *testing.T) {
	type MyInt int
	type S struct {
		MyInt
	}
	var s S
	err := bencode.Unmarshal([]byte("de"), &s)
	require.Error(t, err)
}

// --- encode map with [N]byte keys ---

func TestMarshal_map_byte_array_key(t *testing.T) {
	m := map[[2]byte]int{
		{0, 1}: 42,
		{2, 3}: 100,
	}
	b, err := bencode.Marshal(m)
	require.NoError(t, err)
	// keys sorted by byte comparison: [0,1] < [2,3]
	require.Equal(t, "d2:\x00\x01i42e2:\x02\x03i100ee", string(b))
}

// --- skipValue default: skipped value starts with invalid character ---

func TestUnmarshal_skipValue_default(t *testing.T) {
	type S struct {
		W int `bencode:"w"`
	}
	var s S
	// key 'x' doesn't match struct, its value starts with 'x' (invalid bencode)
	err := bencode.Unmarshal([]byte("d1:wi1e1:xxe"), &s)
	require.Error(t, err)
}

// --- skipDictionary: key without value in skipped dict ---

func TestUnmarshal_skipDict_key_without_value(t *testing.T) {
	type S struct {
		V string `bencode:"v"`
	}
	var s S
	// 'a' key exists but value ('d') starts a dict with key 'x' having no value
	err := bencode.Unmarshal([]byte("d1:ad1:xe1:vi1ee"), &s)
	require.Error(t, err)
}

// --- readString: size overflow ---

func TestUnmarshal_readString_size_overflow(t *testing.T) {
	var v any
	// dict key claims length 5 but only 2 chars available
	err := bencode.Unmarshal([]byte("d5:abe"), &v)
	require.Error(t, err)
}

// --- array decode: truncated at start ---

func TestUnmarshal_array_truncated(t *testing.T) {
	var a [5]int
	err := bencode.Unmarshal([]byte("l"), &a)
	require.Error(t, err)
}

// --- map decode: data too short ---

func TestUnmarshal_map_truncated(t *testing.T) {
	var m map[string]int
	err := bencode.Unmarshal([]byte("di1e"), &m)
	require.Error(t, err)
}

// --- struct decode: truncated dict ---

func TestUnmarshal_struct_truncated(t *testing.T) {
	type S struct {
		F int `bencode:"f"`
	}
	var s S
	// missing closing 'e'
	err := bencode.Unmarshal([]byte("d1:fi1e"), &s)
	require.Error(t, err)
}

// --- bool decode: invalid first char ---

func TestUnmarshal_bool_not_int(t *testing.T) {
	var b bool
	err := bencode.Unmarshal([]byte("1:a"), &b)
	require.Error(t, err)
}

// --- int decode: missing 'e' ---

func TestUnmarshal_int_no_closing_e(t *testing.T) {
	var i int
	err := bencode.Unmarshal([]byte("i42"), &i)
	require.Error(t, err)
}

// --- bool decode: missing 'e' after digit ---

func TestUnmarshal_bool_truncated(t *testing.T) {
	var b bool
	err := bencode.Unmarshal([]byte("i1x"), &b)
	require.Error(t, err)
}

// --- array decode: non-list input ---

func TestUnmarshal_array_not_list(t *testing.T) {
	var a [5]int
	err := bencode.Unmarshal([]byte("i1e"), &a)
	require.Error(t, err)
}

// --- map decode: non-dict input ---

func TestUnmarshal_map_not_dict(t *testing.T) {
	var m map[string]int
	err := bencode.Unmarshal([]byte("i1e"), &m)
	require.Error(t, err)
}

// --- slice decode: non-list input ---

func TestUnmarshal_slice_not_list(t *testing.T) {
	var s []int
	err := bencode.Unmarshal([]byte("i1e"), &s)
	require.Error(t, err)
}

// --- any: truncated list ---

func TestUnmarshal_any_truncated_list(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte("l"), &v)
	require.Error(t, err)
}

// --- any: truncated dict ---

func TestUnmarshal_any_truncated_dict(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte("d"), &v)
	require.Error(t, err)
}

// --- any: list with depth exceeded ---

func TestUnmarshal_any_list_depth_exceeded(t *testing.T) {
	var v any
	raw := make([]byte, 0, 10002)
	for i := 0; i < 10001; i++ {
		raw = append(raw, 'l')
	}
	raw = append(raw, 'e')
	err := bencode.Unmarshal(raw, &v)
	require.Error(t, err)
}

// --- any: dict with depth exceeded ---

func TestUnmarshal_any_dict_depth_exceeded(t *testing.T) {
	var v any
	raw := make([]byte, 0, 10002)
	for i := 0; i < 10001; i++ {
		raw = append(raw, 'd', '1', ':', 'a')
	}
	raw = append(raw, 'i', '0', 'e')
	raw = append(raw, 'e')
	err := bencode.Unmarshal(raw, &v)
	require.Error(t, err)
}

// --- slice decode: depth exceeded ---

func TestUnmarshal_slice_depth_exceeded(t *testing.T) {
	var s []any
	raw := make([]byte, 0, 10002)
	for i := 0; i < 10001; i++ {
		raw = append(raw, 'l')
	}
	for i := 0; i < 10001; i++ {
		raw = append(raw, 'e')
	}
	err := bencode.Unmarshal(raw, &s)
	require.Error(t, err)
}

// --- int decode: non-digit chars ---

func TestUnmarshal_int_non_digit(t *testing.T) {
	var i int
	err := bencode.Unmarshal([]byte("i1a0e"), &i)
	require.Error(t, err)
}

// --- decodeAny: empty buffer ---

func TestUnmarshal_any_empty(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte(""), &v)
	require.Error(t, err)
}

// --- struct: unknown key with skipped list error ---

func TestUnmarshal_struct_skip_list_error(t *testing.T) {
	type S struct {
		F string `bencode:"f"`
	}
	var s S
	// key 'x' doesn't match, value starts with 'l' but is incomplete
	err := bencode.Unmarshal([]byte("d1:xl1:fi1ee"), &s)
	require.Error(t, err)
}

// --- int decode: first char not 'i' ---

func TestUnmarshal_decodeIntegerBytes_not_i(t *testing.T) {
	var i int
	err := bencode.Unmarshal([]byte("1:a"), &i)
	require.Error(t, err)
}

// --- Encode with unsupported type (hits MarshalCtx error in Encode) ---

func TestEncoder_Encode_unsupported(t *testing.T) {
	var buf strings.Builder
	enc := bencode.NewEncoder(&buf)
	err := enc.Encode(chan int(nil))
	require.Error(t, err)
}

// --- Marshal pointer to unsupported type (hits compilePtr inner compile error) ---

func TestMarshal_ptr_to_unsupported(t *testing.T) {
	v := new(chan int)
	_, err := bencode.Marshal(v)
	require.Error(t, err)
}

// --- skipList depth exceeded (during struct unmarshal, skip deeply nested list) ---

func TestUnmarshal_skipList_depth_exceeded(t *testing.T) {
	type S struct {
		X int `bencode:"x"`
	}
	var s S
	nested := strings.Repeat("l", 10001) + strings.Repeat("e", 10001)
	raw := []byte("d1:xi1e1:y" + nested + "e")
	err := bencode.Unmarshal(raw, &s)
	require.Error(t, err)
}

// --- skipDictionary depth exceeded ---

func TestUnmarshal_skipDict_depth_exceeded(t *testing.T) {
	type S struct {
		X int `bencode:"x"`
	}
	var s S
	nested := strings.Repeat("d1:a", 10001) + "i0e" + strings.Repeat("e", 10001)
	raw := []byte("d1:xi1e1:y" + nested + "e")
	err := bencode.Unmarshal(raw, &s)
	require.Error(t, err)
}

// --- skipDictionary: key without value in skipped dict ---

func TestUnmarshal_skipDict_key_no_value(t *testing.T) {
	type S struct {
		X int `bencode:"x"`
	}
	var s S
	// key 'y' doesn't match, value is dict "d1:z" with key but no value
	err := bencode.Unmarshal([]byte("d1:xi1e1:yd1:zee"), &s)
	require.Error(t, err)
}

// --- Marshal **int (hits compilePtr nested ptr check) ---

func TestMarshal_nested_ptr_standalone(t *testing.T) {
	p := new(*int)
	_, err := bencode.Marshal(p)
	require.Error(t, err)
}

// --- skipList: truncated list during struct skip ---

func TestUnmarshal_skipList_truncated(t *testing.T) {
	type S struct {
		X int `bencode:"x"`
	}
	var s S
	err := bencode.Unmarshal([]byte("d1:xi1e1:yle"), &s)
	require.Error(t, err)
}

// --- readString: leading zero in multi-digit length ---

func TestUnmarshal_readString_leading_zero(t *testing.T) {
	var v any
	err := bencode.Unmarshal([]byte("d0010:helloworlde"), &v)
	require.Error(t, err)
}

// --- int decode: single non-digit char between i and e ---

func TestUnmarshal_int_single_non_digit(t *testing.T) {
	var i int
	err := bencode.Unmarshal([]byte("ixe"), &i)
	require.Error(t, err)
}

// --- int decode: non-digit in multi-char integer ---

func TestUnmarshal_int_validIntBytes_fail(t *testing.T) {
	var i int
	err := bencode.Unmarshal([]byte("i1x2e"), &i)
	require.Error(t, err)
}

// --- big.Int decode: invalid integer ---

func TestUnmarshal_bigInt_invalid(t *testing.T) {
	type S struct {
		V *big.Int `bencode:"v"`
	}
	var s S
	err := bencode.Unmarshal([]byte("d1:v1:xe"), &s)
	require.Error(t, err)
}
