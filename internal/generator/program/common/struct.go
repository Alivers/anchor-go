package common

import (
	"fmt"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	ag_binary "github.com/gagliardetto/binary"
)

func GenerateMarshalWithEncoderForStruct(
	ctx *model.GenerateCtx,
	marshalReceiverName string,
	fields []idl.IdlField,
	structDiscriminatorName *string,
	checkFieldNil bool,
	program *idl.Idl,
) Code {
	code := Empty()
	code.Func().Params(Id("obj").Id(marshalReceiverName)).Id("MarshalWithEncoder").
		Params(
			Id("encoder").Op("*").Qual(model.PkgDfuseBinary, "Encoder"),
		).
		Params(
			Err().Error(),
		).
		BlockFunc(func(body *Group) {
			if structDiscriminatorName != nil && *structDiscriminatorName != "" {
				body.Comment("Write account discriminator:")
				body.Err().Op("=").Id("encoder").Dot("WriteBytes").Call(Id(*structDiscriminatorName).Index(Op(":")), False())
				body.If(Err().Op("!=").Nil()).Block(
					Return(Err()),
				)
			}

			for _, field := range fields {
				exportedArgName := helper.ToCamelCase(field.Name)
				if field.Type.IsOption() {
					body.Commentf("Serialize `%s` param (optional):", exportedArgName)
				} else {
					body.Commentf("Serialize `%s` param:", exportedArgName)
				}

				if ctx.IsComplexEnumByType(&field.Type) {
					// Must be a defined type ref.
					enumTypeName := field.Type.GetDefined().Name
					body.BlockFunc(func(argBody *Group) {
						argBody.List(Id("tmp")).Op(":=").Id(GetEnumVariantsContainerName(enumTypeName)).Block()
						argBody.Switch(Id("realvalue").Op(":=").Id("obj").Dot(exportedArgName).Op(".").Parens(Type())).
							BlockFunc(func(switchGroup *Group) {
								interfaceType := program.FindTypeByName(enumTypeName)
								for variantIndex, variant := range interfaceType.Type.Variants {
									variantTypeNameStruct := GetComplexEnumVariantTypeName(enumTypeName, variant.Name)

									switchGroup.Case(Op("*").Id(variantTypeNameStruct)).
										BlockFunc(func(caseGroup *Group) {
											caseGroup.Id("tmp").Dot("Enum").Op("=").Lit(variantIndex)
											caseGroup.Id("tmp").Dot(helper.ToCamelCase(variant.Name)).Op("=").Op("*").Id("realvalue")
										})
								}
							})

						argBody.Err().Op(":=").Id("encoder").Dot("Encode").Call(Id("tmp"))

						argBody.If(
							Err().Op("!=").Nil(),
						).Block(
							Return(Err()),
						)

					})
				} else {
					if field.Type.IsOption() {
						if checkFieldNil {
							body.BlockFunc(func(optGroup *Group) {
								// if nil:
								optGroup.If(Id("obj").Dot(helper.ToCamelCase(field.Name)).Op("==").Nil()).Block(
									Err().Op("=").Id("encoder").Dot("WriteBool").Call(False()),
									If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									),
								).Else().Block(
									Err().Op("=").Id("encoder").Dot("WriteBool").Call(True()),
									If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									),
									Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName)),
									If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									),
								)
							})
						} else {
							body.BlockFunc(func(optGroup *Group) {
								// Write as if not nil:
								optGroup.Err().Op("=").Id("encoder").Dot("WriteBool").Call(True())
								optGroup.If(Err().Op("!=").Nil()).Block(
									Return(Err()),
								)
								optGroup.Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName))
								optGroup.If(Err().Op("!=").Nil()).Block(
									Return(Err()),
								)
							})
						}

					} else {
						body.Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName))
						body.If(Err().Op("!=").Nil()).Block(
							Return(Err()),
						)
					}
				}

			}

			body.Return(Nil())
		})
	return code
}

func GenerateUnmarshalWithDecoderForStruct(
	ctx *model.GenerateCtx,
	marshalReceiverName string,
	fields []idl.IdlField,
	structDiscriminatorName *string,
	structDiscriminator *ag_binary.TypeID,
	program *idl.Idl,
) Code {
	code := Empty()
	code.Func().Params(Id("obj").Op("*").Id(marshalReceiverName)).Id("UnmarshalWithDecoder").
		Params(
			Id("decoder").Op("*").Qual(model.PkgDfuseBinary, "Decoder"),
		).
		Params(
			Err().Error(),
		).
		BlockFunc(func(body *Group) {
			if structDiscriminatorName != nil && *structDiscriminatorName != "" {
				body.Comment("Read and check account discriminator:")
				body.BlockFunc(func(discReadBody *Group) {
					discReadBody.List(Id("discriminator"), Err()).Op(":=").Id("decoder").Dot("ReadTypeID").Call()
					discReadBody.If(Err().Op("!=").Nil()).Block(
						Return(Err()),
					)
					discReadBody.If(Op("!").Id("discriminator").Dot("Equal").Call(Id(*structDiscriminatorName).Index(Op(":")))).Block(
						Return(
							Qual("fmt", "Errorf").Call(
								Line().Lit("wrong discriminator: wanted %s, got %s"),
								Line().Lit(fmt.Sprintf("%v", structDiscriminator[:])),
								Line().Qual("fmt", "Sprint").Call(Id("discriminator").Index(Op(":"))),
							),
						),
					)
				})
			}

			for _, field := range fields {
				exportedArgName := helper.ToCamelCase(field.Name)
				if field.Type.IsOption() {
					body.Commentf("Deserialize `%s` (optional):", exportedArgName)
				} else {
					body.Commentf("Deserialize `%s`:", exportedArgName)
				}

				if ctx.IsComplexEnumByType(&field.Type) {
					enumName := field.Type.GetDefined().Name
					body.BlockFunc(func(argBody *Group) {

						argBody.List(Id("tmp")).Op(":=").New(Id(GetEnumVariantsContainerName(enumName)))

						argBody.Err().Op(":=").Id("decoder").Dot("Decode").Call(Id("tmp"))

						argBody.If(
							Err().Op("!=").Nil(),
						).Block(
							Return(Err()),
						)

						argBody.Switch(Id("tmp").Dot("Enum")).
							BlockFunc(func(switchGroup *Group) {
								interfaceType := program.FindTypeByName(enumName)
								for variantIndex, variant := range interfaceType.Type.Variants {
									variantTypeNameComplex := GetComplexEnumVariantTypeName(enumName, variant.Name)

									if variant.IsUint8Variant() {
										switchGroup.Case(Lit(variantIndex)).
											BlockFunc(func(caseGroup *Group) {
												caseGroup.Id("obj").Dot(exportedArgName).Op("=").
													Parens(Op("*").Id(variantTypeNameComplex)).
													Parens(Op("&").Id("tmp").Dot(variant.Name))
											})
									} else {
										switchGroup.Case(Lit(variantIndex)).
											BlockFunc(func(caseGroup *Group) {
												caseGroup.Id("obj").Dot(exportedArgName).Op("=").Op("&").Id("tmp").Dot(helper.ToCamelCase(variant.Name))
											})
									}
								}
								switchGroup.Default().
									BlockFunc(func(caseGroup *Group) {
										caseGroup.Return(Qual("fmt", "Errorf").Call(Lit("unknown enum index: %v"), Id("tmp").Dot("Enum")))
									})
							})

					})
				} else {
					if field.Type.IsOption() {
						body.BlockFunc(func(optGroup *Group) {
							// For optional fields, we need to check if there is remaining data.
							// then read the bool to check if the field is nil or not.
							// if not nil(bool is true), then read the field value.
							optGroup.If(Op("!").Id("decoder").Dot("HasRemaining").Call()).Block(
								Return(Nil()),
							)
							// if nil:
							optGroup.List(Id("ok"), Err()).Op(":=").Id("decoder").Dot("ReadBool").Call()
							optGroup.If(Err().Op("!=").Nil()).Block(
								Return(Err()),
							)
							optGroup.If(Id("ok")).Block(
								Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName)),
								If(Err().Op("!=").Nil()).Block(
									Return(Err()),
								),
							)
						})
					} else if field.Type.IsSimple() && field.Type.GetSimple() == idl.IdlTypeSimpleBool {
						// Special case for bool, we need to check if there is remaining data.
						body.If(Op("!").Id("decoder").Dot("HasRemaining").Call()).Block(
							Return(Nil()),
						)
						body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName))
						body.If(Err().Op("!=").Nil()).Block(
							Return(Err()),
						)
					} else {
						body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName))
						body.If(Err().Op("!=").Nil()).Block(
							Return(Err()),
						)
					}
				}
			}

			body.Return(Nil())
		})
	return code
}
