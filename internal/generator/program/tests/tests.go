package tests

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateTests(ctx *model.GenerateCtx, program *idl.Idl, instruction *idl.IdlInstruction) *File {
	file := helper.NewGoFile(ctx)

	genTestingFuncsForInstruction(ctx, file, instruction, program)

	return file
}

func genTestingFuncsForInstruction(ctx *model.GenerateCtx, file *File, instruction *idl.IdlInstruction, program *idl.Idl) {
	instExportedName := helper.ToCamelCase(instruction.Name)
	// Declare test: encode, decode:
	file.Func().Id("TestEncodeDecode_" + instExportedName).
		Params(
			ListFunc(func(params *Group) {
				// Parameters:
				params.Id("t").Op("*").Qual("testing", "T")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				// Results:
			}),
		).
		BlockFunc(func(body *Group) {
			// Body:
			body.Id("fu").Op(":=").Qual(model.PkgGoFuzz, "New").Call().Dot("NilChance").Call(Lit(0))

			body.For(
				Id("i").Op(":=").Lit(0),
				Id("i").Op("<").Lit(1),
				Id("i").Op("++"),
			).BlockFunc(func(forGroup *Group) {
				forGroup.Id("t").Dot("Run").Call(
					Lit(instExportedName).Op("+").Qual("strconv", "Itoa").Call(Id("i")),
					Func().Params(Id("t").Op("*").Qual("testing", "T")).Block(
						BlockFunc(func(tFunGroup *Group) {

							if isAnyFieldComplexEnum(ctx, instruction.Args...) {
								genTestWithComplexEnum(ctx, tFunGroup, instExportedName, instruction, program)
							} else {
								genTestNOComplexEnum(tFunGroup, instExportedName)
							}

						}),
					),
				)
			})
		})
}

func genTestNOComplexEnum(tFunGroup *Group, insExportedName string) {
	tFunGroup.Id("params").Op(":=").New(Id(insExportedName))

	tFunGroup.Id("fu").Dot("Fuzz").Call(Id("params"))
	tFunGroup.Id("params").Dot("AccountMetaSlice").Op("=").Nil()

	tFunGroup.Id("buf").Op(":=").New(Qual("bytes", "Buffer"))
	tFunGroup.Id("err").Op(":=").Id("encodeT").Call(Op("*").Id("params"), Id("buf"))
	tFunGroup.Qual(model.PkgTestifyRequire, "NoError").Call(Id("t"), Err())

	tFunGroup.Comment("//")

	tFunGroup.Id("got").Op(":=").New(Id(insExportedName))
	tFunGroup.Id("err").Op("=").Id("decodeT").Call(Id("got"), Id("buf").Dot("Bytes").Call())
	tFunGroup.Id("got").Dot("AccountMetaSlice").Op("=").Nil()
	tFunGroup.Qual(model.PkgTestifyRequire, "NoError").Call(Id("t"), Err())
	tFunGroup.Qual(model.PkgTestifyRequire, "Equal").Call(Id("t"), Id("params"), Id("got"))
}

func genTestWithComplexEnum(ctx *model.GenerateCtx, tFunGroup *Group, insExportedName string, instruction *idl.IdlInstruction, program *idl.Idl) {
	// Create a test for each complex enum argument:
	for _, arg := range instruction.Args {
		if !ctx.IsComplexEnumByType(&arg.Type) {
			continue
		}
		exportedArgName := helper.ToCamelCase(arg.Name)

		tFunGroup.BlockFunc(func(enumBlock *Group) {
			enumName := arg.Type.GetDefined().Name
			interfaceType := program.FindTypeByName(enumName)
			for _, variant := range interfaceType.Type.GetEnum().Variants {
				enumBlock.BlockFunc(func(variantBlock *Group) {
					variantBlock.Id("params").Op(":=").New(Id(insExportedName))

					variantBlock.Id("fu").Dot("Fuzz").Call(Id("params"))
					variantBlock.Id("params").Dot("AccountMetaSlice").Op("=").Nil()
					variantBlock.Id("tmp").Op(":=").New(Id(common.GetComplexEnumVariantTypeName(enumName, variant.Name)))
					variantBlock.Id("fu").Dot("Fuzz").Call(Id("tmp"))
					variantBlock.Id("params").Dot("Set" + exportedArgName).Call(Id("tmp"))

					variantBlock.Id("buf").Op(":=").New(Qual("bytes", "Buffer"))
					variantBlock.Id("err").Op(":=").Id("encodeT").Call(Op("*").Id("params"), Id("buf"))
					variantBlock.Qual(model.PkgTestifyRequire, "NoError").Call(Id("t"), Err())

					variantBlock.Comment("//")

					variantBlock.Id("got").Op(":=").New(Id(insExportedName))
					variantBlock.Id("err").Op("=").Id("decodeT").Call(Id("got"), Id("buf").Dot("Bytes").Call())
					variantBlock.Id("got").Dot("AccountMetaSlice").Op("=").Nil()
					variantBlock.Qual(model.PkgTestifyRequire, "NoError").Call(Id("t"), Err())

					variantBlock.Comment("to prevent garbage buffer fill by fuzz")
					variantBlock.If(Qual("reflect", "TypeOf").Call(Op("*").Id("tmp")).Dot("Kind").Call().Op("!=").Qual("reflect", "Struct")).Block(
						Id("got").Dot(exportedArgName).Op("=").Id("params").Dot(exportedArgName),
					)

					variantBlock.Qual(model.PkgTestifyRequire, "Equal").Call(Id("t"), Id("params"), Id("got"))
				})
			}

		})
	}
}

func isAnyFieldComplexEnum(ctx *model.GenerateCtx, envelopes ...idl.IdlField) bool {
	for _, v := range envelopes {
		if ctx.IsComplexEnumByType(&v.Type) {
			return true
		}
	}
	return false
}
