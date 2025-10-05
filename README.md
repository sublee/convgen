# Convgen

Type-safe conversion code generator for Go.

## Overview

Convgen eliminates tons of manual boilerplate code in type conversion. Declare a conversion with a type pair and its configuration once, and the generator produces the converter implementation. Type-safe settings catch configuration errors at compile time, while unmatched fields are diagnosed at generation time, enabling fast and confident refactoring.

## Features

- **Struct-to-struct** conversions with automatic field matching
- **Enum-to-enum** conversions with value mapping
- **Union-to-union** conversions for idiomatic interface patterns
- **Type-safe configuration** - catch errors at compile time
- **Flexible renaming rules** - handle naming convention differences
- **Custom conversion functions** - integrate your own logic
- **Error-aware conversions** - support fallible transformations
- **Module system** - share configurations and reuse converters

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
