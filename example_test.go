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
		Value map[string]string `php:"value" json:"value"`
	}
	raw := `a:1:{s:5:"value";a:5:{s:3:"one";s:1:"1";s:3:"two";s:1:"2";s:5:"three";s:1:"3";s:4:"four";s:1:"4";s:4:"five";s:1:"5";}}`

	err := bencode.Unmarshal([]byte(raw), &v)
	if err != nil {
		panic(err)
	}

	fmt.Println(v.Value["five"])
	// Output: 5
}
