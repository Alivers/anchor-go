package idl

import (
	"encoding/json"
	"errors"
	"fmt"
)

type IdlField struct {
	Name string   `json:"name"`
	Docs []string `json:"docs,omitempty"`
	Type IdlType  `json:"type"`
}

type IdlTypeDef struct {
	Name          string              `json:"name"`
	Docs          []string            `json:"docs,omitempty"`
	Serialization IdlSerialization    `json:"serialization,omitempty"`
	Repr          *IdlRepr            `json:"repr,omitempty"`
	Generics      []IdlTypeDefGeneric `json:"generics,omitempty"`
	Type          IdlTypeDefTy        `json:"type"`
}

type IdlSerialization string

const (
	IdlSerializationBorsh          IdlSerialization = "borsh"
	IdlSerializationBytemuck       IdlSerialization = "bytemuck"
	IdlSerializationBytemuckUnsafe IdlSerialization = "bytemuckunsafe"
	IdlSerializationCustom         IdlSerialization = "custom"
)

// ===== IdlTypeDefGeneric Types =====
type IdlTypeDefGeneric struct {
	*IdlTypeDefGenericType
	*IdlTypeDefGenericConst
}

type IdlTypeDefGenericType struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type IdlTypeDefGenericConst struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ===== IdlTypeDefTy Types =====
type IdlTypeDefTy struct {
	*IdlTypeDefTyStruct
	*IdlTypeDefTyEnum
	*IdlTypeDefTyType
}

type IdlTypeDefTyStruct struct {
	Kind   string            `json:"kind"`
	Fields *IdlDefinedFields `json:"fields,omitempty"`
}

type IdlTypeDefTyEnum struct {
	Kind     string           `json:"kind"`
	Variants []IdlEnumVariant `json:"variants"`
}

type IdlTypeDefTyType struct {
	Kind  string  `json:"kind"`
	Alias IdlType `json:"alias"`
}

type IdlEnumVariant struct {
	Name   string            `json:"name"`
	Fields *IdlDefinedFields `json:"fields,omitempty"`
}

// ===== IdlDefinedFields Types =====
type IdlDefinedFields struct {
	*IdlDefinedFieldsNamed
	*IdlDefinedFieldsTuple
}

type IdlDefinedFieldsNamed struct {
	Fields []IdlField
}

type IdlDefinedFieldsTuple struct {
	Types []IdlType
}

// ===== IdlRepr Types =====
type IdlRepr struct {
	*IdlReprRust
	*IdlReprC
	*IdlReprTransparent
}

type IdlReprRust struct {
	Kind     string          `json:"kind"`
	Modifier IdlReprModifier `json:"modifier,omitempty"`
}

type IdlReprC struct {
	Kind     string          `json:"kind"`
	Modifier IdlReprModifier `json:"modifier,omitempty"`
}

type IdlReprTransparent struct {
	Kind string `json:"kind"`
}

type IdlReprModifier struct {
	Packed bool  `json:"packed,omitempty"`
	Align  *uint `json:"align,omitempty"`
}

func (def *IdlTypeDefGeneric) IsType() bool {
	return def.IdlTypeDefGenericType != nil
}

func (def *IdlTypeDefGeneric) IsConst() bool {
	return def.IdlTypeDefGenericConst != nil
}

func (def *IdlTypeDefGeneric) GetType() *IdlTypeDefGenericType {
	return def.IdlTypeDefGenericType
}

func (def *IdlTypeDefGeneric) GetConst() *IdlTypeDefGenericConst {
	return def.IdlTypeDefGenericConst
}

func (def *IdlTypeDefGeneric) UnmarshalJSON(data []byte) error {
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
		return errors.New("typedef generic missing kind")
	}

	switch kind {
	case "type":
		var name string
		if nameData, ok := objMap["name"]; ok {
			if err := json.Unmarshal(nameData, &name); err != nil {
				return err
			}
		} else {
			return errors.New("type generic missing name")
		}

		def.IdlTypeDefGenericType = &IdlTypeDefGenericType{
			Kind: "type",
			Name: name,
		}
		return nil

	case "const":
		var name string
		if nameData, ok := objMap["name"]; ok {
			if err := json.Unmarshal(nameData, &name); err != nil {
				return err
			}
		} else {
			return errors.New("const generic missing name")
		}

		var typeStr string
		if typeData, ok := objMap["type"]; ok {
			if err := json.Unmarshal(typeData, &typeStr); err != nil {
				return err
			}
		} else {
			return errors.New("const generic missing type")
		}

		def.IdlTypeDefGenericConst = &IdlTypeDefGenericConst{
			Kind: "const",
			Name: name,
			Type: typeStr,
		}
		return nil

	default:
		return fmt.Errorf("unknown typedef generic kind: %s", kind)
	}
}

// IsUint8Variant checks if the variant is a simple uint8 variant
// The variant has no fields data, will be encoded as a simple uint8 enum in `rust`
func (variant *IdlEnumVariant) IsUint8Variant() bool {
	return variant.Fields == nil
}

func (enum *IdlTypeDefTyEnum) IsUint8Enum() bool {
	for _, variant := range enum.Variants {
		// it's a simple uint8 enum if there is no fields data
		if variant.IsUint8Variant() {
			continue
		}
		return false
	}
	return true
}

func (defTy *IdlTypeDefTy) IsStruct() bool {
	return defTy.IdlTypeDefTyStruct != nil
}

func (defTy *IdlTypeDefTy) IsEnum() bool {
	return defTy.IdlTypeDefTyEnum != nil
}

func (defTy *IdlTypeDefTy) IsType() bool {
	return defTy.IdlTypeDefTyType != nil
}

func (defTy *IdlTypeDefTy) GetStruct() *IdlTypeDefTyStruct {
	return defTy.IdlTypeDefTyStruct
}

func (defTy *IdlTypeDefTy) GetEnum() *IdlTypeDefTyEnum {
	return defTy.IdlTypeDefTyEnum
}

func (defTy *IdlTypeDefTy) GetType() *IdlTypeDefTyType {
	return defTy.IdlTypeDefTyType
}

func (defTy *IdlTypeDefTy) UnmarshalJSON(data []byte) error {
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
		return errors.New("typedef type missing kind")
	}

	switch kind {
	case "struct":
		var ty IdlTypeDefTyStruct
		if err := json.Unmarshal(data, &ty); err != nil {
			return err
		}
		defTy.IdlTypeDefTyStruct = &ty
	case "enum":
		var ty IdlTypeDefTyEnum
		if err := json.Unmarshal(data, &ty); err != nil {
			return err
		}
		defTy.IdlTypeDefTyEnum = &ty
	case "type":
		var ty IdlTypeDefTyType
		if err := json.Unmarshal(data, &ty); err != nil {
			return err
		}
		defTy.IdlTypeDefTyType = &ty
	default:
		return fmt.Errorf("unknown typedef type kind: %s", kind)
	}

	return nil
}

func (def *IdlDefinedFields) IsNamed() bool {
	return def != nil && def.IdlDefinedFieldsNamed != nil
}

func (def *IdlDefinedFields) IsTuple() bool {
	return def != nil && def.IdlDefinedFieldsTuple != nil
}

func (def *IdlDefinedFields) GetNamed() *IdlDefinedFieldsNamed {
	return def.IdlDefinedFieldsNamed
}

func (def *IdlDefinedFields) GetTuple() *IdlDefinedFieldsTuple {
	return def.IdlDefinedFieldsTuple
}

func (def *IdlDefinedFields) UnmarshalJSON(data []byte) error {
	var named []IdlField
	if err := json.Unmarshal(data, &named); err == nil && len(named) > 0 && named[0].Name != "" {
		def.IdlDefinedFieldsNamed = &IdlDefinedFieldsNamed{Fields: named}
		return nil
	}

	var types []IdlType
	if err := json.Unmarshal(data, &types); err == nil {
		def.IdlDefinedFieldsTuple = &IdlDefinedFieldsTuple{Types: types}
		return nil
	}

	return errors.New("unable to unmarshal defined fields")
}

func (idlRepr *IdlRepr) IsRust() bool {
	return idlRepr.IdlReprRust != nil
}

func (idlRepr *IdlRepr) IsC() bool {
	return idlRepr.IdlReprC != nil
}

func (idlRepr *IdlRepr) IsTransparent() bool {
	return idlRepr.IdlReprTransparent != nil
}

func (idlRepr *IdlRepr) GetRust() *IdlReprRust {
	return idlRepr.IdlReprRust
}

func (idlRepr *IdlRepr) GetC() *IdlReprC {
	return idlRepr.IdlReprC
}

func (idlRepr *IdlRepr) GetTransparent() *IdlReprTransparent {
	return idlRepr.IdlReprTransparent
}

func (idlRepr *IdlRepr) UnmarshalJSON(data []byte) error {
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
		return errors.New("repr missing kind")
	}

	switch kind {
	case "rust":
		var repr IdlReprRust
		if err := json.Unmarshal(data, &repr); err != nil {
			return err
		}
		idlRepr.IdlReprRust = &repr
	case "c":
		var repr IdlReprC
		if err := json.Unmarshal(data, &repr); err != nil {
			return err
		}
		idlRepr.IdlReprC = &repr
	case "transparent":
		var repr IdlReprTransparent
		if err := json.Unmarshal(data, &repr); err != nil {
			return err
		}
		idlRepr.IdlReprTransparent = &repr
	default:
		return fmt.Errorf("unknown repr kind: %s", kind)
	}

	return nil
}
