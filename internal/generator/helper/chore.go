package helper

import (
	"strconv"

	. "github.com/dave/jennifer/jen"
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

func CodeIf(condition bool, code Code) Code {
	if condition {
		return code
	}
	return Null()
}
