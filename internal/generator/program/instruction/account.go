package instruction

import (
	"encoding/hex"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/idlcode"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/solana-go"
)

func generateInstAccountAccessorsCode(
	accessorReceiverName string,
	accountIndex int,
	account *idl.IdlInstructionAccount,
) Code {
	accountVarName := helper.ToLowerCamelCase(account.Name)
	accountExportedName := helper.ToCamelCase(account.Name)

	code := Line()

	setterName := instAccountAccessorName("Set", accountExportedName)
	code.Commentf("%s sets the %q account.", setterName, account.Name).Line()
	for _, doc := range account.Docs {
		code.Comment(doc).Line()
	}

	// Create account setters:
	code.Func().Params(Id("inst").Op("*").Id(accessorReceiverName)).Id(setterName).
		Params(
			ListFunc(func(params *Group) {
				params.Id(accountVarName).Qual(model.PkgSolanaGo, "PublicKey")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Op("*").Id(accessorReceiverName)
			}),
		).
		BlockFunc(func(body *Group) {
			def := Id("inst").Dot("AccountMetaSlice").Index(Lit(accountIndex)).
				Op("=").Qual(model.PkgSolanaGo, "Meta").Call(Id(accountVarName))
			if account.Writable {
				def.Dot("WRITE").Call()
			}
			if account.Signer {
				def.Dot("SIGNER").Call()
			}
			body.Add(def)

			body.Return().Id("inst")
		}).Line()

	code.Line()

	getterName := instAccountAccessorName("Get", accountExportedName)
	if account.Optional {
		code.Commentf("%s gets the %q account (optional).", getterName, account.Name).Line()
	} else {
		code.Commentf("%s gets the %q account.", getterName, account.Name).Line()
	}
	for _, doc := range account.Docs {
		code.Comment(doc).Line()
	}
	code.Func().Params(Id("inst").Op("*").Id(accessorReceiverName)).Id(getterName).
		Params(
			ListFunc(func(params *Group) {}),
		).
		Params(
			ListFunc(func(results *Group) {
				results.Op("*").Qual(model.PkgSolanaGo, "AccountMeta")
			}),
		).
		BlockFunc(func(body *Group) {
			body.Return(Id("inst").Dot("AccountMetaSlice").Dot("Get").Call(Lit(accountIndex)))
		}).Line()

	return code
}

func generateInstPdaAccountAddressDerivationCode(
	ctx *model.GenerateCtx,
	derivationReceiverName string,
	account *idl.IdlInstructionAccount,
	instruction *idl.IdlInstruction,
	program *idl.Idl,
) Code {
	programPdaSeed, pdaSeeds := resolveInstructionAccountPda(account, instruction, program)
	if programPdaSeed == nil && len(pdaSeeds) == 0 {
		return Empty()
	}

	accountExportedName := helper.ToCamelCase(account.Name)
	derivationPrivateName := instPdaAccountDerivationPrivateFuncName(accountExportedName)
	derivationExportedName := instPdaAccountDerivationExportedFuncName(accountExportedName)
	derivationWithBumpSeedName := instPdaAccountDerivationWithBumpSeedFuncName(accountExportedName)

	derivationParamsIdent, derivationParamsType := generateDerivationFuncParamsCode(programPdaSeed, pdaSeeds)

	code := Line()

	code.Func().Params(Id("inst").Op("*").Id(derivationReceiverName)).Id(derivationPrivateName).
		ParamsFunc(func(group *Group) {
			for i, ident := range derivationParamsIdent {
				group.Add(ident).Add(derivationParamsType[i])
			}
			group.Add(Id("knownBumpSeed").Uint8())
		}).
		Params(
			Id("pda").Qual(model.PkgSolanaGo, "PublicKey"),
			Id("bumpSeed").Uint8(),
			Id("err").Error(),
		).
		BlockFunc(func(body *Group) {
			body.Add(Var().Id("seeds").Index().Index().Byte())

			for _, seed := range pdaSeeds {
				if seed.SeedConst != nil {
					body.Commentf("const: 0x%s", hex.EncodeToString(seed.SeedConst))
					body.Add(Id("seeds").Op("=").Append(Id("seeds"), Index().Byte().Values(idlcode.IdlBytesToValuesCode(seed.SeedConst)...)))
				} else {
					seedRef := seed.SeedRef
					// seedTyp := seedRef.RefType
					body.Commentf("path: %s", seedRef.SeedRefPath)
					// res, err := ag_solanago.MarshalBorsh(seedRef.SeedRefName)
					body.Add(List(Id("res"), Id("err")).Op(":=").Qual(model.PkgDfuseBinary, "MarshalBorsh").Call(Id(seedRef.SeedRefName)))
					body.Add(If(Id("err").Op("!=").Nil()).Block(Return(Id("err"))))
					body.Add(Id("seeds")).Op("=").Append(Id("seeds"), Id("res"))
				}
			}

			body.Line()

			/// !!! Notice: By default, use own `ProgramID` (it is defined in the `instructions`)
			seedProgramId := Id("ProgramID")
			if programPdaSeed != nil {
				seedProgramId = Id("programID")
				if programPdaSeed.SeedConst != nil {
					address := solana.PublicKeyFromBytes(programPdaSeed.SeedConst).String()
					body.Add(Id("programID").Op(":=").Id("Addresses").Index(Lit(address)))
					ctx.SetAddress(address)
				} else {
					seedProgramRef := programPdaSeed.SeedRef
					body.Commentf("path: %s", seedProgramRef.SeedRefPath)
					body.Add(seedProgramId.Op(":=").Id(seedProgramRef.SeedRefName))
				}
			}

			body.Line().Add(
				If(Id("knownBumpSeed").Op("!=").Lit(0)).BlockFunc(func(group *Group) {
					// seeds = append(seeds, []bytes{byte(bumpSeed)})
					group.Add(Id("seeds").Op("=").Append(Id("seeds"), Index().Byte().Values(Byte().Call(Id("bumpSeed")))))
					group.Add(List(Id("pda"), Id("err")).Op("=").Add(Qual(model.PkgSolanaGo, "CreateProgramAddress").Call(Id("seeds"), seedProgramId)))
				}).
					Else().BlockFunc(func(group *Group) {
					group.Add(List(Id("pda"), Id("bumpSeed"), Id("err")).Op("=").Add(Qual(model.PkgSolanaGo, "FindProgramAddress").Call(Id("seeds"), seedProgramId)))
				}),
			)

			body.Return()
		}).Line()

	// Notice. Don't access the args and accounts of instruction self in the function block, as they may be nil.(not set by the caller)
	code.Commentf("%s calculates %s account address with given seeds and a known bump seed.", derivationWithBumpSeedName, accountExportedName).Line()
	code.Comment("pda program and seeds which refer to the instruction accounts or args should be provided as parameters.").Line()
	code.Func().Params(Id("inst").Op("*").Id(derivationReceiverName)).Id(derivationWithBumpSeedName).
		ParamsFunc(func(group *Group) {
			for i, ident := range derivationParamsIdent {
				group.Add(ident).Add(derivationParamsType[i])
			}
			group.Add(Id("bumpSeed").Uint8())
		}).
		Params(
			ListFunc(func(results *Group) {
				results.Id("pda").Qual(model.PkgSolanaGo, "PublicKey")
				results.Id("err").Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Add(List(Id("pda"), Id("_"), Id("err")).Op("=").Id("inst").Dot(derivationPrivateName).CallFunc(func(group *Group) {
				for _, ident := range derivationParamsIdent {
					group.Add(ident)
				}
				group.Add(Id("bumpSeed"))
			}))

			body.Return()
		}).Line()

	code.Func().Params(Id("inst").Op("*").Id(derivationReceiverName)).Id("Must" + derivationWithBumpSeedName).
		ParamsFunc(func(group *Group) {
			for i, ident := range derivationParamsIdent {
				group.Add(ident).Add(derivationParamsType[i])
			}
			group.Add(Id("bumpSeed").Uint8())
		}).
		Params(
			ListFunc(func(results *Group) {
				results.Id("pda").Qual(model.PkgSolanaGo, "PublicKey")
			}),
		).
		BlockFunc(func(body *Group) {
			body.Add(List(Id("pda"), Id("_"), Id("err")).Op(":=").Id("inst").Dot(derivationPrivateName).CallFunc(func(group *Group) {
				for _, ident := range derivationParamsIdent {
					group.Add(ident)
				}
				group.Add(Id("bumpSeed"))
			}))

			body.Add(If(Id("err").Op("!=").Nil()).Block(Panic(Id("err"))))

			body.Return()
		}).Line()

	code.Commentf("%s finds %s account address with given seeds.", derivationExportedName, accountExportedName).Line()
	code.Func().Params(Id("inst").Op("*").Id(derivationReceiverName)).Id(derivationExportedName).
		ParamsFunc(func(group *Group) {
			for i, ident := range derivationParamsIdent {
				group.Add(ident).Add(derivationParamsType[i])
			}
		}).
		Params(
			ListFunc(func(results *Group) {
				results.Id("pda").Qual(model.PkgSolanaGo, "PublicKey")
				results.Id("bumpSeed").Uint8()
				results.Id("err").Error()
			}),
		).
		BlockFunc(func(body *Group) {
			body.Add(List(Id("pda"), Id("bumpSeed"), Id("err")).Op("=").Id("inst").Dot(derivationPrivateName).CallFunc(func(group *Group) {
				for _, ident := range derivationParamsIdent {
					group.Add(ident)
				}
				// Bump seed is 0 by default (let it finds in [0-255] range)
				group.Add(Lit(0))
			}))

			body.Return()
		}).Line()

	code.Func().Params(Id("inst").Op("*").Id(derivationReceiverName)).Id("Must" + derivationExportedName).
		ParamsFunc(func(group *Group) {
			for i, ident := range derivationParamsIdent {
				group.Add(ident).Add(derivationParamsType[i])
			}
		}).
		Params(
			ListFunc(func(results *Group) {
				results.Id("pda").Qual(model.PkgSolanaGo, "PublicKey")
			}),
		).
		BlockFunc(func(body *Group) {
			body.Add(List(Id("pda"), Id("_"), Id("err")).Op(":=").Id("inst").Dot(derivationPrivateName).CallFunc(func(group *Group) {
				for _, ident := range derivationParamsIdent {
					group.Add(ident)
				}
				// Bump seed is 0 by default (let it finds in [0-255] range)
				group.Add(Lit(0))
			}))

			body.Add(If(Id("err").Op("!=").Nil()).Block(Panic(Id("err"))))

			body.Return()
		}).Line()

	return code
}

func generateDerivationFuncParamsCode(programPdaSeed *pdaSeedValue, pdaSeeds []*pdaSeedValue) (identList []Code, typeList []Code) {
	if programPdaSeed == nil && len(pdaSeeds) == 0 {
		return nil, nil
	}

	identList = make([]Code, 0, len(pdaSeeds)+1)
	typeList = make([]Code, 0, len(pdaSeeds)+1)

	for _, seed := range pdaSeeds {
		// If the seed is a constant seed, we don't need to pass it as a parameter
		if seed.SeedConst != nil {
			continue
		}

		ref := seed.SeedRef
		identList = append(identList, Id(ref.SeedRefName))
		typeList = append(typeList, idlcode.IdlTypeToCode(*ref.RefType))
	}

	if programPdaSeed != nil && programPdaSeed.SeedConst == nil {
		// Only add the program ID if it's not a constant seed
		// If it is a constant seed, we don't need to pass it as a parameter
		// !!! Notice: programID must be `PublicKey` type
		typ := idl.IdlTypeSimplePubkey
		identList = append(identList, Id(programPdaSeed.SeedRef.SeedRefName))
		typeList = append(typeList, idlcode.IdlTypeSimpleToCode(typ))
	}

	return identList, typeList
}
