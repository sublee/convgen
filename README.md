# Convgen

![Convgen Logo](assets/convgen.jpg)]

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
- **Type-safe configuration**  
  All options are validated at compile time — no reflection, tags, string, or
  comment-based directives.
- **Detailed diagnostics**  
  *All* matching and conversion errors in a single pass are reported together,
  so you can fix everything at once instead of stopping at the first error.

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

## Installation

```bash
go install github.com/sublee/convgen
```

## Quick Start

1. Add a build constraint to files containing Convgen directives:

```go
//go:build convgen
```

2. Declare your conversions:

```go
//go:build convgen
package main

import "github.com/sublee/convgen"

// Simple struct conversion
var EncodeUser = convgen.Struct[User, api.User](nil)
```

3. Run the generator:

```bash
go run github.com/sublee/convgen/cmd/convgen ./...
```

This generates `convgen_gen.go` with the implementation:

```go
func EncodeUser(in User) (out api.User) {
    out.Name = in.Name
    out.Email = in.Email
    return
}
```

## Example

```go
//go:build convgen

package main

import "github.com/sublee/convgen"

// Create a module with shared configuration
var mod = convgen.Module(
    convgen.RenameToLower(true, true),
)

// Declare conversions
var (
    EncodeUser = convgen.Struct[User, api.User](mod,
        convgen.Match(User{}.ID, api.User{}.Id), // explicit field matching
    )
    EncodeRole = convgen.Enum[Role, api.Role](mod,
        api.ROLE_UNSPECIFIED, // default value
        convgen.RenameTrimCommonPrefix(true, true),
    )
)
```

## Configuration

When field mappings are ambiguous, Convgen provides detailed diagnostics:

```
main.go:10:10: invalid match between User and api.User
    FAIL: ID -> ?  // missing
    FAIL: ?  -> Id // missing
```

Resolve with renaming rules:

```go
var EncodeUser = convgen.Struct[User, api.User](nil,
    convgen.RenameToLower(true, true),
)
```

Or explicit matching:

```go
var EncodeUser = convgen.Struct[User, api.User](nil,
    convgen.Match(User{}.ID, api.User{}.Id),
)
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

Heungsub Lee
