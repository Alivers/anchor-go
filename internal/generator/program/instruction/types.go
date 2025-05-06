package instruction

import "github.com/alivers/anchor-go/internal/idl"

type pdaSeedValue struct {
	OriginIdlSeed *idl.IdlSeed
	// Const value
	SeedConst []byte
	// Ref value
	SeedRef *pdaSeedRef
}

type pdaSeedRef struct {
	SeedRefPath string
	SeedRefName string
	RefType     *idl.IdlType
}
