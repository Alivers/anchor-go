package generator

import (
	"fmt"
	"os"
	"path"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/accounts"
	"github.com/alivers/anchor-go/internal/generator/program/addresses"
	"github.com/alivers/anchor-go/internal/generator/program/constants"
	"github.com/alivers/anchor-go/internal/generator/program/errors"
	"github.com/alivers/anchor-go/internal/generator/program/events"
	"github.com/alivers/anchor-go/internal/generator/program/instruction"
	"github.com/alivers/anchor-go/internal/generator/program/instructions"
	"github.com/alivers/anchor-go/internal/generator/program/tests"
	"github.com/alivers/anchor-go/internal/generator/program/types"
	"github.com/alivers/anchor-go/internal/idl"
	ag_utilz "github.com/gagliardetto/utilz"

	"github.com/dave/jennifer/jen"
	mapset "github.com/deckarep/golang-set/v2"
)

func Generate(dstFolder string, generateTests bool, program *idl.Idl) {
	pkgName := helper.ToRustSnakeCase(program.Metadata.Name)
	dstFolder = path.Join(dstFolder, pkgName)

	ctx := &model.GenerateCtx{
		PkgName:                pkgName,
		ProgramName:            program.Metadata.Name,
		DiscriminatorType:      deriveDiscriminatorType(program),
		Encoder:                model.EncoderTypeBorsh,
		AddressTable:           map[string]string{},
		IdentifierTypeRegistry: map[string]*idl.IdlTypeDefTy{},
		ComplexEnumRegistry:    mapset.NewSet[string](),
	}

	registerIdentifiers(ctx, program)
	registerComplexEnum(ctx, program)

	files := make([]*jen.File, 0, 8+2*len(program.Instructions))
	fileName := make([]string, 0, 8+2*len(program.Instructions))

	{
		file := instructions.GenerateInstructions(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "instructions.go")
	}

	if generateTests {
		file := tests.GenerateTestUtils(ctx)
		files = append(files, file)
		fileName = append(fileName, "test_utils.go")
	}

	for _, inst := range program.Instructions {
		instName, _, file := instruction.GenerateInstruction(ctx, program, &inst)

		files = append(files, file)
		fileName = append(fileName, helper.ToRustSnakeCase(instName)+".go")

		if generateTests {
			testFile := tests.GenerateTests(ctx, program, &inst)
			files = append(files, testFile)
			fileName = append(fileName, helper.ToRustSnakeCase(instName)+"_test.go")
		}
	}

	{
		file := accounts.GenerateAccounts(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "accounts.go")
	}

	{
		file := addresses.GenerateAddresses(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "addresses.go")
	}

	{
		file := events.GenerateEvents(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "events.go")
	}

	{
		file := types.GenerateTypes(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "types.go")
	}

	{
		file := constants.GenerateConstants(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "constants.go")
	}

	{
		file := errors.GenerateErrors(ctx, program)
		files = append(files, file)
		fileName = append(fileName, "errors.go")
	}

	ag_utilz.MustCreateFolderIfNotExists(dstFolder, os.ModePerm)
	for i, file := range files {
		name := fileName[i]
		err := file.Save(path.Join(dstFolder, name))
		if err != nil {
			fmt.Printf("%v", err)
		}
	}
}

func deriveDiscriminatorType(program *idl.Idl) model.DiscriminatorType {
	if len(program.Instructions) == 0 {
		return model.DiscriminatorTypeDefault
	}

	instruction := program.Instructions[0]
	if instruction.Discriminator != nil {
		return model.DiscriminatorTypeAnchor
	} else if instruction.Discriminant != nil {
		if instruction.Discriminant.Type == model.DiscriminatorTypeUint8.String() {
			return model.DiscriminatorTypeUint8
		} else if instruction.Discriminant.Type == model.DiscriminatorTypeUint32.String() {
			return model.DiscriminatorTypeUint32
		} else if instruction.Discriminant.Type == model.DiscriminatorTypeUvarint32.String() {
			return model.DiscriminatorTypeUvarint32
		} else {
			panic(fmt.Sprintf("Unsupported discriminant type(%s) in instruction(%s)", instruction.Discriminant.Type, instruction.Name))
		}
	} else {
		return model.DiscriminatorTypeDefault
	}
}

func registerIdentifiers(ctx *model.GenerateCtx, program *idl.Idl) {
	for _, typ := range program.Types {
		ctx.SetIdentifier(typ.Name, &typ.Type)
	}
	for _, accountDef := range program.Accounts {
		if accountDef.Type.IsStruct() || accountDef.Type.IsEnum() || accountDef.Type.IsType() {
			ctx.SetIdentifier(accountDef.Name, &accountDef.Type)
		}
	}
}

func registerComplexEnum(ctx *model.GenerateCtx, program *idl.Idl) {
	for _, typ := range program.Types {
		if typ.Type.IsEnum() {
			enum := typ.Type.GetEnum()
			// If it is a uint8 enum, we don't need to generate a complex enum.
			if !enum.IsUint8Enum() {
				ctx.SetComplexEnum(typ.Name)
			}
		}
	}
	for _, accountDef := range program.Accounts {
		if accountDef.Type.IsEnum() {
			enum := accountDef.Type.GetEnum()
			if !enum.IsUint8Enum() {
				ctx.SetComplexEnum(accountDef.Name)
			}
		}
	}
}
