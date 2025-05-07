package types

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateTypes(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)

	// Generate types for all IDL types
	for _, typ := range program.Types {
		identType := ctx.GetIdentifierTy(typ.Name)
		if identType == nil {
			panic("type " + typ.Name + " not found in IDL types")
		}

		typeName := typ.Name
		if ctx.IsGeneratedIdentifier(typ.Name) {
			typeName += "Struct"
			file.Comment(typ.Name + " conflict with other type, so we add `Struct` suffix.")
		}

		file.Add(
			common.GenerateTypeDefCode(
				ctx,
				typeName,
				identType,
				nil,
				program,
			),
		)
	}

	return file
}
