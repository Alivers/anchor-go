package accounts

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/generator/program/common"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateAccounts(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)

	for _, acc := range program.Accounts {
		identType := ctx.GetIdentifierTy(acc.Name)
		if identType == nil {
			panic("account " + acc.Name + " not found in IDL types")
		}

		var discriminator *[8]byte
		if acc.Discriminator != nil {
			discriminator = (*[8]byte)(acc.Discriminator)
		}

		file.Add(
			common.GenerateTypeDefCode(
				ctx,
				acc.Name+"Account",
				identType,
				discriminator,
				program,
			),
		)
	}

	return file
}
