package errors

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateErrors(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)

	file.Add(Var().DefsFunc(func(group *Group) {
		errDict := Dict{}
		for _, errDef := range program.Errors {
			name := "Err" + helper.ToCamelCase(errDef.Name)
			group.Add(Id(name).Op("=").Op("&").Id("customErrorDef").Values(Dict{
				Id("code"): Lit(errDef.Code),
				Id("name"): Lit(errDef.Name),
				Id("msg"):  Lit(errDef.Msg),
			}))
			errDict[Lit(errDef.Code)] = Id(name)
		}
		group.Add(Id("Errors").Op("=").Map(Int()).Id("CustomError").Values(errDict))
	}))
	generateErrorSnippet(file)

	return file
}
