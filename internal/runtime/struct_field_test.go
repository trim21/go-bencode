package runtime

import (
	"reflect"
	"testing"
)

type tagFixture struct {
	Default    int
	Custom     int `bencode:"custom,omitempty,unknown"`
	Ignored    int `bencode:"-"`
	Invalid    int `bencode:"bad\\tag"`
	unexported int
}

func TestStructFieldMetadata(t *testing.T) {
	rt := reflect.TypeFor[tagFixture]()

	if IsIgnoredStructField(rt.Field(0)) {
		t.Fatal("exported untagged field must not be ignored")
	}
	if !IsIgnoredStructField(rt.Field(2)) {
		t.Fatal("field tagged with '-' must be ignored")
	}
	if !IsIgnoredStructField(rt.Field(4)) {
		t.Fatal("unexported field must be ignored")
	}

	defaultTag := StructTagFromField(rt.Field(0))
	if got := defaultTag.Name(); got != "Default" {
		t.Fatalf("default field name = %q, want %q", got, "Default")
	}
	if got := (StructTag{Field: rt.Field(0)}).Name(); got != "Default" {
		t.Fatalf("empty key field name = %q, want %q", got, "Default")
	}

	customTag := StructTagFromField(rt.Field(1))
	if got := customTag.Name(); got != "custom" {
		t.Fatalf("custom field name = %q, want %q", got, "custom")
	}
	if !customTag.IsOmitEmpty {
		t.Fatal("omitempty option was not detected")
	}

	invalidTag := StructTagFromField(rt.Field(3))
	if got := invalidTag.Name(); got != "Invalid" {
		t.Fatalf("invalid tag should fall back to field name; got %q", got)
	}
}

func TestValidTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{name: "empty", tag: "", want: false},
		{name: "letters and digits", tag: "\u5b57\u6bb542", want: true},
		{name: "allowed punctuation", tag: "a/b:c? d", want: true},
		{name: "quote", tag: `a"b`, want: false},
		{name: "backslash", tag: `a\b`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTag(tt.tag); got != tt.want {
				t.Fatalf("isValidTag(%q) = %t, want %t", tt.tag, got, tt.want)
			}
		})
	}
}
