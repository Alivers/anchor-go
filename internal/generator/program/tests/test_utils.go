package tests

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	. "github.com/dave/jennifer/jen"
)

func GenerateTestUtils(ctx *model.GenerateCtx) *File {
	file := helper.NewGoFile(ctx)

	generateTestUtils(ctx, file)

	return file
}

func generateTestUtils(ctx *model.GenerateCtx, file *File) {
	file.Func().Id("encodeT").
		Params(
			Id("data").Interface(),
			Id("buf").Op("*").Qual("bytes", "Buffer"),
		).
		Params(
			Error(),
		).
		BlockFunc(func(body *Group) {
			body.If(
				Err().Op(":=").Qual(model.PkgDfuseBinary, ctx.Encoder.GetNewEncoderName()).Call(Id("buf")).Dot("Encode").Call(Id("data")),
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual("fmt", "Errorf").Call(Lit("unable to encode instruction: %w"), Err())),
			)
			body.Return(Nil())
		})

	file.Func().Id("decodeT").
		Params(
			Id("dst").Interface(),
			Id("data").Index().Byte(),
		).
		Params(
			Error(),
		).
		BlockFunc(func(body *Group) {
			body.Return(Qual(model.PkgDfuseBinary, ctx.Encoder.GetNewDecoderName()).Call(Id("data")).Dot("Decode").Call(Id("dst")))
		})
}
