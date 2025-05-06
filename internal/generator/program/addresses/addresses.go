package addresses

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
)

func GenerateAddresses(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)

	file.Var().Id("Addresses").Op("=").Map(String()).Qual(model.PkgSolanaGo, "PublicKey").Values(DictFunc(func(dict Dict) {
		for address := range ctx.AddressTable {
			dict[Lit(address)] = Qual(model.PkgSolanaGo, "MustPublicKeyFromBase58").Call(Lit(address))
		}
	})).Line()

	return file
}
