package model

const (
	PkgSolanaGo       = "github.com/gagliardetto/solana-go"
	PkgSolanaGoText   = "github.com/gagliardetto/solana-go/text"
	PkgDfuseBinary    = "github.com/gagliardetto/binary"
	PkgTreeout        = "github.com/gagliardetto/treeout"
	PkgFormat         = "github.com/gagliardetto/solana-go/text/format"
	PkgGoFuzz         = "github.com/gagliardetto/gofuzz"
	PkgMsgpack        = "github.com/vmihailenco/msgpack/v5"
	PkgTestifyRequire = "github.com/stretchr/testify/require"
	PkgAgRpc          = "github.com/gagliardetto/solana-go/rpc"
	PkgSpew           = "github.com/davecgh/go-spew/spew"
	PkgEncodingBinary = "encoding/binary"
	PkgFmt            = "fmt"
	PkgBytes          = "bytes"
	PkgBigInt         = "math/big"
)

type DiscriminatorType string

const (
	DiscriminatorTypeUvarint32 DiscriminatorType = "uvarint32"
	DiscriminatorTypeUint32    DiscriminatorType = "u32"
	DiscriminatorTypeUint8     DiscriminatorType = "u8"
	DiscriminatorTypeAnchor    DiscriminatorType = "anchor"
	DiscriminatorTypeDefault   DiscriminatorType = "default"
)

func (d DiscriminatorType) String() string {
	return string(d)
}

type EncoderType string

const (
	// github.com/gagliardetto/binary: NewBinEncoder, NewBinDecoder
	EncoderTypeBin EncoderType = "bin"
	// github.com/gagliardetto/binary: NewBorshEncoder, NewBorshDecoder
	EncoderTypeBorsh EncoderType = "borsh"
	// https://docs.solana.com/developing/programming-model/transactions#compact-array-format
	EncoderTypeCompactU16 EncoderType = "compact-u16"
)

func (e EncoderType) String() string {
	return string(e)
}

func (name EncoderType) GetNewEncoderName() string {
	switch name {
	case EncoderTypeBin:
		return "NewBinEncoder"
	case EncoderTypeBorsh:
		return "NewBorshEncoder"
	case EncoderTypeCompactU16:
		return "NewCompact16Encoder"
	default:
		panic(name)
	}
}

func (name EncoderType) GetNewDecoderName() string {
	switch name {
	case EncoderTypeBin:
		return "NewBinDecoder"
	case EncoderTypeBorsh:
		return "NewBorshDecoder"
	case EncoderTypeCompactU16:
		return "NewCompact16Decoder"
	default:
		panic(name)
	}
}
