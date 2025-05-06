# anchor-go

## Idl Spec

[Anchor Idl Spec](./internal/idl/idl.go#L7)

## Usage

```bash
$ go build
$ ./anchor-go -src=./example/dummy_idl.json -pkg=dummy -dst=./generated/dummy
```

Generated Code will be generated and saved to `./generated/`.
And check `./example/dummy_test.go` for generated code usage.

## Test

## TODO

- [x] instructions
- [x] accounts
- [x] types
- [x] events
- [x] errors
- [x] handle tuple types
- [x] constants
