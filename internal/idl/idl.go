package idl

import (
	"slices"
)

// Ref: https://github.com/solana-foundation/anchor/blob/v0.31.1/idl/spec/src/lib.rs
const IDL_SPEC = "0.31.1"

type Idl struct {
	Address      string           `json:"address"`
	Metadata     IdlMetadata      `json:"metadata"`
	Docs         []string         `json:"docs,omitempty"`
	Instructions []IdlInstruction `json:"instructions"`
	Accounts     []IdlAccount     `json:"accounts,omitempty"`
	Events       []IdlEvent       `json:"events,omitempty"`
	Errors       []IdlErrorCode   `json:"errors,omitempty"`
	Types        []IdlTypeDef     `json:"types,omitempty"`
	Constants    []IdlConst       `json:"constants,omitempty"`
}

type IdlMetadata struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Spec         string          `json:"spec"`
	Address      *string         `json:"address,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Repository   *string         `json:"repository,omitempty"`
	Dependencies []IdlDependency `json:"dependencies,omitempty"`
	Contact      *string         `json:"contact,omitempty"`
	Deployments  *IdlDeployments `json:"deployments,omitempty"`
}

type IdlDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type IdlDeployments struct {
	Mainnet  *string `json:"mainnet,omitempty"`
	Testnet  *string `json:"testnet,omitempty"`
	Devnet   *string `json:"devnet,omitempty"`
	Localnet *string `json:"localnet,omitempty"`
}

type IdlInstruction struct {
	Name          string                      `json:"name"`
	Docs          []string                    `json:"docs,omitempty"`
	Discriminator IdlDiscriminator            `json:"discriminator"`
	Accounts      []IdlInstructionAccountItem `json:"accounts"`
	Args          []IdlField                  `json:"args"`
	Returns       IdlType                     `json:"returns,omitempty"`
	// !!! Notice: `Discriminant` is not in the original spec.
	Discriminant *IdlDiscriminant `json:"discriminant,omitempty"`
}

type IdlAccount struct {
	Name          string           `json:"name"`
	Discriminator IdlDiscriminator `json:"discriminator"`
	// !!! Notice: `Type` is for the old idl spec.
	// It is not in the original spec.
	Type IdlTypeDefTy `json:"type,omitempty"`
}

type IdlEvent struct {
	Name          string           `json:"name"`
	Discriminator IdlDiscriminator `json:"discriminator"`
}

type IdlConst struct {
	Name  string   `json:"name"`
	Docs  []string `json:"docs,omitempty"`
	Type  IdlType  `json:"type"`
	Value string   `json:"value"`
}

type IdlErrorCode struct {
	Code int     `json:"code"`
	Name string  `json:"name"`
	Msg  *string `json:"msg,omitempty"`
}

type IdlDiscriminator []byte

// !!! Notice: `Discriminant` is not in the original spec.
type IdlDiscriminant struct {
	Type  string `json:"type"`
	Value uint   `json:"value"`
}

func (idl *Idl) FindTypeByName(name string) *IdlTypeDef {
	for _, item := range idl.Types {
		if item.Name == name {
			return &item
		}
	}
	return nil
}

func (ins *IdlInstruction) GetAccountNum() int {
	if ins == nil {
		return 0
	}
	num := 0
	for _, item := range ins.Accounts {
		num += item.GetAccountNum()
	}
	return num
}

type instructionAccount struct {
	Account       *IdlInstructionAccount
	Parents       []*IdlInstructionAccounts
	IndexInParent int
}

func (ins *IdlInstruction) GetAccounts() []*IdlInstructionAccount {
	accounts := ins.GetAccountsWithRelation()
	result := make([]*IdlInstructionAccount, len(accounts))
	for i, account := range accounts {
		result[i] = account.Account
	}
	return result
}

func (ins *IdlInstruction) GetAccountsWithRelation() []*instructionAccount {
	if ins == nil {
		return nil
	}

	accounts := make([]*instructionAccount, 0)
	for _, item := range ins.Accounts {
		result := item.flatenAccounts(nil, -1)
		accounts = append(accounts, result...)
	}
	return accounts
}

func (accountItem IdlInstructionAccountItem) flatenAccounts(parents []*IdlInstructionAccounts, indexInParent int) []*instructionAccount {
	if accountItem.IsAccount() {
		return []*instructionAccount{
			{
				Account:       accountItem.GetAccount(),
				Parents:       parents,
				IndexInParent: indexInParent,
			},
		}
	} else if accountItem.IsAccounts() {
		accounts := accountItem.GetAccounts()

		newParents := append(slices.Clone(parents), accounts)

		result := make([]*instructionAccount, 0)
		for i, account := range accounts.Accounts {
			tmp := account.flatenAccounts(newParents, i)
			result = append(result, tmp...)
		}
		return result
	}
	return nil
}
