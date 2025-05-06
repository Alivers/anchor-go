package common

import (
	"github.com/alivers/anchor-go/internal/generator/helper"
)

func GetInstructionEnumName(instructionExportedName string) string {
	return "Instruction_" + instructionExportedName
}

func formatSimpleEnumVariantName(enumTypeName string, enumVariantName string) string {
	return helper.ToCamelCase(enumTypeName + "_" + enumVariantName)
}

func GetComplexEnumVariantTypeName(enumTypeName string, enumVariantName string) string {
	return helper.ToCamelCase(enumTypeName + "_" + enumVariantName)
}

func GetEnumVariantsContainerName(enumTypeName string) string {
	return helper.ToLowerCamelCase(enumTypeName) + "Container"
}

func GetComplexEnumInterfaceMethodName(enumTypeName string) string {
	return "is" + helper.ToCamelCase(enumTypeName)
}

func GetTupleStructElementName(fieldIndex int) string {
	return "Elem_" + helper.IntToStr(fieldIndex)
}

func GetInstructionConstructorName(instructionExportedName string) string {
	return "New" + instructionExportedName + "Instruction"
}

func GetDiscriminatorName(indentName string) string {
	return indentName + "Discriminator"
}
