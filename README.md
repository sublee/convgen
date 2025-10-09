# Convgen

[![CI](https://github.com/sublee/convgen/actions/workflows/ci.yaml/badge.svg)](https://github.com/sublee/convgen/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sublee/convgen.svg)](https://pkg.go.dev/github.com/sublee/convgen)

> [!WARNING]
> Convgen is in early development. If you have any feedback, please open an
> issue.

<p align="center">
<img src="assets/convgen.png" alt="Convgen Logo" width="320" />
</p>

Convgen generates **type-to-type conversion code** for Go with **type-safe
configuration** and **detailed diagnostics**.

```go
//go:build convgen

var EncodeUser = convgen.Struct[User, api.User](nil,
    convgen.RenameReplace("", "", "Id", "ID"), // Replace Id with ID in output types before matching
    convgen.Match(User{}.Name, api.User{}.Username), // Explicit field matching
)
```

## Features

- **Struct, Union, and Enum conversions**  
  Automatically matches fields, implementations, and members by name.

  ```go
  convgen.Struct[User, api.User]      // func(User) api.User
  convgen.StructErr[User, api.User]   // func(User) (api.User, error)
  convgen.Union[Job, api.Job]         // func(Job) api.Job -- type UploadJob, type OrderJob, ...
  convgen.UnionErr[Job, api.Job]      // func(Job) (api.Job, error)
  convgen.Enum[Status, api.Status]    // func(Status) api.Status -- const StatusTodo, const StatusPending, ...
  convgen.EnumErr[Status, api.Status] // func(Status) (api.Status, error)
  ```

- **Type-safe configuration**  
  All options are validated at compile time — no reflection, tags, string, or
  comment-based directives.

  ```go
  // Custom conversion functions must be reachable.
  convgen.ImportFunc(strconv.Itoa)

  // If User{}.Name is renamed by a refactoring tool,
  // this directive will be updated accordingly.
  convgen.Match(User{}.Name, api.User{}.Username)
  ```
  
- **Detailed diagnostics**  
  *All* matching and conversion errors in a single pass are reported together,
  so you can fix everything at once instead of stopping at the first error. In
  addition, Convgen provides [Lint](#lint) support for real-time feedback during
  development.

  ```
  main.go:10:10: invalid match between User and api.User
      ok:   Name    -> Username // forced at main.go:12:2
      ok:   ID      -> ID [Id]
      ok:   GroupID -> GroupID [GroupId]
      FAIL: ?       -> Email // missing
      FAIL: EMail   -> ?     // missing
  ```

## Motivation

Convgen is inspired by both [goverter](https://github.com/jmattheis/goverter)
and [Wire](https://github.com/google/wire). While goverter is powerful for
generating type conversion code, it relies on comment-based directives that are
not validated at compile time. Moreover, because it stops at the first error,
refactoring becomes difficult when target types change. In contrast, Wire offers
type-safe configuration and detailed diagnostics, but focuses on dependency
injection. Convgen combines the best of both worlds, bringing **type-safe
configuration** and **comprehensive diagnostics** to
**type conversion code generation**.

## Quick Start

1. Install Convgen:

    ```bash
    go install github.com/sublee/convgen
    ```

2. Add a build constraint to files containing Convgen directives:

    ```go
    //go:build convgen
    ```

3. Declare your conversions:

    ```go
    var EncodeUser = convgen.Struct[User, api.User](nil)
    ```

4. Run the generator:

    ```bash
    convgen ./...
    ```

5. Convgen generates a `convgen_gen.go` file by copying your `//go:build convgen`
   files and rewriting Convgen directives:

    ```go
    func EncodeUser(in User) (out api.User) {
        out.Name = in.Name
        out.Email = in.Email
        return
    }
    ```

## Lint

Convgen provides a [golangci-lint](https://github.com/golangci/golangci-lint)
plugin that validates Convgen directives during development inside your IDE.

It's not an official plugin, so you need to build it manually:

```bash
make golangci-lint-convgen
```

Then, add the following configuration to your `.golangci.yaml`. Because Convgen
directives are only valid when the `convgen` build tag is set, make sure to
include it under the `run.build-tags` section:

```yaml
version: "2"

run:
  build-tags:
    - convgen

linters:
  settings:
    custom:
      convgen:
        type: module
```

Now the `golangci-lint-convgen` binary can validate Convgen directives in your
project. To make this process seamless in your IDE, you can configure it as the
default linter. For Visual Studio Code, add the following to your settings:

```json
{
  "go.lintTool": "golangci-lint-v2",
  "go.lintFlags": ["--build-tags=convgen"],
  "go.alternateTools": {
    "golangci-lint-v2": "/path/to/golangci-lint-convgen",
  }
}
```

With this setup, your Convgen directives will be validated in real time as you
code.

## License

MIT License — see [LICENSE](LICENSE) for details.

## Author

[Heungsub Lee](https://subl.ee/), with assistance from ChatGPT
