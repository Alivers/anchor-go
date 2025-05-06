package idl

import (
	"encoding/json"
	"errors"
	"fmt"
)

type IdlInstructionAccountItem struct {
	*IdlInstructionAccount
	*IdlInstructionAccounts
}

type IdlInstructionAccount struct {
	Name      string   `json:"name"`
	Docs      []string `json:"docs,omitempty"`
	Writable  bool     `json:"writable,omitempty"`
	Signer    bool     `json:"signer,omitempty"`
	Optional  bool     `json:"optional,omitempty"`
	Address   *string  `json:"address,omitempty"`
	Pda       *IdlPda  `json:"pda,omitempty"`
	Relations []string `json:"relations,omitempty"`
}

type IdlInstructionAccounts struct {
	Name     string                      `json:"name"`
	Accounts []IdlInstructionAccountItem `json:"accounts"`
}

type IdlPda struct {
	Seeds   []IdlSeed `json:"seeds"`
	Program *IdlSeed  `json:"program,omitempty"`
}

type IdlSeed struct {
	*IdlSeedConst
	*IdlSeedArg
	*IdlSeedAccount
}

type IdlSeedConst struct {
	Kind  string `json:"kind"`
	Value []byte `json:"value"`
}

type IdlSeedArg struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

// 1. Account is not none
// Ref: https://github.com/solana-foundation/anchor/blob/v0.31.1/lang/syn/src/idl/accounts.rs#L311
// Account is the account's Type of the account that path points to.
// ie.
// IdlSeed::Account(
//
//	IdlSeedAccount {
//	    path: "user_profile.owner".into(),
//	    account: Some("UserProfile".into()),  // Type name of the account `user_profile`
//	}
//
// )
// 2. Account is none
// Ref: https://github.com/solana-foundation/anchor/blob/v0.31.1/lang/syn/src/idl/accounts.rs#L364
type IdlSeedAccount struct {
	Kind    string  `json:"kind"`
	Path    string  `json:"path"`
	Account *string `json:"account,omitempty"`
}

func (item *IdlInstructionAccountItem) GetAccountNum() int {
	if item.IsAccount() {
		return 1
	} else if item.IsAccounts() {
		return item.GetAccounts().GetAccountNum()
	} else {
		return 0
	}
}

func (accounts *IdlInstructionAccounts) GetAccountNum() int {
	if accounts == nil {
		return 0
	}
	num := 0
	for _, item := range accounts.Accounts {
		num += item.GetAccountNum()
	}
	return num
}

func (item *IdlInstructionAccountItem) IsAccount() bool {
	return item.IdlInstructionAccount != nil
}

func (item *IdlInstructionAccountItem) IsAccounts() bool {
	return item.IdlInstructionAccounts != nil
}

func (item *IdlInstructionAccountItem) GetAccount() *IdlInstructionAccount {
	return item.IdlInstructionAccount
}

func (item *IdlInstructionAccountItem) GetAccounts() *IdlInstructionAccounts {
	return item.IdlInstructionAccounts
}

func (item *IdlInstructionAccountItem) UnmarshalJSON(data []byte) error {
	var objMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objMap); err != nil {
		return err
	}

	if _, hasAccounts := objMap["accounts"]; hasAccounts {
		var accounts IdlInstructionAccounts
		if err := json.Unmarshal(data, &accounts); err != nil {
			return err
		}
		item.IdlInstructionAccounts = &accounts
	} else {
		var account IdlInstructionAccount
		if err := json.Unmarshal(data, &account); err != nil {
			return err
		}
		item.IdlInstructionAccount = &account
	}

	return nil
}

func (seed *IdlSeed) IsConst() bool {
	return seed.IdlSeedConst != nil
}

func (seed *IdlSeed) IsArg() bool {
	return seed.IdlSeedArg != nil
}

func (seed *IdlSeed) IsAccount() bool {
	return seed.IdlSeedAccount != nil
}

func (seed *IdlSeed) GetConst() *IdlSeedConst {
	return seed.IdlSeedConst
}

func (seed *IdlSeed) GetArg() *IdlSeedArg {
	return seed.IdlSeedArg
}

func (seed *IdlSeed) GetAccount() *IdlSeedAccount {
	return seed.IdlSeedAccount
}

func (idlSeed *IdlSeed) UnmarshalJSON(data []byte) error {
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
		return errors.New("seed missing kind")
	}

	switch kind {
	case "const":
		var seed IdlSeedConst
		if err := json.Unmarshal(data, &seed); err != nil {
			return err
		}
		idlSeed.IdlSeedConst = &seed
	case "arg":
		var seed IdlSeedArg
		if err := json.Unmarshal(data, &seed); err != nil {
			return err
		}
		idlSeed.IdlSeedArg = &seed
	case "account":
		var seed IdlSeedAccount
		if err := json.Unmarshal(data, &seed); err != nil {
			return err
		}
		idlSeed.IdlSeedAccount = &seed
	default:
		return fmt.Errorf("unknown seed kind: %s", kind)
	}

	return nil
}
