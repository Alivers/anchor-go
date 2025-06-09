# anchor-go

A Go client generator for Solana programs written using the Anchor framework. This project is a rewrite and enhancement of several existing implementations, including [gagliardetto/anchor-go](https://github.com/gagliardetto/anchor-go), [fragmetric-labs/solana-anchor-go](https://github.com/fragmetric-labs/solana-anchor-go) and [daog1/solana-anchor-go](https://github.com/daog1/solana-anchor-go).

## Features

- Extensive support for Anchor IDL specification (most common use cases)
- Improved code generation with better maintainability
- Enhanced type safety and error handling
- Support for uint8 discriminant in instructions
- Support for all Anchor program components:
  - Instructions
  - Accounts
  - Types
  - Events
  - Errors
  - Tuple types
  - Constants

## Idl Spec

[Anchor Idl Spec](./internal/idl/idl.go#L7)

## Usage

```bash
# Build the project
$ go build

# Generate Go client from an Anchor IDL
$ ./anchor-go -src=./example/dummy_idl.json -pkg=dummy -dst=./generated/dummy
```

Generated code will be saved to the specified destination directory (`./generated/` in the example above).

## Development Status

All core features have been implemented and are actively maintained:

- [x] Instructions
- [x] Accounts
- [x] Types
- [x] Events
- [x] Errors
- [x] Tuple types
- [x] Constants

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
