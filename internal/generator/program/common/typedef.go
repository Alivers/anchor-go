package common

import (
	"fmt"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/idlcode"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
)

func GenerateTypeDefCode(ctx *model.GenerateCtx, typeName string, typeDef *idl.IdlTypeDefTy, anchorDiscriminator *[8]byte, program *idl.Idl) Code {
	code := Empty()
	switch {
	case typeDef.IsStruct():
		structDef := typeDef.GetStruct()
		exportedStructName := helper.ToCamelCase(typeName)
		code.Add(generateStructTypeDefCode(ctx, exportedStructName, structDef, anchorDiscriminator, program)).Line()
	case typeDef.IsEnum():
		enumDef := typeDef.GetEnum()
		enumTypeName := typeName

		if enumDef.IsUint8Enum() {
			code.Add(generateUint8EnumCode(enumTypeName, enumDef)).Line()
		} else {
			code.Add(generateComplexEnumCode(ctx, enumTypeName, enumDef, program)).Line()
		}
	case typeDef.IsType():
		aliasDef := typeDef.GetType()
		code.Type().Id(typeName).Op("=").Add(idlcode.IdlTypeToCode(aliasDef.Alias)).Line()
	default:
		panic(fmt.Sprintf("not implemented: %s", spew.Sdump(typeDef)))
	}
	return code
}

func generateUint8EnumCode(enumTypeName string, enumDef *idl.IdlTypeDefTyEnum) Code {
	code := Empty()
	code.Type().Id(enumTypeName).Qual(model.PkgDfuseBinary, "BorshEnum")
	code.Line().Const().DefsFunc(func(gr *Group) {
		for variantIndex, variant := range enumDef.Variants {
			gr.Id(formatSimpleEnumVariantName(enumTypeName, variant.Name)).Add(func() Code {
				if variantIndex == 0 {
					return Id(enumTypeName).Op("=").Iota()
				}
				return nil
			}()).Line()
		}
	}).Line()

	// Generate stringer for the uint8 enum values:
	code.Line().Func().Params(Id("value").Id(enumTypeName)).Id("String").
		Params().
		Params(String()).
		BlockFunc(func(body *Group) {
			body.Switch(Id("value")).BlockFunc(func(switchBlock *Group) {
				for _, variant := range enumDef.Variants {
					switchBlock.Case(Id(formatSimpleEnumVariantName(enumTypeName, variant.Name))).Line().Return(Lit(variant.Name))
				}
				switchBlock.Default().Line().Return(Lit(""))
			})

		}).Line()
	return code
}

func generateComplexEnumCode(ctx *model.GenerateCtx, enumTypeName string, enumDef *idl.IdlTypeDefTyEnum, program *idl.Idl) Code {
	containerName := GetEnumVariantsContainerName(enumTypeName)
	interfaceMethodName := GetComplexEnumInterfaceMethodName(enumTypeName)

	code := Type().Id(enumTypeName).Interface(
		Id(interfaceMethodName).Call(),
	).Line().Line()

	enumVariantNames := make([]string, len(enumDef.Variants))

	// Declare the enum variants container (non-exported, used internally)
	code.Type().Id(containerName).StructFunc(
		func(structGroup *Group) {
			structGroup.Id("Enum").Qual(model.PkgDfuseBinary, "BorshEnum").Tag(map[string]string{
				"borsh_enum": "true",
			})

			for i, variant := range enumDef.Variants {
				enumVariantNames[i] = GetComplexEnumVariantTypeName(enumTypeName, variant.Name)
				structGroup.Id(helper.ToCamelCase(variant.Name)).Id(enumVariantNames[i])
			}
		},
	).Line().Line()

	for variantIndex, variant := range enumDef.Variants {
		variantTypeNameComplex := enumVariantNames[variantIndex]
		var complexVariantFields []idl.IdlField

		// Declare the enum variant types:
		if variant.IsUint8Variant() {
			code.Type().Id(variantTypeNameComplex).Uint8()
		} else {
			code.Type().Id(variantTypeNameComplex).StructFunc(
				func(structGroup *Group) {
					switch {
					case variant.Fields.IsNamed():
						namedFields := variant.Fields.GetNamed().Fields
						complexVariantFields = namedFields
						for _, variantField := range namedFields {
							structGroup.Add(idlcode.IdlFieldToCode(variantField, idlcode.FieldCodeOption{
								AsPointer:   variantField.Type.IsOption(),
								ComplexEnum: ctx.IsComplexEnumByType(&variantField.Type),
							}))
						}
					case variant.Fields.IsTuple():
						tupleFields := variant.Fields.GetTuple().Types
						for i, variantTupleItem := range tupleFields {
							variantField := idl.IdlField{
								Name: GetTupleStructElementName(i),
								Type: variantTupleItem,
							}
							complexVariantFields = append(complexVariantFields, variantField)
							structGroup.Add(idlcode.IdlFieldToCode(variantField, idlcode.FieldCodeOption{
								AsPointer:   variantField.Type.IsOption(),
								ComplexEnum: ctx.IsComplexEnumByType(&variantField.Type),
							}))
						}
					}
				},
			)
		}

		code.Line().Line()

		if variant.IsUint8Variant() {
			// Uint8 Enum' serialization is handled by the base type (uint8), leave them empty
			// Each complex enum variant will have its own MarshalWithEncoder and UnmarshalWithDecoder methods.
			// Decoder will read the enum id of the enum container, and then decode the enum variant which enum id refers to.

			// Declare MarshalWithEncoder
			code.Line().Line().Func().Params(Id("obj").Id(variantTypeNameComplex)).Id("MarshalWithEncoder").
				Params(
					Id("encoder").Op("*").Qual(model.PkgDfuseBinary, "Encoder"),
				).
				Params(
					Err().Error(),
				).
				BlockFunc(func(body *Group) {
					body.Return(Nil())
				}).Line()

			// Declare UnmarshalWithDecoder
			code.Func().Params(Id("obj").Op("*").Id(variantTypeNameComplex)).Id("UnmarshalWithDecoder").
				Params(
					Id("decoder").Op("*").Qual(model.PkgDfuseBinary, "Decoder"),
				).
				Params(
					Err().Error(),
				).
				BlockFunc(func(body *Group) {
					body.Return(Nil())
				}).Line()
		} else {
			code.Add(
				GenerateMarshalWithEncoderForStruct(
					ctx,
					variantTypeNameComplex,
					complexVariantFields,
					nil,
					true,
					program,
				),
			).Line()

			code.Add(
				GenerateUnmarshalWithDecoderForStruct(
					ctx,
					variantTypeNameComplex,
					complexVariantFields,
					nil,
					nil,
					program,
				),
			).Line()
		}

		code.Line()

		// Declare the method to implement the parent enum interface:
		if variant.IsUint8Variant() {
			code.Func().Params(Id("_").Op("*").Id(variantTypeNameComplex)).Id(interfaceMethodName).Params().Block()
		} else {
			code.Func().Params(Id("_").Op("*").Id(variantTypeNameComplex)).Id(interfaceMethodName).Params().Block()
		}

		code.Line().Line()
	}

	return code
}

func generateStructTypeDefCode(ctx *model.GenerateCtx, exportedStructName string, structDef *idl.IdlTypeDefTyStruct, anchorDiscriminator *[8]byte, program *idl.Idl) Code {
	var structFields []idl.IdlField
	code := Type().Id(exportedStructName).StructFunc(func(fieldsGroup *Group) {
		switch {
		case structDef.Fields.IsNamed():
			namedFileds := structDef.Fields.GetNamed().Fields
			structFields = namedFileds
			for fieldIndex, field := range namedFileds {
				for docIndex, doc := range field.Docs {
					if docIndex == 0 && fieldIndex > 0 {
						fieldsGroup.Line()
					}
					fieldsGroup.Comment(doc)
				}
				fieldsGroup.Add(idlcode.IdlFieldToCode(field, idlcode.FieldCodeOption{
					AsPointer:   field.Type.IsOption(),
					ComplexEnum: ctx.IsComplexEnumByType(&field.Type),
				}))
			}
		case structDef.Fields.IsTuple():
			tupleFields := structDef.Fields.GetTuple().Types
			for fieldIndex, typ := range tupleFields {
				named := idl.IdlField{
					Name: GetTupleStructElementName(fieldIndex),
					Docs: []string{
						fmt.Sprintf("Tuple struct field %d", fieldIndex),
					},
					Type: typ,
				}
				structFields = append(structFields, named)
				for docIndex, doc := range named.Docs {
					if docIndex == 0 && fieldIndex > 0 {
						fieldsGroup.Line()
					}
					fieldsGroup.Comment(doc)
				}
				fieldsGroup.Add(idlcode.IdlFieldToCode(named, idlcode.FieldCodeOption{
					AsPointer:   typ.IsOption(),
					ComplexEnum: ctx.IsComplexEnumByType(&typ),
				}))
			}
		}
	}).Line()

	if ctx.Encoder != model.EncoderTypeBorsh {
		return code
	}

	// generate encoder and decoder methods (for borsh):
	var discriminatorName *string
	var structDiscriminator *ag_binary.TypeID

	if anchorDiscriminator != nil {
		typeId := ag_binary.TypeID(*anchorDiscriminator)

		structDiscriminator = &typeId
		discriminatorName = helper.StrPtr(GetDiscriminatorName(exportedStructName))

		code.Var().Id(*discriminatorName).Op("=").Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
			for _, byteVal := range structDiscriminator[:] {
				byteGroup.Lit(int(byteVal))
			}
		}).Op("}")
	}

	code.Line().Line().Add(
		GenerateMarshalWithEncoderForStruct(
			ctx,
			exportedStructName,
			structFields,
			discriminatorName,
			true,
			program,
		),
	)
	code.Line().Line().Add(
		GenerateUnmarshalWithDecoderForStruct(
			ctx,
			exportedStructName,
			structFields,
			discriminatorName,
			structDiscriminator,
			program,
		),
	)

	return code
}
