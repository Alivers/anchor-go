package instructions

import (
	"fmt"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	ag_binary "github.com/gagliardetto/binary"
)

func GenerateInstructions(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)
	addHeaderComment(file, program)
	addProgramId(ctx, file, program)
	addInit(file)
	addInstructionEnum(ctx, file, program)
	addInstructionIdToName(ctx, file, program)
	addInstructionVariants(ctx, file, program)
	addDecoderRegistry(file)
	addDecodeFunction(ctx, file)
	addDecodeInstructionsFunc(file)

	return file
}

func addHeaderComment(file *File, program *idl.Idl) {
	file.HeaderComment(fmt.Sprintf("Program: %s", program.Metadata.Name))
	if program.Metadata.Version != "" {
		file.HeaderComment(fmt.Sprintf("Version: %s", program.Metadata.Version))
	}
	if program.Metadata.Spec != "" {
		file.HeaderComment(fmt.Sprintf("Spec: %s", program.Metadata.Spec))
	}
	if program.Metadata.Description != nil {
		file.HeaderComment(fmt.Sprintf("Description: %s", *program.Metadata.Description))
	}
	if program.Metadata.Repository != nil {
		file.HeaderComment(fmt.Sprintf("Repository: %s", *program.Metadata.Repository))
	}
	for _, doc := range program.Docs {
		file.HeaderComment(doc)
	}
}

func addProgramId(ctx *model.GenerateCtx, file *File, program *idl.Idl) {
	programId := program.Address
	if programId == "" && program.Metadata.Address != nil {
		programId = *program.Metadata.Address
	}

	if programId == "" {
		file.Comment("Program ID is not set. Please set it manually.")
		file.Var().Id("ProgramID").Qual(model.PkgSolanaGo, "PublicKey").Line()
	} else {
		file.Var().Id("ProgramID").Qual(model.PkgSolanaGo, "PublicKey").
			Op("=").Qual(model.PkgSolanaGo, "MustPublicKeyFromBase58").Call(Lit(programId)).Line()

	}

	file.Func().Id("SetProgramID").Params(Id("pubkey").Qual(model.PkgSolanaGo, "PublicKey")).Block(
		Id("ProgramID").Op("=").Id("pubkey"),
		Qual(model.PkgSolanaGo, "RegisterInstructionDecoder").Call(Id("ProgramID"), Id("registryDecodeInstruction")),
	).Line()

	file.Const().Id("ProgramName").Op("=").Lit(ctx.ProgramName).Line()
}

func addInit(file *File) {
	file.Func().Id("init").Call().Block(
		If(
			Op("!").Id("ProgramID").Dot("IsZero").Call(),
		).Block(
			Qual(model.PkgSolanaGo, "RegisterInstructionDecoder").Call(Id("ProgramID"), Id("registryDecodeInstruction")),
		),
	).Line()
}

func addInstructionEnum(ctx *model.GenerateCtx, file *File, program *idl.Idl) {
	code := Empty()
	for _, instruction := range program.Instructions {
		insExportedName := helper.ToCamelCase(instruction.Name)

		ins := Empty()
		for _, doc := range instruction.Docs {
			ins.Comment(doc).Line()
		}
		ins.Id(common.GetInstructionEnumName(insExportedName))

		switch ctx.DiscriminatorType {
		case model.DiscriminatorTypeUint8:
			ins.Uint8().Op("=").Lit(instruction.Discriminant.Value).Line()
		case model.DiscriminatorTypeUvarint32, model.DiscriminatorTypeUint32:
			ins.Uint32().Op("=").Lit(instruction.Discriminant.Value).Line()
		case model.DiscriminatorTypeAnchor:
			ins.Op("=").Qual(model.PkgDfuseBinary, "TypeID").Call(
				Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
					for _, byteVal := range instruction.Discriminator[:] {
						byteGroup.Lit(int(byteVal))
					}
				}).Op("}"),
			)
		case model.DiscriminatorTypeDefault:
			ins.Op("=").Qual(model.PkgDfuseBinary, "TypeID").Call(
				Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
					instName := helper.ToRustSnakeCase(instruction.Name)
					sighash := ag_binary.SighashTypeID(ag_binary.SIGHASH_GLOBAL_NAMESPACE, instName)
					for _, byteVal := range sighash[:] {
						byteGroup.Lit(int(byteVal))
					}
				}).Op("}"),
			)
		}
		code.Add(ins.Line())
	}
	file.Var().Parens(code).Line()
}

func addInstructionIdToName(ctx *model.GenerateCtx, file *File, program *idl.Idl) {
	idCode := Empty()
	switch ctx.DiscriminatorType {
	case model.DiscriminatorTypeUvarint32, model.DiscriminatorTypeUint32:
		idCode.Id("id").Uint32()
	case model.DiscriminatorTypeUint8:
		idCode.Id("id").Uint8()
	case model.DiscriminatorTypeAnchor:
		idCode.Id("id").Qual(model.PkgDfuseBinary, "TypeID")
	case model.DiscriminatorTypeDefault:
		idCode.Id("id").Qual(model.PkgDfuseBinary, "TypeID")
	}
	file.Comment("InstructionIDToName returns the name of the instruction given its ID.").Line()
	file.Func().Id("InstructionIDToName").
		Params(idCode).
		Params(String()).
		BlockFunc(func(body *Group) {
			body.Switch(Id("id")).BlockFunc(func(switchBlock *Group) {
				for _, instruction := range program.Instructions {
					insExportedName := helper.ToCamelCase(instruction.Name)
					switchBlock.Case(Id("Instruction_" + insExportedName)).Line().Return(Lit(insExportedName))
				}
				switchBlock.Default().Line().Return(Lit(""))
			})
		}).Line()
}

func addInstructionVariants(ctx *model.GenerateCtx, file *File, program *idl.Idl) {
	file.Type().Id("Instruction").Struct(
		Qual(model.PkgDfuseBinary, "BaseVariant"),
	).Line()

	file.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("EncodeToTree").
		Params(Id("parent").Qual(model.PkgTreeout, "Branches")).
		Params().
		BlockFunc(func(body *Group) {
			body.If(
				List(Id("enToTree"), Id("ok")).Op(":=").Id("inst").Dot("Impl").Op(".").Parens(Qual(model.PkgSolanaGoText, "EncodableToTree")).
					Op(";").
					Id("ok"),
			).Block(
				Id("enToTree").Dot("EncodeToTree").Call(Id("parent")),
			).Else().Block(
				Id("parent").Dot("Child").Call(Qual(model.PkgSpew, "Sdump").Call(Id("inst"))),
			)
		}).Line()

	implDefParam := Line()
	switch ctx.DiscriminatorType {
	case model.DiscriminatorTypeUvarint32:
		implDefParam.Qual(model.PkgDfuseBinary, "Uvarint32TypeIDEncoding").Op(",").Line()
	case model.DiscriminatorTypeUint32:
		implDefParam.Qual(model.PkgDfuseBinary, "Uint32TypeIDEncoding").Op(",").Line()
	case model.DiscriminatorTypeUint8:
		implDefParam.Qual(model.PkgDfuseBinary, "Uint8TypeIDEncoding").Op(",").Line()
	case model.DiscriminatorTypeAnchor:
		implDefParam.Qual(model.PkgDfuseBinary, "AnchorTypeIDEncoding").Op(",").Line()
	case model.DiscriminatorTypeDefault:
		implDefParam.Qual(model.PkgDfuseBinary, "AnchorTypeIDEncoding").Op(",").Line()
	}

	file.Var().Id("InstructionImplDef").Op("=").Qual(model.PkgDfuseBinary, "NewVariantDefinition").
		Parens(
			implDefParam.Index().Qual(model.PkgDfuseBinary, "VariantType").
				BlockFunc(func(variantBlock *Group) {
					for _, instruction := range program.Instructions {
						insName := helper.ToCamelCase(instruction.Name)
						insExportedName := helper.ToCamelCase(instruction.Name)
						variantBlock.Block(
							List(Id("Name").Op(":").Lit(insName), Id("Type").Op(":").Parens(Op("*").Id(insExportedName)).Parens(Nil())).Op(","),
						).Op(",")
					}
				}).Op(",").Line(),
		).Line()

	file.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("ProgramID").Params().
		Parens(Qual(model.PkgSolanaGo, "PublicKey")).
		BlockFunc(func(body *Group) {
			body.Return(
				Id("ProgramID"),
			)
		}).Line()

	file.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("Accounts").Params().
		Parens(Id("out").Index().Op("*").Qual(model.PkgSolanaGo, "AccountMeta")).
		BlockFunc(func(body *Group) {
			body.Return(
				Id("inst").Dot("Impl").Op(".").Parens(Qual(model.PkgSolanaGo, "AccountsGettable")).Dot("GetAccounts").Call(),
			)
		}).Line()

	file.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("Data").
		Params().
		Params(
			ListFunc(func(results *Group) {
				results.Index().Byte()
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Id("buf").Op(":=").New(Qual(model.PkgBytes, "Buffer"))
			body.If(
				Err().Op(":=").Qual(model.PkgDfuseBinary, ctx.Encoder.GetNewEncoderName()).Call(Id("buf")).Dot("Encode").Call(Id("inst")).
					Op(";").
					Err().Op("!=").Nil(),
			).Block(
				Return(List(Nil(), Qual(model.PkgFmt, "Errorf").Call(Lit("unable to encode instruction: %w"), Err()))),
			)
			body.Return(Id("buf").Dot("Bytes").Call(), Nil())
		})

	file.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("TextEncode").
		Params(
			ListFunc(func(params *Group) {
				params.Id("encoder").Op("*").Qual(model.PkgSolanaGoText, "Encoder")
				params.Id("option").Op("*").Qual(model.PkgSolanaGoText, "Option")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Return(Id("encoder").Dot("Encode").Call(Id("inst").Dot("Impl"), Id("option")))
		})

	file.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("UnmarshalWithDecoder").
		Params(
			ListFunc(func(params *Group) {
				params.Id("decoder").Op("*").Qual(model.PkgDfuseBinary, "Decoder")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Return(Id("inst").Dot("BaseVariant").Dot("UnmarshalBinaryVariant").Call(Id("decoder"), Id("InstructionImplDef")))
		})

	file.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("MarshalWithEncoder").
		Params(
			ListFunc(func(params *Group) {
				params.Id("encoder").Op("*").Qual(model.PkgDfuseBinary, "Encoder")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			switch ctx.DiscriminatorType {
			case model.DiscriminatorTypeUvarint32:
				body.Err().Op(":=").Id("encoder").Dot("WriteUVarInt").Call(Id("inst").Dot("TypeID").Dot("Uvarint32").Call())
			case model.DiscriminatorTypeUint32:
				body.Err().Op(":=").Id("encoder").Dot("WriteUint32").Call(Id("inst").Dot("TypeID").Dot("Uint32").Call(), Qual(model.PkgEncodingBinary, "LittleEndian"))
			case model.DiscriminatorTypeUint8:
				body.Err().Op(":=").Id("encoder").Dot("WriteUint8").Call(Id("inst").Dot("TypeID").Dot("Uint8").Call())
			case model.DiscriminatorTypeAnchor:
				body.Err().Op(":=").Id("encoder").Dot("WriteBytes").Call(Id("inst").Dot("TypeID").Dot("Bytes").Call(), False())
			case model.DiscriminatorTypeDefault:
				body.Err().Op(":=").Id("encoder").Dot("WriteBytes").Call(Id("inst").Dot("TypeID").Dot("Bytes").Call(), False())
			}

			body.If(
				Err().Op("!=").Nil(),
			).Block(
				Return(Qual(model.PkgFmt, "Errorf").Call(Lit("unable to write variant type: %w"), Err())),
			)
			body.Return(Id("encoder").Dot("Encode").Call(Id("inst").Dot("Impl")))
		})
}

func addDecoderRegistry(file *File) {
	file.Func().Id("registryDecodeInstruction").
		Params(
			ListFunc(func(params *Group) {
				params.Id("accounts").Index().Op("*").Qual(model.PkgSolanaGo, "AccountMeta")
				params.Id("data").Index().Byte()
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Any()
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.List(Id("inst"), Err()).Op(":=").Id("DecodeInstruction").Call(Id("accounts"), Id("data"))

			body.If(
				Err().Op("!=").Nil(),
			).Block(
				Return(Nil(), Err()),
			)
			body.Return(Id("inst"), Nil())
		})
}

func addDecodeFunction(ctx *model.GenerateCtx, file *File) {
	file.Func().Id("DecodeInstruction").
		Params(
			ListFunc(func(params *Group) {
				params.Id("accounts").Index().Op("*").Qual(model.PkgSolanaGo, "AccountMeta")
				params.Id("data").Index().Byte()
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Op("*").Id("Instruction")
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Id("inst").Op(":=").New(Id("Instruction"))
			body.If(
				Err().Op(":=").Qual(model.PkgDfuseBinary, ctx.Encoder.GetNewDecoderName()).Call(Id("data")).Dot("Decode").Call(Id("inst")).
					Op(";").
					Err().Op("!=").Nil(),
			).Block(
				Return(
					Nil(),
					Qual(model.PkgFmt, "Errorf").Call(Lit("unable to decode instruction: %w"), Err()),
				),
			)

			body.If(
				List(Id("v"), Id("ok")).Op(":=").Id("inst").Dot("Impl").Op(".").Parens(Qual(model.PkgSolanaGo, "AccountsSettable")).
					Op(";").
					Id("ok"),
			).BlockFunc(func(gr *Group) {
				gr.Err().Op(":=").Id("v").Dot("SetAccounts").Call(Id("accounts"))
				gr.If(Err().Op("!=").Nil()).Block(
					Return(
						Nil(),
						Qual(model.PkgFmt, "Errorf").Call(Lit("unable to set accounts for instruction: %w"), Err()),
					),
				)
			})

			body.Return(Id("inst"), Nil())
		})
}

func addDecodeInstructionsFunc(file *File) {
	file.Func().Id("DecodeInstructions").Params(
		Id("message").Op("*").Qual(model.PkgSolanaGo, "Message"),
	).Params(
		Id("instructions").Index().Op("*").Id("Instruction"),
		Err().Error(),
	).Block(
		For(List(Id("_"), Id("ins")).Op(":=").Range().Id("message").Dot("Instructions")).Block(
			Var().Id("programID").Qual(model.PkgSolanaGo, "PublicKey"),
			If(
				List(Id("programID"), Err()).Op("=").Id("message").Dot("Program").Call(Id("ins").Dot("ProgramIDIndex")),
				Err().Op("!=").Nil(),
			).Block(
				Return(),
			),
			If(Op("!").Id("programID").Dot("Equals").Call(Id("ProgramID"))).Block(
				Continue(),
			),
			Var().Id("accounts").Index().Op("*").Qual(model.PkgSolanaGo, "AccountMeta"),
			If(
				List(Id("accounts"), Err()).Op("=").Id("ins").Dot("ResolveInstructionAccounts").Call(Id("message")),
				Err().Op("!=").Nil(),
			).Block(
				Return(),
			),
			Var().Id("insDecoded").Op("*").Id("Instruction"),
			If(
				List(Id("insDecoded"), Err()).Op("=").Id("DecodeInstruction").Call(Id("accounts"), Id("ins").Dot("Data")),
				Err().Op("!=").Nil(),
			).Block(
				Return(),
			),
			Id("instructions").Op("=").Append(Id("instructions"), Id("insDecoded")),
		),
		Return(),
	)
}
