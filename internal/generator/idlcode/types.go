package idlcode

import (
	"strconv"

	"github.com/alivers/anchor-go/internal/generator/model"
	"github.com/alivers/anchor-go/internal/idl"
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
)

func IdlTypeSimpleToCode(typ idl.IdlTypeSimple) Code {
	switch typ {
	case idl.IdlTypeSimpleBool:
		return Bool()
	case idl.IdlTypeSimpleU8:
		return Uint8()
	case idl.IdlTypeSimpleI8:
		return Int8()
	case idl.IdlTypeSimpleU16:
		return Uint16()
	case idl.IdlTypeSimpleI16:
		return Int16()
	case idl.IdlTypeSimpleU32:
		return Uint32()
	case idl.IdlTypeSimpleI32:
		return Int32()
	case idl.IdlTypeSimpleF32:
		return Float32()
	case idl.IdlTypeSimpleU64:
		return Uint64()
	case idl.IdlTypeSimpleI64:
		return Int64()
	case idl.IdlTypeSimpleF64:
		return Float64()
	case idl.IdlTypeSimpleU128:
		return Qual(model.PkgDfuseBinary, "Uint128")
	case idl.IdlTypeSimpleI128:
		return Qual(model.PkgDfuseBinary, "Int128")
	case idl.IdlTypeSimpleU256:
		panic("Uint256 is not supported yet")
	case idl.IdlTypeSimpleI256:
		panic("Int256 is not supported yet")
	case idl.IdlTypeSimpleBytes:
		return Index().Byte()
	case idl.IdlTypeSimpleString:
		return String()
	case idl.IdlTypeSimplePubkey:
		return Qual(model.PkgSolanaGo, "PublicKey")
	default:
		panic("unknown type: " + typ)
	}
}

func IdlTypeToCode(typ idl.IdlType) Code {
	code := Empty()
	switch {
	case typ.IsSimple():
		code.Add(IdlTypeSimpleToCode(typ.GetSimple()))
	case typ.IsOption():
		opt := typ.GetOption()
		code.Add(IdlTypeToCode(opt.Option))
	case typ.IsVec():
		vec := typ.GetVec()
		code.Index().Add(IdlTypeToCode(vec.Vec))
	case typ.IsArray():
		arr := typ.GetArray()
		switch {
		case arr.Len.IsValue():
			len := strconv.Itoa(int(arr.Len.GetValue().Value))
			code.Index(Id(len)).Add(IdlTypeToCode(arr.Elem))
		case arr.Len.IsGeneric():
		}
	case typ.IsDefined():
		/// !!! Notice: Generic params are not supported yet
		code.Add(Id(typ.GetDefined().Name))
	case typ.IsGeneric():
		/// !!! Notice: Generic type are not supported yet
		_ = typ.GetGeneric()
	case typ.IsHashMap():
		hashMap := typ.GetHashMap()
		code.Map(IdlTypeToCode(hashMap.Key)).Add(IdlTypeToCode(hashMap.Val))
	default:
		panic("unknown type: " + spew.Sdump(typ))
	}

	return code
}

func IdlBytesToValuesCode(bytes []byte) []Code {
	code := make([]Code, 0, len(bytes))
	for _, b := range bytes {
		code = append(code, LitByte(b))
	}
	return code
}
