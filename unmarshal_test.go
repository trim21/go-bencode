package bencode_test

import (
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

	for i := 0; i < b.N; i++ {
		var v Data
		err := bencode.Unmarshal(encoded, &v)
		if err != nil {
			panic(err)
		}
	}
}
