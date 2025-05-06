package idlcode

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

type FieldCodeOption struct {
	AsPointer   bool
	ComplexEnum bool
}

func IdlFieldToCode(field idl.IdlField, options FieldCodeOption) Code {
	code := Id(helper.ToCamelCase(field.Name)).
		Add(func() Code {
			if options.ComplexEnum {
				return nil
			}
			if options.AsPointer {
				return Op("*")
			}
			return nil
		}()).
		Add(IdlTypeToCode(field.Type))

	if field.Type.IsOption() {
		code.Add(Tag(map[string]string{
			"bin": "optional",
		}))
	}

	return code
}
