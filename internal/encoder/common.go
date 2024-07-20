package encoder

func appendEmptyMap(b []byte) []byte {
	return append(b, "de"...)
}

func appendEmptyList(b []byte) []byte {
	return append(b, "le"...)
}
