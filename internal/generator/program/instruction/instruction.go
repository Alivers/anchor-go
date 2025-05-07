package instruction

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/idlcode"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	mapset "github.com/deckarep/golang-set/v2"
)

func GenerateInstruction(ctx *model.GenerateCtx, program *idl.Idl, instruction *idl.IdlInstruction) (instName string, instExportedName string, file *File) {
	instExportedName = helper.ToCamelCase(instruction.Name)
	file = helper.NewGoFile(ctx)

	addHeaderComment(file, instruction)
	addInstructionStruct(ctx, file, instExportedName, instruction)
	addInstructionBuilder(ctx, file, instExportedName, instruction)
	addInstructionArgsSetter(ctx, file, instExportedName, instruction)
	addInstructionAccountsGetterSetter(ctx, file, instExportedName, instruction, program)
	addInstructionBuildMethod(ctx, file, instExportedName)
	addInstructionValidateMethod(file, instExportedName, instruction)
	addInstructionValidateAndBuildMethod(file, instExportedName)
	addInstructionEncodeToTreeMethod(ctx, file, instExportedName, instruction)
	addInstructionStructSerializeMethod(ctx, file, instExportedName, instruction, program)
	addInstructionConstructor(file, instExportedName, instruction)

	return instruction.Name, instExportedName, file
}

func addHeaderComment(file *File, instruction *idl.IdlInstruction) {
	file.Line()
	for _, doc := range instruction.Docs {
		file.Comment(doc).Line()
	}
}

func addInstructionStruct(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction) {
	file.Commentf("%s is the `%s` instruction.", instExportedName, instruction.Name)

	ctx.AddGeneratedIdentifier(instExportedName)

	file.Type().Id(instExportedName).StructFunc(func(fieldsGroup *Group) {
		for _, arg := range instruction.Args {
			for _, doc := range arg.Docs {
				fieldsGroup.Line().Comment(doc)
			}
			isComplexEnum := ctx.IsComplexEnum(arg.Name)
			fieldsGroup.Add(idlcode.IdlFieldToCode(arg, idlcode.FieldCodeOption{AsPointer: true, ComplexEnum: isComplexEnum}))
		}
		fieldsGroup.Line()

		instAccounts := instruction.GetAccountsWithRelation()
		prevGroupPath := ""
		for accountIdx, account := range instAccounts {
			groupPath := buildInstAccountGroupPath(account.Parents)

			comments := buildInstructionAccountComments(
				len(instAccounts),
				accountIdx,
				prevGroupPath,
				groupPath,
				account.Account,
			)
			prevGroupPath = groupPath
			fieldsGroup.Comment(comments)
		}

		fieldsGroup.Qual(model.PkgSolanaGo, "AccountMetaSlice").Tag(map[string]string{
			"bin": "-",
		}).Line()
	})
}

func addInstructionBuilder(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction) {
	builderFuncName := newInstructionBuilderName(instExportedName)
	file.Commentf("%s creates a new `%s` instruction builder.", builderFuncName, instExportedName)
	file.Func().Id(builderFuncName).Params().Op("*").Id(instExportedName).
		BlockFunc(func(body *Group) {
			instAccounts := instruction.GetAccountsWithRelation()
			body.Id("nd").Op(":=").Op("&").Id(instExportedName).Block(
				Id("AccountMetaSlice").Op(":").Make(Qual(model.PkgSolanaGo, "AccountMetaSlice"), Lit(len(instAccounts))).Op(","),
			)

			for accountIdx, accountWrapper := range instAccounts {
				account := accountWrapper.Account

				if account.Address != nil && *account.Address != "" {
					def := Qual(model.PkgSolanaGo, "Meta").Call(Id("Addresses").Index(Lit(*account.Address)))
					ctx.SetAddress(*account.Address)
					if account.Writable {
						def.Dot("WRITE").Call()
					}
					if account.Signer {
						def.Dot("SIGNER").Call()
					}
					body.Id("nd").Dot("AccountMetaSlice").Index(Lit(accountIdx)).Op("=").Add(def)
				}
			}

			body.Return(Id("nd"))
		}).Line()
}

func addInstructionArgsSetter(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction) {
	for _, arg := range instruction.Args {
		exportedArgName := helper.ToCamelCase(arg.Name)

		file.Line().Line()
		name := "Set" + exportedArgName
		file.Commentf("%s sets the %q parameter.", name, arg.Name)
		for _, doc := range arg.Docs {
			file.Comment(doc).Line()
		}

		file.Func().Params(Id("inst").Op("*").Id(instExportedName)).Id(name).
			Params(
				Id(arg.Name).Add(idlcode.IdlTypeToCode(arg.Type)),
			).
			Params(
				Op("*").Id(instExportedName),
			).
			BlockFunc(func(body *Group) {
				body.Id("inst").Dot(exportedArgName).Op("=").
					Add(func() Code {
						// Complex enum should be a interface, don't use ref `&`.
						if ctx.IsComplexEnum(arg.Name) {
							return nil
						}
						return Op("&")
					}()).
					Id(arg.Name)
				body.Return().Id("inst")
			})
	}
}

func addInstructionAccountsGetterSetter(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction, program *idl.Idl) {
	groupAccountIdx := 0
	declaredReceivers := mapset.NewSet[string]()
	var groupAccountReceiverName string

	instAccounts := instruction.GetAccountsWithRelation()
	for accountIdx, accountWrapper := range instAccounts {
		if len(accountWrapper.Parents) == 0 {
			// This is a top level account, so we can use the instruction receiver.
			// The account getter and setter will be on the instruction self.
			groupAccountIdx = accountIdx
			groupAccountReceiverName = instExportedName
		} else {
			internalGroup := buildInstAccountGroupPath(accountWrapper.Parents)
			// This is a child account, so we need to create a new receiver for the group.
			builderStructName := instAccountsBuilderStructName(instExportedName, helper.ToCamelCase(internalGroup))
			groupAccountReceiverName = builderStructName
			// 1. If the group is already declared, we just need to add the method to the group receiver.
			// 2. If the group don't exist, create it and add the method to the instruction receiver.
			if !declaredReceivers.Contains(builderStructName) {
				groupAccountIdx = 0
				declaredReceivers.Add(builderStructName)

				file.Line().Type().Id(builderStructName).Struct(
					Qual(model.PkgSolanaGo, "AccountMetaSlice").Tag(map[string]string{
						"bin": "-",
					}),
				)
				// func that returns a new builder for this account group:
				file.Line().Line().Func().Id("New" + builderStructName).Params().Op("*").Id(builderStructName).
					BlockFunc(func(gr *Group) {
						gr.Return().Op("&").Id(builderStructName).Block(
							Id("AccountMetaSlice").Op(":").Make(
								Qual(model.PkgSolanaGo, "AccountMetaSlice"),
								Lit(accountWrapper.Parents[len(accountWrapper.Parents)-1].GetAccountNum()),
							).Op(","),
						)
					}).Line().Line()

				// Method on intruction builder that accepts the accounts group builder, and copies the accounts:
				file.Line().Line().Func().Params(Id("inst").Op("*").Id(instExportedName)).Id(instAccountSetterWithBuilderName(helper.ToCamelCase(internalGroup))).
					Params(
						Id(helper.ToLowerCamelCase(builderStructName)).Op("*").Id(builderStructName),
					).
					Params(
						Op("*").Id(instExportedName),
					).
					BlockFunc(func(gr *Group) {
						tpIndex := accountIdx
						for _, subAccount := range accountWrapper.Parents[len(accountWrapper.Parents)-1].Accounts {
							if subAccount.IdlInstructionAccount != nil {
								exportedAccountName := helper.ToCamelCase(subAccount.IdlInstructionAccount.Name)

								def := Id("inst").Dot("AccountMetaSlice").Index(Lit(tpIndex)).
									Op("=").Id(helper.ToLowerCamelCase(builderStructName)).Dot(instAccountAccessorName("Get", exportedAccountName)).Call()

								gr.Add(def)
							}
							tpIndex++
						}

						gr.Return().Id("inst")
					})
			} else {
				// If the group is already declared, the index of the account is the next index of the group.
				groupAccountIdx += 1
			}
		}

		accessorsCode := generateInstAccountAccessorsCode(
			groupAccountReceiverName,
			groupAccountIdx,
			accountWrapper.Account,
		)
		file.Add(accessorsCode).Line()

		pdaDerivationCode := generateInstPdaAccountAddressDerivationCode(
			ctx,
			groupAccountReceiverName,
			accountWrapper.Account,
			instruction,
			program,
		)
		file.Add(pdaDerivationCode).Line()
	}
}

func addInstructionBuildMethod(ctx *model.GenerateCtx, file *File, instExportedName string) {
	file.Line().Line().Func().Params(Id("inst").Id(instExportedName)).Id("Build").
		Params().
		Params(
			ListFunc(func(results *Group) {
				results.Op("*").Id("Instruction")
			}),
		).
		BlockFunc(func(body *Group) {
			instEnumName := common.GetInstructionEnumName(instExportedName)
			var typeIDCode Code

			switch ctx.DiscriminatorType {
			case model.DiscriminatorTypeUvarint32:
				typeIDCode = Qual(model.PkgDfuseBinary, "TypeIDFromUvarint32").Call(Id(instEnumName))
			case model.DiscriminatorTypeUint32:
				typeIDCode = Qual(model.PkgDfuseBinary, "TypeIDFromUint32").Call(Id(instEnumName), Qual(model.PkgEncodingBinary, "LittleEndian"))
			case model.DiscriminatorTypeUint8:
				typeIDCode = Qual(model.PkgDfuseBinary, "TypeIDFromUint8").Call(Id(instEnumName))
			case model.DiscriminatorTypeAnchor:
				typeIDCode = Id(instEnumName)
			case model.DiscriminatorTypeDefault:
				typeIDCode = Id(instEnumName)
			}

			body.Return().Op("&").Id("Instruction").Values(
				Dict{
					Id("BaseVariant"): Qual(model.PkgDfuseBinary, "BaseVariant").Values(
						Dict{
							Id("TypeID"): typeIDCode,
							Id("Impl"):   Id("inst"),
						},
					),
				},
			)
		}).Line()
}

func addInstructionValidateMethod(file *File, instExportedName string, instruction *idl.IdlInstruction) {
	file.Line().Line().Func().Params(Id("inst").Op("*").Id(instExportedName)).Id("Validate").
		Params().
		Params(
			Error(),
		).
		BlockFunc(func(body *Group) {
			if len(instruction.Args) > 0 {
				body.Comment("Check whether all (required) parameters are set:")

				body.BlockFunc(func(paramVerifyBody *Group) {
					for _, arg := range instruction.Args {
						exportedArgName := helper.ToCamelCase(arg.Name)
						// Optional params can be empty.
						if arg.Type.IsOption() {
							continue
						}

						paramVerifyBody.If(Id("inst").Dot(exportedArgName).Op("==").Nil()).Block(
							Return(
								Qual("errors", "New").Call(Lit(fmt.Sprintf("%s parameter is not set", exportedArgName))),
							),
						)
					}
				})
				body.Line()
			}

			body.Comment("Check whether all (required) accounts are set:")
			body.BlockFunc(func(accountValidationBlock *Group) {
				for accountIndex, accountWrapper := range instruction.GetAccountsWithRelation() {
					account := accountWrapper.Account
					groupPath := buildInstAccountGroupPath(accountWrapper.Parents)
					exportedAccountName := helper.ToCamelCase(filepath.Join(groupPath, account.Name))
					if account.Optional {
						accountValidationBlock.Line().Commentf(
							"[%v] = %s is optional",
							accountIndex,
							exportedAccountName,
						).Line()
					} else {
						accountValidationBlock.If(Id("inst").Dot("AccountMetaSlice").Index(Lit(accountIndex)).Op("==").Nil()).Block(
							Return(Qual("errors", "New").Call(Lit(fmt.Sprintf("accounts.%s is not set", exportedAccountName)))),
						)
					}

				}
			})

			body.Return(Nil())
		})
}

func addInstructionValidateAndBuildMethod(file *File, instExportedName string) {
	file.Line().Line().
		Comment("ValidateAndBuild validates the instruction parameters and accounts;").
		Line().
		Comment("if there is a validation error, it returns the error.").
		Line().
		Comment("Otherwise, it builds and returns the instruction.").
		Line().
		Func().Params(Id("inst").Id(instExportedName)).Id("ValidateAndBuild").
		Params().
		Params(
			ListFunc(func(results *Group) {
				results.Op("*").Id("Instruction")
				results.Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.If(
				Err().Op(":=").Id("inst").Dot("Validate").Call(),
				Err().Op("!=").Nil(),
			).Block(
				Return(Nil(), Err()),
			)

			body.Return(Id("inst").Dot("Build").Call(), Nil())
		})
}

func addInstructionEncodeToTreeMethod(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction) {
	file.Line().Line().Func().Params(Id("inst").Op("*").Id(instExportedName)).Id("EncodeToTree").
		Params(
			Id("parent").Qual(model.PkgTreeout, "Branches"),
		).
		BlockFunc(func(body *Group) {
			body.Id("parent").Dot("Child").Call(Qual(model.PkgFormat, "Program").Call(Id("ProgramName"), Id("ProgramID"))).Op(".").
				Line().
				Id("ParentFunc").Parens(Func().Parens(Id("programBranch").Qual(model.PkgTreeout, "Branches")).BlockFunc(
				func(programBranchGroup *Group) {
					programBranchGroup.Id("programBranch").Dot("Child").Call(Qual(model.PkgFormat, "Instruction").Call(Lit(instExportedName))).Op(".").
						Line().
						Id("ParentFunc").Parens(Func().Parens(Id("instructionBranch").Qual(model.PkgTreeout, "Branches")).BlockFunc(
						func(instructionBranchGroup *Group) {

							instructionBranchGroup.Comment("Parameters of the instruction:")

							instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit(fmt.Sprintf("Params[len=%v]", len(instruction.Args)))).Dot("ParentFunc").Parens(Func().Parens(Id("paramsBranch").Qual(model.PkgTreeout, "Branches")).BlockFunc(func(paramsBranchGroup *Group) {
								maxLen := 0
								if len(instruction.Args) > 0 {
									maxLen = len(slices.MaxFunc(instruction.Args, func(a, b idl.IdlField) int {
										return len(a.Name) - len(b.Name)
									}).Name)
								}
								for _, arg := range instruction.Args {
									exportedArgName := helper.ToCamelCase(arg.Name)
									paramsBranchGroup.Id("paramsBranch").Dot("Child").
										Call(
											Qual(model.PkgFormat, "Param").Call(
												Lit(strings.Repeat(" ", maxLen-len(exportedArgName))+exportedArgName+helper.StrIf(arg.Type.IsOption(), " (OPT)")),
												Add(helper.CodeIf(!arg.Type.IsOption() && !ctx.IsComplexEnumByType(&arg.Type), Op("*"))).Id("inst").Dot(exportedArgName),
											),
										)
								}
							}))

							instructionBranchGroup.Comment("Accounts of the instruction:")

							instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit(fmt.Sprintf("Accounts[len=%v]", instruction.GetAccountNum()))).Dot("ParentFunc").Parens(
								Func().Parens(Id("accountsBranch").Qual(model.PkgTreeout, "Branches")).BlockFunc(func(accountsBranchGroup *Group) {

									accounts := instruction.GetAccountsWithRelation()
									maxLen := 0
									exportedAccountNames := make([]string, len(accounts))
									for accountIndex, accountWrapper := range accounts {
										cleanedName := accountWrapper.Account.Name
										if len(cleanedName) > len("account") {
											if strings.HasSuffix(cleanedName, "account") {
												cleanedName = strings.TrimSuffix(cleanedName, "account")
											} else if strings.HasSuffix(cleanedName, "Account") {
												cleanedName = strings.TrimSuffix(cleanedName, "Account")
											}
										}
										groupPath := buildInstAccountGroupPath(accountWrapper.Parents)
										exportedAccountName := filepath.Join(groupPath, cleanedName)
										exportedAccountNames[accountIndex] = exportedAccountName
										if len(exportedAccountName) > maxLen {
											maxLen = len(exportedAccountName)
										}
									}

									for accountIndex, exportedAccountName := range exportedAccountNames {
										access := Id("accountsBranch").Dot("Child").Call(Qual(model.PkgFormat, "Meta").Call(Lit(strings.Repeat(" ", maxLen-len(exportedAccountName))+exportedAccountName), Id("inst").Dot("AccountMetaSlice").Dot("Get").Call(Lit(accountIndex))))
										accountsBranchGroup.Add(access)
									}
								}))
						}))
				}))
		}).Line()
}

func addInstructionStructSerializeMethod(ctx *model.GenerateCtx, file *File, instExportedName string, instruction *idl.IdlInstruction, program *idl.Idl) {
	marshalCodes := common.GenerateMarshalWithEncoderForStruct(
		ctx,
		instExportedName,
		instruction.Args,
		nil,
		true,
		program,
	)
	file.Add(marshalCodes).Line()

	unmarshalCodes := common.GenerateUnmarshalWithDecoderForStruct(
		ctx,
		instExportedName,
		instruction.Args,
		nil,
		nil,
		program,
	)
	file.Add(unmarshalCodes).Line()
}

func addInstructionConstructor(
	file *File,
	instExportedName string,
	instruction *idl.IdlInstruction,
) {
	instAccounts := instruction.GetAccountsWithRelation()
	paramNames := mapset.NewSetWithSize[string](len(instruction.Args) + len(instAccounts))

	constructorName := common.GetInstructionConstructorName(instExportedName)
	file.Commentf("%s declares a new %s instruction with the provided parameters and accounts.", constructorName, instExportedName)
	file.Func().Id(constructorName).
		ParamsFunc(
			func(params *Group) {
				for argIndex, arg := range instruction.Args {
					paramNames.Add(arg.Name)
					paramCode := Empty()
					if argIndex == 0 {
						paramCode.Line().Comment("Parameters:")
					}
					paramCode.Line().Id(arg.Name).Add(idlcode.IdlTypeToCode(arg.Type))
					params.Add(paramCode)
				}
				for accountIndex, wrapper := range instAccounts {
					account := wrapper.Account
					accountParamName := helper.ToLowerCamelCase(account.Name)
					if len(wrapper.Parents) > 0 {
						accountParamName = helper.ToLowerCamelCase(buildInstAccountGroupPath(wrapper.Parents) + "/" + accountParamName)
					}
					if paramNames.Contains(accountParamName) {
						accountParamName += "Account"
					}

					paramCode := Empty()
					if accountIndex == 0 {
						paramCode.Line().Comment("Accounts:")
					}
					paramCode.Line().Id(accountParamName).Qual(model.PkgSolanaGo, "PublicKey")
					params.Add(paramCode)
				}
				if len(instruction.Args) > 0 || len(instAccounts) > 0 {
					params.Line()
				}
			},
		).
		Params(
			Op("*").Id(instExportedName),
		).
		BlockFunc(func(body *Group) {
			builder := body.Return().Id(newInstructionBuilderName(instExportedName)).Call()
			for _, arg := range instruction.Args {
				exportedArgName := helper.ToCamelCase(arg.Name)
				builder.Op(".").Line().Id("Set" + exportedArgName).Call(Id(arg.Name))
			}

			declaredReceivers := mapset.NewSetWithSize[string](len(instAccounts))
			for _, wrapper := range instAccounts {
				account := wrapper.Account
				accountParamName := helper.ToLowerCamelCase(account.Name)

				if len(wrapper.Parents) > 0 {
					internalGroup := buildInstAccountGroupPath(wrapper.Parents)
					builderStructName := instAccountsBuilderStructName(instExportedName, helper.ToCamelCase(internalGroup))
					if !declaredReceivers.Contains(builderStructName) {
						declaredReceivers.Add(builderStructName)
						builder.Op(".").Line().Id(instAccountSetterWithBuilderName(helper.ToCamelCase(internalGroup))).Call(
							Line().Id("New"+builderStructName).Call().CustomFunc(
								Options{Multi: false},
								func(gr *Group) {
									hasSetParam := false
									for subIndex, subAccount := range wrapper.Parents[len(wrapper.Parents)-1].Accounts {
										if subAccount.IdlInstructionAccount != nil {
											exportedAccountName := helper.ToCamelCase(subAccount.IdlInstructionAccount.Name)
											accountParamName = helper.ToLowerCamelCase(internalGroup + "/" + helper.ToLowerCamelCase(exportedAccountName))
											if paramNames.Contains(accountParamName) {
												accountParamName += "Account"
											}

											gr.Op(".").Line()
											if subIndex == 0 {
												gr.Line()
											}
											gr.Id(instAccountAccessorName("Set", exportedAccountName)).Call(Id(accountParamName))

											hasSetParam = true
										}
									}
									if hasSetParam {
										gr.Op(",").Line()
									}
								},
							),
						)
					}
				} else {
					if paramNames.Contains(accountParamName) {
						accountParamName += "Account"
					}
					builder.Op(".").Line().Id(instAccountAccessorName("Set", helper.ToCamelCase(account.Name))).Call(Id(accountParamName))
				}
			}
		})
}
