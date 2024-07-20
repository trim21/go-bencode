package bencode_test

import (
	"fmt"

	"github.com/trim21/go-bencode"
)

func ExampleMarshal() {
	type User struct {
		ID   uint32 `bencode:"id,string"`
		Name string `bencode:"name"`
	}

	type Inner struct {
		V int    `bencode:"v"`
		S string `bencode:"a long string name replace field name"`
	}

	type With struct {
		Users   []User `bencode:"users,omitempty"`
		Obj     Inner  `bencode:"obj"`
		Ignored bool   `bencode:"-"`
	}

	var data = With{
		Users: []User{
			{ID: 1, Name: "sai"},
			{ID: 2, Name: "trim21"},
		},
		Obj: Inner{V: 2, S: "vvv"},
	}
	var b, err = bencode.Marshal(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
	// Output: d3:objd37:a long string name replace field name3:vvv1:vi2ee5:usersld2:idi1e4:name3:saied2:idi2e4:name6:trim21eee
}

func ExampleUnmarshal() {
	var v struct {
		S     map[[20]byte]struct{} `bencode:"s"`
		Value map[string]string     `bencode:"value" json:"value"`
	}
	raw := `d5:valued4:five1:5ee`

	err := bencode.Unmarshal([]byte(raw), &v)
	if err != nil {
		panic(err)
	}

	fmt.Println(v.Value["five"])
	// Output: 5
}
