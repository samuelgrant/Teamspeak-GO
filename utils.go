package ts3

import (
	"strconv"
	"strings"
)

var (
	// encoder performs white space and special character encoding
	// as required by the ServerQuery protocol.
	encoder = strings.NewReplacer(
		` `, `+`,
	)

	webEncoder = strings.NewReplacer(
		` `, `%20`,
		`=`, `%3D`,
	)

	// decoder performs white space and special character decoding
	// as required by the ServerQuery protocol.
	decoder = strings.NewReplacer(
		`+`, " ",
	)
)

func Encode(s string) string {
	return encoder.Replace(s)
}

func WebEncode(s string) string {
	return webEncoder.Replace(s)
}

// Converts int64 to a string
func i64tostr(i int64) string {
	return strconv.FormatInt(i, 10)
}
