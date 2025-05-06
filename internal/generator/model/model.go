package model

import (
	"github.com/alivers/anchor-go/internal/idl"
	mapset "github.com/deckarep/golang-set/v2"
)

type GenerateCtx struct {
	PkgName           string
	ProgramName       string
	DiscriminatorType DiscriminatorType
	Encoder           EncoderType

	AddressTable           map[string]string
	IdentifierTypeRegistry map[string]*idl.IdlTypeDefTy
	// Rust enums which contain variant data
	// e.g. enum Foo { Bar(u8), Baz(Struct) }
	// These enums will be generated into `interface` which have various variants in Go.
	ComplexEnumRegistry mapset.Set[string]
}

func (ctx *GenerateCtx) GetIdentifier(identName string) *idl.IdlTypeDefTy {
	if ty, ok := ctx.IdentifierTypeRegistry[identName]; ok {
		return ty
	}
	return nil
}

func (ctx *GenerateCtx) SetAddress(address string) {
	ctx.AddressTable[address] = address
}

func (ctx *GenerateCtx) SetIdentifier(name string, typ *idl.IdlTypeDefTy) {
	ctx.IdentifierTypeRegistry[name] = typ
}

func (ctx *GenerateCtx) SetComplexEnum(name string) {
	ctx.ComplexEnumRegistry.Add(name)
}

func (ctx *GenerateCtx) IsComplexEnum(name string) bool {
	return ctx.ComplexEnumRegistry.Contains(name)
}

func (ctx *GenerateCtx) IsComplexEnumByType(typ *idl.IdlType) bool {
	if typ.IsDefined() {
		return ctx.IsComplexEnum(typ.GetDefined().Name)
	}
	return false
}
