package helper

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
)

func IntToStr(i int) string {
	return strconv.Itoa(i)
}

func StrPtr(s string) *string {
	return &s
}

func StrIf(condition bool, str string) string {
	if condition {
		return str
	}
	return ""
}

func StrOrEmpty(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func CodeIf(condition bool, code Code) Code {
	if condition {
		return code
	}
	return Null()
}

func BytesStrToBytes(str string) []byte {
	values := strings.Split(strings.TrimSuffix(strings.TrimPrefix(str, "["), "]"), ",")
	bytes := make([]byte, len(values))
	for i, v := range values {
		v = strings.Trim(v, " ")
		b, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(err, str)))
		}
		bytes[i] = byte(b)
	}
	return bytes
}
