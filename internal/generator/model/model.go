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

	AddressTable                map[string]string
	IdentifierTypeRegistry      map[string]*idl.IdlTypeDefTy
	GeneratedIdentifierRegistry mapset.Set[string]
	// Rust enums which contain variant data
	// e.g. enum Foo { Bar(u8), Baz(Struct) }
	// These enums will be generated into `interface` which have various variants in Go.
	ComplexEnumRegistry mapset.Set[string]
}

func NewGenerateCtx(packageName, programName string, discriminatorType DiscriminatorType, encoder EncoderType) *GenerateCtx {
	ctx := &GenerateCtx{
		PkgName:                     packageName,
		ProgramName:                 programName,
		DiscriminatorType:           discriminatorType,
		Encoder:                     encoder,
		AddressTable:                map[string]string{},
		IdentifierTypeRegistry:      map[string]*idl.IdlTypeDefTy{},
		GeneratedIdentifierRegistry: mapset.NewSet[string](),
		ComplexEnumRegistry:         mapset.NewSet[string](),
	}

	return ctx
}

func (ctx *GenerateCtx) GetIdentifierTy(identName string) *idl.IdlTypeDefTy {
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

func (ctx *GenerateCtx) IsComplexEnumByTypeName(typeName string) bool {
	return ctx.ComplexEnumRegistry.Contains(typeName)
}

func (ctx *GenerateCtx) IsComplexEnumByType(typ *idl.IdlType) bool {
	if typ.IsDefined() {
		return ctx.IsComplexEnumByTypeName(typ.GetDefined().Name)
	}
	return false
}

func (ctx *GenerateCtx) IsGeneratedIdentifier(identName string) bool {
	return ctx.GeneratedIdentifierRegistry.Contains(identName)
}

func (ctx *GenerateCtx) AddGeneratedIdentifier(identName string) {
	ctx.GeneratedIdentifierRegistry.Add(identName)
}
