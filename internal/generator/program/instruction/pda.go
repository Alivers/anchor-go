package instruction

import (
	"fmt"
	"strings"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/idl"
	ag_solanago "github.com/gagliardetto/solana-go"
)

func resolveInstructionAccountPda(account *idl.IdlInstructionAccount, instruction *idl.IdlInstruction, program *idl.Idl) (pdaProgram *pdaSeedValue, pdaSeeds []*pdaSeedValue) {
	if account.Pda == nil {
		return nil, nil
	}

	instAccounts := instruction.GetAccounts()

	if account.Pda.Program != nil {
		pdaProgram = resolveInstructionAccountSeed(account.Pda.Program, instAccounts, instruction, program)
	}

	if len(account.Pda.Seeds) > 0 {
		pdaSeeds = make([]*pdaSeedValue, len(account.Pda.Seeds))
		for i, seed := range account.Pda.Seeds {
			pdaSeeds[i] = resolveInstructionAccountSeed(&seed, instAccounts, instruction, program)
		}
	}

	return pdaProgram, pdaSeeds
}

func resolveInstructionAccountSeed(seed *idl.IdlSeed, instAccounts []*idl.IdlInstructionAccount, instruction *idl.IdlInstruction, program *idl.Idl) *pdaSeedValue {
	switch {
	case seed.IsConst():
		constSeed := seed.GetConst()
		if constSeed.Value == nil {
			panic(fmt.Sprintf("[IdlSeedConst]Invalid seed value for const: %v", constSeed))
		}
		return &pdaSeedValue{
			OriginIdlSeed: seed,
			SeedConst:     constSeed.Value,
			SeedRef:       nil,
		}
	case seed.IsArg():
		argSeed := seed.GetArg()
		argField := findInstArgByName(argSeed.Path, instruction.Args)
		if argField == nil {
			argParts := strings.Split(argSeed.Path, ".")
			if len(argParts) != 2 {
				panic(fmt.Sprintf("[IdlSeedArg]Invalid argument path to split by accesor: %v", argSeed.Path))
			}
			argField = findInstArgByName(argParts[0], instruction.Args)
			if argField == nil {
				panic(fmt.Sprintf("[IdlSeedArg]Argument field not found for path: %v, find by %s", argSeed.Path, argParts[0]))
			}
			switch {
			case argField.Type.IsDefined():
				definedType := argField.Type.GetDefined()
				argType := findStructFieldTypeInProgramTypes(definedType.Name, argParts[1], program.Types)
				if argType == nil {
					panic(fmt.Sprintf("[IdlSeedArg]Type definition not found in program types for: %s.%s", definedType.Name, argParts[1]))
				}
				return &pdaSeedValue{
					OriginIdlSeed: seed,
					SeedConst:     nil,
					SeedRef: &pdaSeedRef{
						SeedRefPath: argSeed.Path,
						SeedRefName: strings.Join(argParts, "_"),
						RefType:     argType,
					},
				}
			}
		} else {
			return &pdaSeedValue{
				OriginIdlSeed: seed,
				SeedConst:     nil,
				SeedRef: &pdaSeedRef{
					SeedRefPath: argSeed.Path,
					SeedRefName: helper.ToLowerCamelCase(argSeed.Path),
					RefType:     &argField.Type,
				},
			}
		}
	case seed.IsAccount():
		accountSeed := seed.GetAccount()
		if accountSeed.Account == nil {
			for _, account := range instAccounts {
				if account.Name == accountSeed.Path {
					if account.Address != nil && *account.Address != "" {
						return &pdaSeedValue{
							OriginIdlSeed: seed,
							SeedConst:     ag_solanago.MustPublicKeyFromBase58(*account.Address).Bytes(),
							SeedRef:       nil,
						}
					} else {
						ty := idl.IdlTypeSimplePubkey
						return &pdaSeedValue{
							OriginIdlSeed: seed,
							SeedConst:     nil,
							SeedRef: &pdaSeedRef{
								SeedRefPath: accountSeed.Path,
								SeedRefName: helper.ToLowerCamelCase(account.Name),
								RefType:     &idl.IdlType{IdlTypeSimple: &ty},
							},
						}
					}
				}
			}
		} else {
			fieldParts := strings.Split(accountSeed.Path, ".")
			if len(fieldParts) != 2 {
				panic(fmt.Sprintf("[IdlSeedAccount]Invalid account path to split by accesor: %v", accountSeed.Path))
			}
			fieldType := findStructFieldTypeInProgramTypes(*accountSeed.Account, fieldParts[1], program.Types)
			if fieldType == nil {
				panic(fmt.Sprintf("[IdlSeedAccount]Type definition not found in program types for: %s.%s", fieldParts[0], fieldParts[1]))
			}
			return &pdaSeedValue{
				OriginIdlSeed: seed,
				SeedConst:     nil,
				SeedRef: &pdaSeedRef{
					SeedRefPath: accountSeed.Path,
					SeedRefName: helper.ToLowerCamelCase(strings.Join(fieldParts, "_")),
					RefType:     fieldType,
				},
			}
		}
	}

	return nil
}

func findInstArgByName(argName string, args []idl.IdlField) *idl.IdlField {
	for _, arg := range args {
		if arg.Name == argName {
			return &arg
		}
	}
	return nil
}

func findStructFieldTypeInProgramTypes(structTypeName, fieldName string, programTypes []idl.IdlTypeDef) *idl.IdlType {
	for _, typ := range programTypes {
		// Match the type name
		if typ.Name == structTypeName {
			switch {
			case typ.Type.IsStruct():
				structType := typ.Type.GetStruct()
				switch {
				case structType.Fields.IsNamed():
					for _, field := range structType.Fields.GetNamed().Fields {
						// Match the field name
						if field.Name == fieldName {
							return &field.Type
						}
					}
				case structType.Fields.IsTuple():
					for i, typ := range structType.Fields.GetTuple().Types {
						// Match the field name(tuple index)
						if helper.IntToStr(i) == fieldName {
							return &typ
						}
					}
				}
			}
		}
	}
	return nil
}
