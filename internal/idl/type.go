package idl

import (
	"encoding/json"
	"errors"
	"fmt"
)

type IdlTypeSimple string

const (
	IdlTypeSimpleBool   IdlTypeSimple = "bool"
	IdlTypeSimpleU8     IdlTypeSimple = "u8"
	IdlTypeSimpleI8     IdlTypeSimple = "i8"
	IdlTypeSimpleU16    IdlTypeSimple = "u16"
	IdlTypeSimpleI16    IdlTypeSimple = "i16"
	IdlTypeSimpleU32    IdlTypeSimple = "u32"
	IdlTypeSimpleI32    IdlTypeSimple = "i32"
	IdlTypeSimpleF32    IdlTypeSimple = "f32"
	IdlTypeSimpleU64    IdlTypeSimple = "u64"
	IdlTypeSimpleI64    IdlTypeSimple = "i64"
	IdlTypeSimpleF64    IdlTypeSimple = "f64"
	IdlTypeSimpleU128   IdlTypeSimple = "u128"
	IdlTypeSimpleI128   IdlTypeSimple = "i128"
	IdlTypeSimpleU256   IdlTypeSimple = "u256"
	IdlTypeSimpleI256   IdlTypeSimple = "i256"
	IdlTypeSimpleBytes  IdlTypeSimple = "bytes"
	IdlTypeSimpleString IdlTypeSimple = "string"
	IdlTypeSimplePubkey IdlTypeSimple = "pubkey"
)

type IdlType struct {
	*IdlTypeSimple
	*IdlTypeOption
	*IdlTypeVec
	*IdlTypeArray
	*IdlTypeDefined
	*IdlTypeGeneric
	*IdlTypeHashMap
}

type IdlTypeOption struct {
	Option IdlType
}

type IdlTypeVec struct {
	Vec IdlType
}

type IdlTypeArray struct {
	Elem IdlType
	Len  IdlArrayLen
}

type IdlTypeDefined struct {
	Name     string          `json:"name"`
	Generics []IdlGenericArg `json:"generics,omitempty"`
}

type IdlTypeGeneric struct {
	Name string
}

// !!! Notice: `HashMap` is not a standard type in the IDL spec.
type IdlTypeHashMap struct {
	Key IdlType
	Val IdlType
}

type IdlArrayLen struct {
	*IdlArrayLenGeneric
	*IdlArrayLenValue
}

type IdlArrayLenGeneric struct {
	Value string
}

type IdlArrayLenValue struct {
	Value uint
}

type IdlGenericArg struct {
	*IdlGenericArgType
	*IdlGenericArgConst
}

type IdlGenericArgType struct {
	Kind string  `json:"kind"`
	Type IdlType `json:"type"`
}

type IdlGenericArgConst struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func (simple IdlTypeSimple) String() string {
	return string(simple)
}

func (idlType *IdlType) IsSimple() bool {
	return idlType.IdlTypeSimple != nil
}

func (idlType *IdlType) IsOption() bool {
	return idlType.IdlTypeOption != nil
}

func (idlType *IdlType) IsVec() bool {
	return idlType.IdlTypeVec != nil
}

func (idlType *IdlType) IsArray() bool {
	return idlType.IdlTypeArray != nil
}

func (idlType *IdlType) IsDefined() bool {
	return idlType.IdlTypeDefined != nil
}

func (idlType *IdlType) IsGeneric() bool {
	return idlType.IdlTypeGeneric != nil
}

func (idlType *IdlType) IsHashMap() bool {
	return idlType.IdlTypeHashMap != nil
}

func (idlType *IdlType) GetSimple() IdlTypeSimple {
	return *idlType.IdlTypeSimple
}

func (idlType *IdlType) GetOption() *IdlTypeOption {
	return idlType.IdlTypeOption
}

func (idlType *IdlType) GetVec() *IdlTypeVec {
	return idlType.IdlTypeVec
}

func (idlType *IdlType) GetArray() *IdlTypeArray {
	return idlType.IdlTypeArray
}

func (idlType *IdlType) GetDefined() *IdlTypeDefined {
	return idlType.IdlTypeDefined
}

func (idlType *IdlType) GetGeneric() *IdlTypeGeneric {
	return idlType.IdlTypeGeneric
}

func (idlType *IdlType) GetHashMap() *IdlTypeHashMap {
	return idlType.IdlTypeHashMap
}

func (idlType *IdlType) UnmarshalJSON(data []byte) error {
	var s IdlTypeSimple
	if err := json.Unmarshal(data, &s); err == nil {
		switch s {
		case IdlTypeSimpleBool, IdlTypeSimpleU8, IdlTypeSimpleI8,
			IdlTypeSimpleU16, IdlTypeSimpleI16, IdlTypeSimpleU32,
			IdlTypeSimpleI32, IdlTypeSimpleF32, IdlTypeSimpleU64,
			IdlTypeSimpleI64, IdlTypeSimpleF64, IdlTypeSimpleU128,
			IdlTypeSimpleI128, IdlTypeSimpleU256, IdlTypeSimpleI256,
			IdlTypeSimpleBytes, IdlTypeSimpleString, IdlTypeSimplePubkey:
			idlType.IdlTypeSimple = &s
			return nil
		default:
			return fmt.Errorf("unknown simple type: %s", s)
		}
	}

	var objMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objMap); err != nil {
		return err
	}

	// {"option": "<innerType>"}
	// {"option": {}}
	if opt, ok := objMap["option"]; ok {
		var inner IdlType
		if err := json.Unmarshal(opt, &inner); err != nil {
			return err
		}
		idlType.IdlTypeOption = &IdlTypeOption{Option: inner}
		return nil
	}

	// {"vec": "<innerType>"}
	// {"vec": {}}
	if vec, ok := objMap["vec"]; ok {
		var inner IdlType
		if err := json.Unmarshal(vec, &inner); err != nil {
			return err
		}
		idlType.IdlTypeVec = &IdlTypeVec{Vec: inner}
		return nil
	}

	// {"array": ["<elemType>", "<len>"]}
	if array, ok := objMap["array"]; ok {
		var inner []any
		if err := json.Unmarshal(array, &inner); err != nil {
			return err
		}

		if len(inner) != 2 {
			return errors.New("array type must have 2 elements")
		}

		var elemType IdlType
		if err := transcodeJSON(inner[0], &elemType); err != nil {
			return err
		}

		var arrayLen IdlArrayLen
		if err := transcodeJSON(inner[1], &arrayLen); err != nil {
			return err
		}

		idlType.IdlTypeArray = &IdlTypeArray{
			Elem: elemType,
			Len:  arrayLen,
		}
		return nil
	}

	// {"defined": {"name": "<name>", "generics": [<genericArgs>]}}
	if defined, ok := objMap["defined"]; ok {
		var inner IdlTypeDefined
		if err := json.Unmarshal(defined, &inner); err != nil {
			return err
		}
		idlType.IdlTypeDefined = &inner
		return nil
	}

	// {"generic": "<name>"}
	if generic, ok := objMap["generic"]; ok {
		var name string
		if err := json.Unmarshal(generic, &name); err != nil {
			return err
		}
		idlType.IdlTypeGeneric = &IdlTypeGeneric{Name: name}
		return nil
	}

	// {"hashMap": ["<keyType>", "<valType>"]}
	if hashMap, ok := objMap["hashMap"]; ok {
		var inner []any
		if err := json.Unmarshal(hashMap, &inner); err != nil {
			return err
		}

		if len(inner) != 2 {
			return errors.New("hashMap type must have 2 elements")
		}

		var elemType IdlType
		if err := transcodeJSON(inner[0], &elemType); err != nil {
			return err
		}

		var valType IdlType
		if err := transcodeJSON(inner[1], &valType); err != nil {
			return err
		}

		idlType.IdlTypeHashMap = &IdlTypeHashMap{
			Key: elemType,
			Val: valType,
		}
		return nil
	}

	return errors.New("unable to unmarshal IdlType")
}

func (arrayLen *IdlArrayLen) IsGeneric() bool {
	return arrayLen.IdlArrayLenGeneric != nil
}

func (arrayLen *IdlArrayLen) IsValue() bool {
	return arrayLen.IdlArrayLenValue != nil
}

func (arrayLen *IdlArrayLen) GetGeneric() *IdlArrayLenGeneric {
	return arrayLen.IdlArrayLenGeneric
}

func (arrayLen *IdlArrayLen) GetValue() *IdlArrayLenValue {
	return arrayLen.IdlArrayLenValue
}

func (arrayLen *IdlArrayLen) UnmarshalJSON(data []byte) error {
	var val uint
	if err := json.Unmarshal(data, &val); err == nil {
		arrayLen.IdlArrayLenValue = &IdlArrayLenValue{Value: val}
		return nil
	}

	var objMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objMap); err != nil {
		return err
	}

	// https://github.com/solana-foundation/anchor/blob/v0.31.1/idl/spec/src/lib.rs#L261
	// generic enum is tagged
	if generic, ok := objMap["generic"]; ok {
		var g string
		if err := json.Unmarshal(generic, &g); err != nil {
			return err
		}
		arrayLen.IdlArrayLenGeneric = &IdlArrayLenGeneric{Value: g}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		arrayLen.IdlArrayLenGeneric = &IdlArrayLenGeneric{Value: s}
		return nil
	}

	return errors.New("unable to unmarshal IdlArrayLen")
}

func (arg *IdlGenericArg) IsType() bool {
	return arg.IdlGenericArgType != nil
}

func (arg *IdlGenericArg) IsConst() bool {
	return arg.IdlGenericArgConst != nil
}

func (arg *IdlGenericArg) GetType() *IdlGenericArgType {
	return arg.IdlGenericArgType
}

func (arg *IdlGenericArg) GetConst() *IdlGenericArgConst {
	return arg.IdlGenericArgConst
}

func (arg *IdlGenericArg) UnmarshalJSON(data []byte) error {
	var objMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objMap); err != nil {
		return err
	}

	var kind string
	if kindData, ok := objMap["kind"]; ok {
		if err := json.Unmarshal(kindData, &kind); err != nil {
			return err
		}
	} else {
		return errors.New("generic arg missing kind")
	}

	switch kind {
	case "type":
		var typeWrapper IdlType
		if typeData, ok := objMap["type"]; ok {
			if err := json.Unmarshal(typeData, &typeWrapper); err != nil {
				return err
			}
		} else {
			return errors.New("type generic arg missing type")
		}

		arg.IdlGenericArgType = &IdlGenericArgType{
			Kind: "type",
			Type: typeWrapper,
		}
		return nil

	case "const":
		var value string
		if valueData, ok := objMap["value"]; ok {
			if err := json.Unmarshal(valueData, &value); err != nil {
				return err
			}
		} else {
			return errors.New("const generic arg missing value")
		}

		arg.IdlGenericArgConst = &IdlGenericArgConst{
			Kind:  "const",
			Value: value,
		}
		return nil

	default:
		return fmt.Errorf("unknown generic arg kind: %s", kind)
	}
}
