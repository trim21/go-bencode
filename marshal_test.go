package bencode_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trim21/go-bencode"
	"github.com/trim21/go-bencode/internal/encoder"
	"github.com/trim21/go-bencode/internal/test"
)

// make sure they are equal
var _ bencode.Marshaler = encoder.Marshaler(nil)
var _ encoder.Marshaler = bencode.Marshaler(nil)

type Container struct {
	Value any `bencode:"value"`
}

type Inner struct {
	V int    `bencode:"v" json:"v"`
	S string `bencode:"a long string name replace field name" json:"a long string name replace field name"`
}

type TestData struct {
	Users []User                     `bencode:"users" json:"users"`
	Obj   Inner                      `bencode:"obj" json:"obj"`
	B     bool                       `bencode:"ok" json:"ok"`
	Map   map[string]struct{ V int } `bencode:"map" json:"map"`
}

type User struct {
	ID   uint64 `bencode:"id" json:"id"`
	Name string `bencode:"name" json:"name"`
}

type Item struct {
	V int `json:"v" bencode:"v"`
}

type ContainerNonAnonymous struct {
	OK   bool
	Item Item
	V    int
}

type Case struct {
	Name     string
	Data     any
	Expected string `bencode:"-" json:"-"`
}

// map in struct is an indirect ptr
type MapPtr struct {
	Users []Item           `bencode:"users" json:"users"`
	Map   map[string]int64 `bencode:"map" json:"map"`
}

// map in struct is a direct ptr
type MapOnly struct {
	Map map[string]int64 `bencode:"map" json:"map"`
}

var testCase = []Case{
	{Name: "bool true", Data: true, Expected: "i1e"},
	{Name: "*bool true", Data: toPtr(true), Expected: "i1e"},

	{Name: "bool false", Data: false, Expected: "i0e"},
	{Name: "*bool false", Data: new(bool), Expected: "i0e"},

	{
		Name: "*bool-as-string-indirect",
		Data: struct {
			Value *bool `bencode:"value"`
			B     *bool
		}{Value: toPtr(false)},
		Expected: `d5:valuei0ee`,
	},

	{Name: "int8", Data: 0, Expected: "i0e"},
	{Name: "*int8", Data: -0, Expected: "i0e"},

	{Name: "int8", Data: int8(7), Expected: "i7e"},
	{Name: "*int8", Data: toPtr(int8(-7)), Expected: "i-7e"},

	{Name: "int16", Data: int16(7), Expected: "i7e"},
	{Name: "*int16", Data: toPtr(int16(7)), Expected: "i7e"},

	{Name: "int32", Data: int32(7), Expected: "i7e"},
	{Name: "*int32", Data: toPtr(int32(9)), Expected: "i9e"},

	{Name: "int64", Data: int64(7), Expected: "i7e"},
	{Name: "*int64", Data: toPtr[int64](10), Expected: "i10e"},
	{Name: "int", Data: int(8), Expected: "i8e"},
	{Name: "*int", Data: toPtr[int](11), Expected: "i11e"},
	{Name: "uint8", Data: uint8(7), Expected: "i7e"},
	{Name: "*uint8", Data: toPtr[uint8](7), Expected: "i7e"},
	{Name: "uint16", Data: uint16(7), Expected: "i7e"},
	{Name: "*uint16", Data: toPtr[uint16](7), Expected: "i7e"},
	{Name: "uint32", Data: uint32(7), Expected: "i7e"},
	{Name: "*uint32", Data: toPtr[uint32](7), Expected: "i7e"},
	{Name: "uint64", Data: uint64(7777), Expected: "i7777e"},
	{Name: "*uint64", Data: toPtr[uint64](7), Expected: "i7e"},
	{Name: "uint", Data: uint(9), Expected: "i9e"},
	{Name: "*uint", Data: toPtr[uint](787), Expected: "i787e"},
	{Name: "string", Data: `qwer"qwer`, Expected: `9:qwer"qwer`},
	{Name: "*string", Data: toPtr(`qwer"qwer`), Expected: `9:qwer"qwer`},
	{Name: "simple slice", Data: []int{1, 4, 6, 2, 3}, Expected: `li1ei4ei6ei2ei3ee`},
	{
		Name:     "struct-slice",
		Data:     []Item{{V: 6}, {V: 5}, {4}, {3}, {2}},
		Expected: `ld1:vi6eed1:vi5eed1:vi4eed1:vi3eed1:vi2eee`,
	},
	{
		Name:     "struct-with-map-indirect",
		Data:     MapPtr{Users: []Item{}, Map: map[string]int64{"one": 1, "two": 2}},
		Expected: `d3:mapd3:onei1e3:twoi2ee5:userslee`,
	},
	{
		Name:     "struct with map embed",
		Data:     MapOnly{Map: map[string]int64{"one": 1, "two": 2}},
		Expected: `d3:mapd3:onei1e3:twoi2eee`,
	},
	{
		Name:     "empty map",
		Data:     map[string]string{},
		Expected: "de",
	},
	{
		Name:     "nil map",
		Data:     (map[string]string)(nil),
		Expected: `de`,
	},

	{
		Name: "nested struct not anonymous",
		Data: ContainerNonAnonymous{
			OK:   true,
			Item: Item{V: 5},
			V:    9999,
		},
		Expected: `d4:Itemd1:vi5ee2:OKi1e1:Vi9999ee`,
	},

	{
		Name: "struct with all",
		Data: TestData{
			Users: []User{
				{ID: 1, Name: "sai"},
				{ID: 2, Name: "trim21"},
			},
			B:   false,
			Obj: Inner{V: 2, S: "vvv"},

			Map: map[string]struct{ V int }{"7": {V: 4}},
		},
		Expected: `d3:mapd1:7d1:Vi4eee3:objd37:a long string name replace field name3:vvv1:vi2ee2:oki0e5:usersld2:idi1e4:name3:saied2:idi2e4:name6:trim21eee`,
	},

	{
		Name:     "nested_map",
		Data:     map[string]map[string]string{"1": {"4": "ok"}},
		Expected: `d1:1d1:42:okee`,
	},

	{
		Name:     "map[type]any(map)",
		Data:     map[string]any{"1": map[string]string{"4": "ok"}},
		Expected: `d1:1d1:42:okee`,
	},

	{
		Name:     "map[type]any(slice)",
		Data:     map[string]any{"1": []int{3, 1, 4}},
		Expected: `d1:1li3ei1ei4eee`,
	},

	{
		Name:     "map[type]any(struct)",
		Data:     map[string]any{"1": User{}},
		Expected: `d1:1d2:idi0e4:name0:ee`,
	},

	{
		Name: "ignore struct field",
		Data: struct {
			V       int
			Ignored string `bencode:"-"`
		}{
			V:       3,
			Ignored: "vvv",
		},
		Expected: `d1:Vi3ee`,
	},
	{
		Name: "private field",
		Data: struct {
			b bool
			D int
		}{D: 10},
		Expected: `d1:Di10ee`,
	},
	{
		Name: "omitempty",
		Data: struct {
			V string `bencode:",omitempty"`
			D string `bencode:",omitempty"`
		}{D: "d"},
		Expected: `d1:D1:de`,
	},
	{
		Name: "omitempty-ptr",
		Data: struct {
			V *string `bencode:",omitempty"`
			D *string `bencode:",omitempty"`
		}{
			D: new(string),
		},
		Expected: `d1:D0:e`,
	},
}

func TestMarshal_concrete_types(t *testing.T) {
	for _, data := range testCase {
		d := data
		t.Run(d.Name, func(t *testing.T) {
			actual, err := bencode.Marshal(d.Data)
			require.NoError(t, err)

			test.StringEqual(t, data.Expected, string(actual))
		})
	}
}

func TestMarshal_interface(t *testing.T) {
	for _, data := range testCase {
		data := data
		t.Run(data.Name, func(t *testing.T) {
			actual, err := bencode.Marshal(data)
			require.NoError(t, err)

			test.StringEqual(t, data.WrappedExpected(), string(actual))
		})
	}
}

func TestMarshal_interface_ptr(t *testing.T) {
	for _, data := range testCase {
		data := data
		t.Run(data.Name, func(t *testing.T) {
			actual, err := bencode.Marshal(&data.Data)
			require.NoError(t, err)

			test.StringEqual(t, data.Expected, string(actual))
		})
	}
}

func TestMarshal_ptr(t *testing.T) {
	t.Run("int-indirect-no-omit", func(t *testing.T) {
		type Indirect struct {
			A *int `bencode:"a"`
			B *int `bencode:"b"`
		}

		var i int = 50

		actual, err := bencode.Marshal(Indirect{B: &i})
		require.NoError(t, err)
		expected := `d1:bi50ee`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("int-indirect-omitempty", func(t *testing.T) {
		type Indirect struct {
			A *int `bencode:"a"`
			B *int `bencode:"b,omitempty"`
		}

		var i int = 50

		actual, err := bencode.Marshal(Indirect{A: &i})
		require.NoError(t, err)
		expected := `d1:ai50ee`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("int-direct", func(t *testing.T) {
		type Direct struct {
			Value *int `bencode:"value"`
		}

		var i int = 50

		t.Run("encode", func(t *testing.T) {
			actual, err := bencode.Marshal(Direct{Value: &i})
			require.NoError(t, err)
			expected := `d5:valuei50ee`
			test.StringEqual(t, expected, string(actual))
		})
	})

	t.Run("nil", func(t *testing.T) {
		type Data struct {
			Value *int `bencode:"value"`
		}
		var data = Data{}

		actual, err := bencode.Marshal(data)
		require.NoError(t, err)
		expected := `de`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("*string", func(t *testing.T) {
		type Data struct {
			Value *string `bencode:"value"`
		}
		var s = "abcdefg"
		var data = Data{&s}

		actual, err := bencode.Marshal(data)
		require.NoError(t, err)
		expected := `d5:value7:abcdefge`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("multiple ptr", func(t *testing.T) {
		type Data struct {
			Value *string `bencode:"value"`
			D     *int    `bencode:"d,omitempty"`
		}

		var s = "abcdefg"
		var data = Data{Value: &s}

		actual, err := bencode.Marshal(&data)
		require.NoError(t, err)
		expected := `d5:value7:abcdefge`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("struct", func(t *testing.T) {
		t.Run("*struct", func(t *testing.T) {
			type Data struct {
				Value int    `bencode:"value"`
				ID    uint32 `bencode:"id"`
			}
			var data = Data{}

			actual, err := bencode.Marshal(&data)
			require.NoError(t, err)
			expected := `d2:idi0e5:valuei0ee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("*struct-nil", func(t *testing.T) {
			type Data struct {
				Value int    `bencode:"value"`
				ID    uint32 `bencode:"id"`
			}
			var data *Data

			_, err := bencode.Marshal(data)
			require.Error(t, err)
			//expected := `N;`
			//test.StringEqual(t, expected, string(actual))
		})

		t.Run("indirect", func(t *testing.T) {
			type Data struct {
				B     *int  `bencode:"b"`
				Value *User `bencode:"value"`
			}

			var b = 20
			var data = Data{B: &b}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d1:bi20ee`
			test.StringEqual(t, expected, string(actual))
		})

		u := User{
			ID:   4,
			Name: "one",
		}

		t.Run("direct", func(t *testing.T) {
			type Data struct {
				Value *User `bencode:"value"`
			}
			var data = Data{}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `de`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("encode direct", func(t *testing.T) {
			type Data struct {
				Value *User `bencode:"value"`
			}
			var data = Data{Value: &u}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valued2:idi4e4:name3:oneee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("encode indirect", func(t *testing.T) {
			type Data struct {
				B     *int  `bencode:"b"`
				Value *User `bencode:"value"`
			}
			var data = Data{Value: &u}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valued2:idi4e4:name3:oneee`
			test.StringEqual(t, expected, string(actual))
		})
	})

	t.Run("array", func(t *testing.T) {
		t.Run("nil-direct", func(t *testing.T) {
			type Data struct {
				Value *[5]int `bencode:"value"`
			}
			var data = Data{}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `de`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("nil-indirect", func(t *testing.T) {
			type Data struct {
				Value *[5]int `bencode:"value"`
				B     *bool   `bencode:"b"`
			}

			var b = true
			var data = Data{B: &b}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d1:bi1ee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("omitempty", func(t *testing.T) {
			type Data struct {
				Value *[5]int `bencode:"value,omitempty"`
			}
			var s = [5]int{1, 6, 4, 7, 9}
			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valueli1ei6ei4ei7ei9eee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("no omitempty", func(t *testing.T) {
			type Data struct {
				Value *[5]int `bencode:"value"`
			}
			var s = [5]int{1, 6, 4, 7, 9}
			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valueli1ei6ei4ei7ei9eee`
			test.StringEqual(t, expected, string(actual))
		})
	})

	t.Run("slice", func(t *testing.T) {
		t.Run("omitempty", func(t *testing.T) {
			type Data struct {
				Value *[]string `bencode:"value,omitempty"`
			}
			var s = strings.Split("abcdefg", "")
			require.Len(t, s, 7)
			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valuel1:a1:b1:c1:d1:e1:f1:gee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("no-omitempty", func(t *testing.T) {
			type Data struct {
				Value *[]string `bencode:"value"`
			}
			var s = strings.Split("abcdefg", "")
			require.Len(t, s, 7)
			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valuel1:a1:b1:c1:d1:e1:f1:gee`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("nil", func(t *testing.T) {
			type Data struct {
				Value *[]string `bencode:"value"`
			}

			var data = Data{}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `de`
			test.StringEqual(t, expected, string(actual))
		})

		t.Run("encode", func(t *testing.T) {
			type Data struct {
				Value *[]string `bencode:"value"`
			}

			var s = []string{"1", "2"}

			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:valuel1:11:2ee`
			test.StringEqual(t, expected, string(actual))
		})
	})

	t.Run("*string-omitempty", func(t *testing.T) {
		type Data struct {
			Value *string `bencode:"value,omitempty"`
		}

		t.Run("not_nil", func(t *testing.T) {
			var s = "abcdefg"
			var data = Data{&s}

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)
			expected := `d5:value7:abcdefge`
			test.StringEqual(t, expected, string(actual))
		})

	})

	t.Run("struct-map", func(t *testing.T) {
		t.Run("direct", func(t *testing.T) {
			type Data struct {
				Value *map[string]int `bencode:"value"`
			}

			t.Run("nil direct", func(t *testing.T) {
				var data = Data{}
				actual, err := bencode.Marshal(data)
				require.NoError(t, err)
				expected := `de`
				test.StringEqual(t, expected, string(actual))
			})

			t.Run("encode", func(t *testing.T) {
				var s = map[string]int{"b": 2}

				actual, err := bencode.Marshal(&s)
				require.NoError(t, err)
				expected := `d1:bi2ee`
				test.StringEqual(t, expected, string(actual))
			})

			t.Run("omitempty encode", func(t *testing.T) {
				type Data struct {
					Value *map[string]int `bencode:"value,omitempty"`
				}

				var s = map[string]int{"1": 2}
				var data = Data{&s}

				actual, err := bencode.Marshal(data)
				require.NoError(t, err)
				expected := `d5:valued1:1i2eee`
				test.StringEqual(t, expected, string(actual))
			})

			t.Run("omitempty nil", func(t *testing.T) {
				type Data struct {
					Value *map[string]int `bencode:"value,omitempty"`
				}
				var data = Data{}

				actual, err := bencode.Marshal(data)
				require.NoError(t, err)
				expected := `de`
				test.StringEqual(t, expected, string(actual))
			})
		})

		t.Run("indirect", func(t *testing.T) {
			type Data struct {
				Value *map[string]int `bencode:"value"`
				Bool  *bool           `bencode:"b"`
			}

			t.Run("nil direct", func(t *testing.T) {
				var data = Data{}
				actual, err := bencode.Marshal(data)
				require.NoError(t, err)
				expected := `de`
				test.StringEqual(t, expected, string(actual))
			})

			t.Run("encode", func(t *testing.T) {
				var s = map[string]int{"q": 2}

				actual, err := bencode.Marshal(&s)
				require.NoError(t, err)
				expected := `d1:qi2ee`
				test.StringEqual(t, expected, string(actual))
			})

			t.Run("omitempty", func(t *testing.T) {
				type Data struct {
					Value *map[string]int `bencode:"value,omitempty"`
					Bool  *bool           `bencode:"b"`
				}

				var s = map[string]int{"a": 2}
				var data = Data{Value: &s}

				actual, err := bencode.Marshal(data)
				require.NoError(t, err)
				expected := `d5:valued1:ai2eee`
				test.StringEqual(t, expected, string(actual))
			})
		})
	})

	t.Run("map", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			type Data struct {
				Value *map[string]int `bencode:"value"`
			}

			var data = Data{}

			_, err := bencode.Marshal(data.Value)
			require.Error(t, err)
		})

		t.Run("encode", func(t *testing.T) {
			var s = map[string]int{"x": 2}

			actual, err := bencode.Marshal(&s)
			require.NoError(t, err)
			expected := `d1:xi2ee`
			test.StringEqual(t, expected, string(actual))
		})
	})

	t.Run("int", func(t *testing.T) {
		type Data struct {
			Value *int `bencode:"value"`
		}
		var s = 644
		var data = Data{&s}

		actual, err := bencode.Marshal(data)
		require.NoError(t, err)
		expected := `d5:valuei644ee`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("nested", func(t *testing.T) {
		type Container struct {
			Value ***uint `bencode:"value"`
		}

		var v uint = 8
		var p = &v
		var a = &p

		_, err := bencode.Marshal(Container{Value: &a})
		require.Error(t, err)
	})

	t.Run("recursive", func(t *testing.T) {
		type Container struct {
			Value any `bencode:"value"`
		}

		var v uint = 8
		var p = &v
		var a any = &p

		expected := `d5:valuei8ee`
		actual, err := bencode.Marshal(Container{Value: &a})
		require.NoError(t, err)
		test.StringEqual(t, expected, string(actual))
	})
}

func TestMarshal_map(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		// map in struct is a direct ptr
		type MapOnly struct {
			Map map[string]int64 `bencode:"map" json:"map"`
		}
		actual, err := bencode.Marshal(MapOnly{Map: nil})
		require.NoError(t, err)
		expected := `d3:mapdee`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("direct", func(t *testing.T) {
		// map in struct is a direct ptr
		type MapOnly struct {
			Map map[string]int64 `bencode:"map" json:"map"`
		}
		actual, err := bencode.Marshal(MapOnly{Map: map[string]int64{"abcdef": 1}})
		require.NoError(t, err)
		expected := `d3:mapd6:abcdefi1eee`
		test.StringEqual(t, expected, string(actual))
	})

	t.Run("indirect", func(t *testing.T) {
		// map in struct is an indirect ptr
		type MapPtr struct {
			Users []Item           `bencode:"users" json:"users"`
			Map   map[string]int64 `bencode:"map" json:"map"`
		}

		actual, err := bencode.Marshal(MapPtr{Map: map[string]int64{"abcdef": 1}})
		require.NoError(t, err)
		expected := `d3:mapd6:abcdefi1ee5:userslee`
		test.StringEqual(t, expected, string(actual))
	})
}

type M interface {
	Bool() bool
}

type mImpl struct {
}

func (m mImpl) Bool() bool {
	return true
}

func TestMarshal_interface_with_method(t *testing.T) {
	var data M = mImpl{}
	actual, err := bencode.Marshal(Container{Value: data})
	require.NoError(t, err)
	expected := `d5:valuedee`
	test.StringEqual(t, expected, string(actual))
}

func TestMarshal_anonymous_field(t *testing.T) {
	type N struct {
		A int
		B int
	}

	type M struct {
		N
		C int
	}

	_, err := bencode.Marshal(M{N: N{
		A: 3,
		B: 2,
	}, C: 1})
	require.Error(t, err)
	require.Regexp(t, regexp.MustCompile("supported for Anonymous struct field has been removed.*"), err.Error())
}

func TestRecursivePanic(t *testing.T) {

	type O struct {
		Name string
		E    []O
	}

	actual, err := bencode.Marshal(O{
		Name: "hello",
		E: []O{
			{
				Name: "BB",
				E: []O{
					{Name: "C C D D E E F F"},
				},
			},
		},
	})
	require.NoError(t, err)
	expected := `d1:Eld1:Eld1:Ele4:Name15:C C D D E E F Fee4:Name2:BBee4:Name5:helloe`
	test.StringEqual(t, expected, string(actual))
}

type userMarshaler struct {
	t time.Time
}

func (u userMarshaler) MarshalBencode() ([]byte, error) {
	return bencode.Marshal(u.t.Format(time.RFC3339))
}

var _ bencode.Marshaler = userMarshaler{}

func TestUserMarshaler(t *testing.T) {

	now, err := time.Parse(time.RFC3339, "2024-07-16T01:02:03+08:00")
	require.NoError(t, err)

	type O struct {
		T userMarshaler
	}

	actual, err := bencode.Marshal(O{
		T: userMarshaler{
			t: now,
		},
	})
	require.NoError(t, err)

	test.StringEqual(t, `d1:T25:2024-07-16T01:02:03+08:00e`, string(actual))
}

type Generic[T any] struct {
	Value T
}

type Generic2[T any] struct {
	B     bool // prevent direct
	Value T
}

func (tc Case) WrappedExpected() string {
	return fmt.Sprintf(`d4:Data%s4:Name%d:%se`, tc.Expected, len(tc.Name), tc.Name)
}

var go118TestCase = []Case{
	{
		Name:     "generic[int]",
		Data:     Generic[int]{1},
		Expected: `d5:Valuei1ee`,
	},
	{
		Name:     "generic[struct]",
		Data:     Generic[User]{User{}},
		Expected: `d5:Valued2:idi0e4:name0:ee`,
	},
	{
		Name:     "generic[map]",
		Data:     Generic[map[string]int]{map[string]int{"one": 1, "two": 2}},
		Expected: `d5:Valued3:onei1e3:twoi2eee`,
	},
	{
		Name:     "generic[slice]",
		Data:     Generic[[]string]{[]string{"hello", "world"}},
		Expected: `d5:Valuel5:hello5:worldee`,
	},
	{
		Name:     "generic2[slice]",
		Data:     Generic2[[]string]{Value: []string{"hello", "world"}},
		Expected: `d1:Bi0e5:Valuel5:hello5:worldee`,
	},
}

func TestMarshal_go118_concrete_types(t *testing.T) {

	for _, data := range go118TestCase {
		data := data
		t.Run(data.Name, func(t *testing.T) {
			actual, err := bencode.Marshal(data.Data)
			require.NoError(t, err)

			test.StringEqual(t, data.Expected, string(actual))
		})
	}
}

func TestMarshal_go118_interface(t *testing.T) {

	for _, data := range go118TestCase {
		data := data
		t.Run(data.Name, func(t *testing.T) {

			actual, err := bencode.Marshal(data)
			require.NoError(t, err)

			test.StringEqual(t, data.WrappedExpected(), string(actual))
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}

func TestMarshal_array_map(t *testing.T) {
	var data = [5]map[string]uint{
		{"-3": 1},
		nil,
		{"-1": 1},
	}

	actual, err := bencode.Marshal(data)
	require.NoError(t, err)
	expected := `ld2:-3i1eeded2:-1i1eededee`
	test.StringEqual(t, expected, string(actual))
}

func TestMarshal_Array_nil(t *testing.T) {
	var data [5]int

	actual, err := bencode.Marshal(data)
	require.NoError(t, err)
	expected := `li0ei0ei0ei0ei0ee`
	test.StringEqual(t, expected, string(actual))
}
