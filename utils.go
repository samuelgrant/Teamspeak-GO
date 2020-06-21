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
		// `/`, `\/`,
		// ` `, `\s`,
		// `|`, `\p`,
		// "\a", `\a`,
		// "\b", `\b`,
		// "\f", `\f`,
		// "\n", `\n`,
		// "\r", `\r`,
		// "\t", `\t`,
		// "\v", `\v`,
	)

	webEncoder = strings.NewReplacer(
		` `, `%20`,
		`=`, `%3D`,
	)

	// decoder performs white space and special character decoding
	// as required by the ServerQuery protocol.
	decoder = strings.NewReplacer(
		`+`, " ",
		// `\/`, "/",
		// `\s`, " ",
		// `\p`, "|",
		// `\a`, "\a",
		// `\b`, "\b",
		// `\f`, "\f",
		// `\n`, "\n",
		// `\r`, "\r",
		// `\t`, "\t",
		// `\v`, "\v",
	)
)

type QueryResponse struct {
	Id        int64
	Msg       string
	IsSuccess bool
}

// Decode and return the string
func Decode(s string) string {
	return decoder.Replace(s)
}

func Encode(s string) string {
	return encoder.Replace(s)
}

func WebEncode(s string) string {
	return webEncoder.Replace(s)
}

func GetVal(s string) string {
	return strings.Split(s, "=")[1]
}

func ParseQueryResponse(res string) (*QueryResponse, string, error) {
	lines := strings.Split(res, "\n")

	qr_lines := strings.Split(lines[len(lines)-2], " ")
	qr_res_id, err := strconv.ParseInt(GetVal(qr_lines[1]), 10, 64)
	if err != nil {
		return nil, "", err
	}

	qr_obj := QueryResponse{
		Id:        qr_res_id,
		Msg:       Decode(GetVal(qr_lines[2])),
		IsSuccess: GetVal(qr_lines[2]) == "ok",
	}

	if len(lines) <= 2 {
		return &qr_obj, "", nil
	}

	return &qr_obj, lines[0], nil
}

// Converts int64 to a string
func i64tostr(i int64) string {
	return strconv.FormatInt(i, 10)
}
