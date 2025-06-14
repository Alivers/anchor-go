package constants

import (
	"fmt"
	"strconv"

	"github.com/alivers/anchor-go/internal/generator/helper"
	"github.com/alivers/anchor-go/internal/generator/idlcode"
	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
)

func GenerateConstants(ctx *model.GenerateCtx, program *idl.Idl) *File {
	file := helper.NewGoFile(ctx)
	for _, c := range program.Constants {
		file.Line().Var().Id(fmt.Sprintf("CONST_%s", c.Name)).Op("=")
		if !c.Type.IsSimple() {
			panic(fmt.Sprintf("unsupported constant type: %s", spew.Sdump(c)))
		}

		simpleTyp := c.Type.GetSimple()

		switch simpleTyp {
		case idl.IdlTypeSimpleString:
			v, err := strconv.Unquote(c.Value)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleBool:
			v, err := strconv.ParseBool(c.Value)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleU8:
			v, err := strconv.ParseUint(c.Value, 10, 8)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleI8:
			v, err := strconv.ParseInt(c.Value, 10, 8)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleU16:
			v, err := strconv.ParseUint(c.Value, 10, 16)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleI16:
			v, err := strconv.ParseInt(c.Value, 10, 16)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleU32:
			v, err := strconv.ParseUint(c.Value, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleI32:
			v, err := strconv.ParseInt(c.Value, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleU64:
			v, err := strconv.ParseUint(c.Value, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleI64:
			v, err := strconv.ParseInt(c.Value, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleF32:
			v, err := strconv.ParseFloat(c.Value, 32)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimpleF64:
			v, err := strconv.ParseFloat(c.Value, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse constant: %s", spew.Sdump(c)))
			}
			file.Lit(v)
		case idl.IdlTypeSimplePubkey:
			file.Qual(model.PkgSolanaGo, "MustPublicKeyFromBase58").Call(Lit(c.Value))
		case idl.IdlTypeSimpleBytes:
			values := helper.BytesStrToBytes(c.Value)
			file.Index().Byte().Values(idlcode.IdlBytesToValuesCode(values)...)
		}
	}

	return file
}
