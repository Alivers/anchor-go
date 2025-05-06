package instruction

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alivers/anchor-go/internal/idl"
)

func newInstructionBuilderName(instExportedName string) string {
	return "New" + instExportedName + "InstructionBuilder"
}

func instAccountsBuilderStructName(instExportedName string, internalGroup string) string {
	return instExportedName + internalGroup + "AccountsBuilder"
}

func instAccountSetterWithBuilderName(internalGroup string) string {
	return "Set" + internalGroup + "AccountsFromBuilder"
}

func instAccountAccessorName(accessor, childAccountName string) string {
	return accessor + childAccountName + "Account"
}

func instPdaAccountDerivationExportedFuncName(accountExportedName string) string {
	return "Find" + accountExportedName + "Address"
}

func instPdaAccountDerivationPrivateFuncName(accountExportedName string) string {
	return "find" + accountExportedName + "Address"
}

func instPdaAccountDerivationWithBumpSeedFuncName(accountExportedName string) string {
	return "find" + accountExportedName + "AddressWithBumpSeed"
}

func buildInstAccountGroupPath(ancestors []*idl.IdlInstructionAccounts) string {
	groupPath := ""
	for _, parent := range ancestors {
		if groupPath == "" {
			groupPath = parent.Name
		} else {
			groupPath += "/" + parent.Name
		}
	}
	return groupPath
}

func buildInstructionAccountComments(
	instructionAccountNum, accountIndexInInst int,
	prevGroupPath, groupPath string,
	account *idl.IdlInstructionAccount,
) (comments string) {
	comment := &strings.Builder{}
	indent := 1
	var prepend int

	if groupPath != "" {
		thisGroupName := filepath.Base(groupPath)
		if strings.Count(groupPath, "/") == 0 {
			prepend = indent
		} else {
			prepend = indent + (strings.Count(groupPath, "/") * 2) + len(strings.TrimSuffix(groupPath, thisGroupName)) - 1
		}

		// Add one more indent for `colon`
		indent = len(thisGroupName) + 1
		if prevGroupPath != groupPath {
			comment.WriteString("// " + strings.Repeat("·", prepend-1) + fmt.Sprintf("%s: ", thisGroupName))
		} else {
			comment.WriteString("// " + strings.Repeat("·", prepend+indent-1) + " ")
		}
	} else {
		comment.WriteString("// ")
	}

	noStr := fmt.Sprintf("[%v] = ", accountIndexInInst)
	comment.WriteString(noStr)
	comment.WriteString("[")
	if account.Writable {
		comment.WriteString("WRITE")
	}
	if account.Signer {
		if account.Writable {
			comment.WriteString(", ")
		}
		comment.WriteString("SIGNER")
	}
	comment.WriteString("] ")
	comment.WriteString(account.Name)

	for _, doc := range account.Docs {
		comment.WriteString("\n// " + strings.Repeat("·", prepend+indent-1+len(noStr)) + " " + doc)
	}
	if accountIndexInInst < instructionAccountNum-1 {
		comment.WriteString("\n//")
	}

	return comment.String()
}
