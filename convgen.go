// Package convgen provides directives for type-safe conversion code generation.
//
// Convgen eliminates tons of manual boilerplate code in type conversion.
// Declare a conversion with a type pair and its configuration once, and the
// generator produces the converter implementation. Type-safe settings catch
// configuration errors at compile time, while unmatched fields are diagnosed at
// generation time, enabling fast and confident refactoring.
//
// To start with Convgen, add a build constraint to files containing Convgen
// directives:
//
//	//go:build convgen
//
// Conversions can be declared with Convgen directives. Conversions between
// struct-to-struct ([Struct]), idiomatic union-style interface-to-interface
// ([Union]), and enum-to-enum ([Enum]) are supported. Convgen automaticall
// matches their fields, implementations, and values by name. It also provides
// configurable renaming rules, integration with custom conversion functions,
// and error-aware conversions for flexible adaptation to various use cases:
//
//	// source:
//	var EncodeUser = convgen.Struct[User, api.User](nil)
//
//	// generated: (simplified)
//	func EncodeUser(in User) (out api.User) {
//		out.Name = in.Name
//		out.Email = in.Email
//		return
//	}
//
// After declaring conversions, run the convgen command. It will generate
// convgen_gen.go for your package:
//
//	go run github.com/sublee/convgen/cmd/convgen
//
// # Configurations
//
// When field mappings are ambiguous or incomplete, Convgen reports detailed
// diagnostics. For example, if our User has an ID field but api.User has Id
// (with a lower case "d") instead, so they don't match exactly:
//
//	main.go:10:10: invalid match between User and api.User
//		FAIL: ID -> ?  // missing
//		FAIL: ?  -> Id // missing
//
// Renaming rules can be applied to resolve those mismatches. In this case, we
// can solve with just [RenameToLower]. It renames User.ID and api.User.Id both
// to become "id":
//
//	// source:
//	var EncodeUser = convgen.Struct[User, api.User](nil,
//		convgen.RenameToLower(true, true),
//	)
//
// There are many out-of-the-box renaming options. See [Option] for your
// use case.
//
// Or, we can match them explicitly with [Match]. Because of it referring the
// fields directly as code, we can detect broken configuration at compile time
// in the future:
//
//	// source:
//	var EncodeUser = convgen.Struct[User, api.User](nil,
//		convgen.Match(User{}.ID, api.User{}.Id),
//	)
//
// Note that many options have separate flags or arguments for input and output
// types symmetrically.
//
// # Modules
//
// A converter may depend on other converters to convert inner types. [Module]
// provides a shared namespace so that they can refer to each other
// automatically. A module also holds the default configurations for the
// underlying converters for uniformity:
//
//	// source:
//	var (
//		enc = convgen.Module(convgen.RenameToLower(true, true))
//		EncodeUser = convgen.Struct[User, api.User](enc)
//		EncodeRole = convgen.Enum[Role, api.Role](enc, api.ROLE_UNSPECIFIED, convgen.RenameTrimCommonPrefix(true, true))
//	)
//
//	// generated: (simplified)
//	func EncodeUser(in User) (out api.User) {
//		out.Name = in.Name
//		out.Email = in.Email
//		out.Role = EncodeRole(in.Role)
//		           ^^^^^^^^^^^^^^^^^^^ // reused converter in the same module
//		return
//	}
//	func EncodeRole(in Role) (out api.Role) {
//		switch in {
//		case RoleAdmin:
//			return api.ROLE_ADMIN
//		case RoleMember:
//			return api.ROLE_MEMBER
//		default:
//			return api.ROLE_UNSPECIFIED
//		}
//	}
//
// Conversions are often more than field-to-field copies. If Convgen cannot
// determine how to convert a type to another type, it reports an error.
// However, a custom function can be used for any type conversion. A custom
// function can be imported into a module by [ImportFunc].
//
// For example, when User.ID is int but api.User.ID is string, a func(int)
// string like strconv.Itoa can be imported to convert them:
//
//	// source:
//	var (
//		enc = convgen.Module(convgen.ImportFunc(strconv.Itoa))
//		EncodeUser = convgen.Struct[User, api.User](enc)
//	)
//
//	// generated: (simplified)
//	func EncodeUser(in User) (out api.User) {
//		out.ID = strconv.Itoa(in.ID)
//		         ^^^^^^^^^^^^^^^^^^^ // custom conversion function
//		out.Name = in.Name
//		out.Email = in.Email
//		return
//	}
//
// # Errors
//
// Converter directives have Err variants: [StructErr], [UnionErr], and
// [EnumErr]. They generate converter functions that may return an error. They
// can use other errorful converters in the same module. Custom errorful
// conversion functions can be imported by [ImportFuncErr].
//
// In the above example, we could encode User to api.User without any error. But
// the reverse is not possible because string to int conversion (api.User.ID to
// User.ID) may fail at runtime. So, we need to use the Err variants:
//
//	// source:
//	var (
//		dec = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))
//		DecodeUser = convgen.StructErr[api.User, User](dec)
//	)
//
//	// generated: (simplified)
//	func DecodeUser(in api.User) (out User, err error) {
//	                                        ^^^^^^^^^ // may return error
//		out.ID, err = strconv.Atoi(in.ID)
//		        ^^^^^^^^^^^^^^^^^^^^^^^^^ // custom conversion function with error
//		out.Name = in.Name
//		out.Email = in.Email
//		return
//	}
//
// Note that converters returning error may use errorless converters in the same
// module, but not vice versa. This restriction ensures that errorless
// converters never return an error at runtime.
package convgen

// module provides a shared namespace and default configurations for underlying
// converters. This is unexported so there is no way to create a module other
// than [Module].
type module *struct{}

type (
	canUseFor interface{ canUseFor() }
	yes       interface{ canUseFor }
	no        interface{ canUseFor }

	// option for [Module]
	moduleOption interface{ moduleOption() yes }

	// option for [ForStruct], [ForUnion], [ForEnum], and [Module]
	forOption interface{ forOption() yes }

	// option for [Struct] and [StructErr]
	structOption interface{ structOption() yes }

	// option for [Union] and [UnionErr]
	unionOption interface{ unionOption() yes }

	// option for [Enum] and [EnumErr]
	enumOption interface{ enumOption() yes }
)

// Module provides a shared namespace of underlying converters to make them
// discover and call each other to convert inner types. Also holds the default
// configurations for the underlying converters for uniformity. Subconverters,
// what are type converters Convgen has generated implicitly, inherit the module
// and its configurations rather than the explicit parent converter.
//
// Pass a module as the first argument of a converter directive, then the
// converter belongs to the module:
//
//	var mod = convgen.Module(convgen.RenameToLower(true, true))
//	var conv = convgen.Struct[Foo, Bar](mod)
//
// To import arbitrary type converters into the namespace, use [ImportFunc] or
// [ImportFuncErr]. To split default configurations for different kinds of
// converters, use [ForStruct], [ForUnion], or [ForEnum] to qualify options. To
// register error wrappers, use [ImportErrWrap].
func Module(opts ...moduleOption) module {
	panic("convgen: not generated")
}

// Struct directive generates a converter function between two struct types
// without error:
//
//	// source:
//	var convUser = convgen.Struct[User, api.User](nil)
//
// The input and output types are declared as type parameters. The variable that
// holds the directive is rewritten to the actual function when Convgen
// generates code:
//
//	// generated: (simplified)
//	func convUser(in User) (out api.User) {
//		out.Name = in.Name
//		out.Email = in.Email
//		out.Address = convgen_Address_api_Address(in.Address) // subconverter implicitly generated
//		return
//	}
//
// By default, fields are matched by name. If any field cannot be matched,
// Convgen reports errors at generation time. Use options such as [Match] or
// renaming rules (e.g., [RenameReplace], [RenameToLower]) to control the
// matching behavior. You can also enable [DiscoverGetters] or [DiscoverSetters]
// to match getter/setter methods instead of accessing fields directly.
//
// Since the generated function does not return an error, it cannot call other
// errorful functions. If error handling is required, use [StructErr] instead.
func Struct[In, Out any](mod module, opts ...structOption) func(In) Out {
	panic("convgen: not generated")
}

// StructErr is the error-returning variant of [Struct]. It generates a
// converter function that returns (Out, error) instead of just Out.
//
// Unlike [Struct], StructErr allows the generated converter to call other
// functions within the same [Module] that may themselves return an error.
func StructErr[In, Out any](mod module, opts ...structOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

// Union directive generates a converter function between two interface types
// without error: Typically, union implementations share a common suffix, so
// [RenameTrimCommonWordSuffix] is often used to match them:
//
//	// source:
//	var convEvent = convgen.Union[Event, api.Event](nil,
//		convgen.RenameTrimCommonWordSuffix(true, true),
//	)
//
// The input and output types are declared as type parameters. The variable that
// holds the directive is rewritten to the actual function when Convgen
// generates code:
//
//	// generated: (simplified)
//	func convEvent(in Event) api.Event {
//		switch in := in.(type) {
//		case ClickEvent:
//			return convgen_ClickEvent_api_ClickEvt(in) // subconverter implicitly generated
//		case ScrollEvent:
//			return convgen_ScrollEvent_api_ScrollEvt(in)
//		}
//		return nil
//	}
//
// By default, Convgen discovers concrete implementations from the package that
// defines each interface. When [DiscoverBySample] is used, Convgen finds
// implementations from the package of the sample value instead.
//
// To customize conversions for each implementation, declare corresponding
// converters within the same [Module]:
//
//	// source:
//	var (
//		mod             = convgen.Module()
//		convEvent       = convgen.Union[Event, api.Event](mod)
//		convClickEvent  = convgen.Struct[ClickEvent, api.ClickEvt](mod, ...)
//		convScrollEvent = convgen.Struct[ScrollEvent, api.ScrollEvt](mod, ...)
//	)
//
//	// generated: (simplified)
//	func convEvent(in Event) api.Event {
//		switch in := in.(type) {
//		case ClickEvent:
//			return convClickEvent(in)
//		case ScrollEvent:
//			return convScrollEvent(in)
//		}
//		return nil
//	}
func Union[In, Out any](mod module, opts ...unionOption) func(In) Out {
	panic("convgen: not generated")
}

// UnionErr is the error-returning variant of [Union]. It generates a converter
// function that returns (Out, error) instead of just Out. When there is no
// match for the input implementation, it returns (nil,
// convgenerrors.ErrNoMatch).
//
// Unlike [Union], UnionErr allows the generated converter to call other
// functions within the same [Module] that may themselves return an error.
func UnionErr[In, Out any](mod module, opts ...unionOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

// Enum directive generates a converter function between two enum types without
// error. The default output member must be specified explicitly. Typically,
// enum members share a common prefix, so [RenameTrimCommonWordPrefix] is often
// used to match them:
//
//	// source:
//	var convStatus = convgen.Enum[Status, api.Status](nil, api.STATUS_UNSPECIFIED,
//		convgen.RenameTrimCommonWordPrefix(true, true),
//	)
//
// The input and output types are declared as type parameters. The variable that
// holds the directive is rewritten to the actual function when Convgen
// generates code:
//
//	// generated: (simplified)
//	func convStatus(in Status) api.Status {
//		switch in {
//		case StatusActive:
//			return api.STATUS_Active
//		case StatusInactive:
//			return api.STATUS_Inactive
//		default:
//			return api.STATUS_UNSPECIFIED // default output member
//		}
//	}
//
// By default, Convgen discovers enum members (constant identifiers) from the
// package that defines each enum type. When [DiscoverBySample] is used, Convgen
// discovers members from the package of the sample value instead.
func Enum[In, Out any](mod module, default_ Out, opts ...enumOption) func(In) Out {
	panic("convgen: not generated")
}

// EnumErr is the error-returning variant of [Enum]. It generates a converter
// function that returns (Out, error) instead of just Out. When there is no
// match for the input value, it returns (unknown, convgenerrors.ErrNoMatch).
//
// Unlike [Enum], EnumErr allows the generated converter to call other functions
// within the same [Module] that may themselves return an error.
func EnumErr[In, Out any](mod module, default_ Out, opts ...enumOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

// Option configures how converters are generated. They are categorized by their
// prefix:
//
//   - Discover: Configures how Convgen discovers targets such as fields,
//     implementations, or enum members.
//   - Import: Registers a custom conversion function or error wrapper so that
//     converters in the module can use them.
//   - Match: Manually matches or skips a specific pair, optionally with a
//     custom matcher function.
//   - Rename: Appends or resets renaming rules before matching fields,
//     implementations, or members. The rules are applied in the order they are
//     registered.
//
// For-prefixed options are meta-options that restrict where the registered
// options apply. For example, an option registered with [ForStruct] affects
// only struct converters within the module.
//
// Not every option can be applied to every converter directive. There are five
// scopes of options:
//
//  1. Module-level options: accepted by [Module] only.
//  2. For-qualifier options: accepted by [ForStruct], [ForUnion], and [ForEnum] only.
//  3. Struct-level options: accepted by [Struct], [StructErr], and [ForStruct] only.
//  4. Union-level options: accepted by [Union], [UnionErr], and [ForUnion] only.
//  5. Enum-level options: accepted by [Enum], [EnumErr], and [ForEnum] only.
//
// The type parameters of [Option] indicate which scopes the option can be
// applied to. For example, Option[yes, no, yes, no, yes] can be applied to
// module-level, struct-level, and enum-level directives. But not for-qualifier
// or union-level ones.
type Option[Module, For, Struct, Union, Enum canUseFor] interface {
	moduleOption() Module
	forOption() For
	structOption() Struct
	unionOption() Union
	enumOption() Enum
}

// ForStruct qualifies options to apply only to struct converters within the module.
//
//	// source:
//	var mod = convgen.Module(
//		convgen.ForStruct(convgen.DiscoverUnexported(true, false)),
//	)
//
//	// Applies to struct converters. Unexported fields of User are discovered.
//	var convUser = convgen.Struct[User, api.User](mod)
//
//	// Does not apply to enum converters.
//	var convStatus = convgen.Enum[Status, api.Status](mod)
//
// When this option is specified multiple times, all of them are applied in
// order.
func ForStruct(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ForUnion qualifies options to apply only to union converters within the module.
//
//	// source:
//	var mod = convgen.Module(
//		convgen.ForUnion(convgen.RenameTrimCommonWordSuffix(true, false)),
//	)
//
//	// Applies to union converters. The common suffix of Event implementations are trimmed.
//	var convEvent = convgen.Union[Event, api.Event](mod)
//
//	// Does not apply to struct converters.
//	var convUser = convgen.Struct[User, api.User](mod)
//
// When this option is specified multiple times, all of them are applied in
// order.
func ForUnion(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ForEnum qualifies options to apply only to enum converters within the module.
//
//	// source:
//	var mod = convgen.Module(
//		convgen.ForEnum(convgen.RenameTrimCommonWordPrefix(true, false)),
//	)
//
//	// Applies to enum converters: The common prefix of Status members are trimmed.
//	var convStatus = convgen.Enum[Status, api.Status](mod)
//
//	// Does not apply to struct converters.
//	var convUser = convgen.Struct[User, api.User](mod)
//
// When this option is specified multiple times, all of them are applied in
// order.
func ForEnum(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ImportFunc registers a custom errorless conversion function (func(In) Out)
// with the module. Converters within the module may call this function when
// converting fields of the corresponding types:
//
//	// source:
//	var mod = convgen.Module(convgen.ImportFunc(strconv.Itoa))
//
//	// generated: (inside a converter in mod)
//	...
//	out.ID = strconv.Itoa(in.ID)
//	...
//
// Multiple functions with the same signature cannot be registered. For
// error-returning conversions, use [ImportFuncErr].
func ImportFunc[In, Out any](fn func(In) Out) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ImportFuncErr is the error-returning variant of [ImportFunc]. It registers a
// custom conversion function (func(In) (Out, error)) with the module.
// Converters within the module may call this function when converting fields
// that can fail.
//
//	// source:
//	var mod = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))
//
//	// generated (inside a converter in mod):
//	// ...
//	out.ID, err = strconv.Atoi(in.ID)
//	// ...
//
// Multiple functions with the same signature cannot be registered.
func ImportFuncErr[In, Out any](fn func(In) (Out, error)) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ImportErrWrap appends an error wrapper function (func(error) error) to the
// module. An error wrapper is typically used to annotate errors with additional
// context, such as stack traces or error codes.
//
//	// source:
//	var mod = convgen.Module(convgen.ImportErrWrap(errtrace.Wrap))
//
//	// generated (inside a converter in mod):
//	// ...
//	if err != nil {
//		err = errtrace.Wrap(err)
//		return
//	}
//	// ...
//
// Multiple error wrappers can be registered. They are applied in the order they
// are registered. To remove all registered wrappers, use [ImportErrWrapReset].
func ImportErrWrap(fn func(error) error) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ImportErrWrapReset clears all error wrappers previously registered by
// [ImportErrWrap].
func ImportErrWrapReset() Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// RenameReplace is a renaming option that registers a renaming rule that
// replaces old with new for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameReplace(inOld, inNew, outOld, outNew string) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

func RenameReplaceRegexp(inRegexp, inRepl, outRegexp, outRepl string) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameToLower is a renaming option that registers a renaming rule that
// converts to lowercase for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameToLower(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameToUppser is a renaming option that registers a renaming rule that
// converts to uppercase for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameToUpper(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimPrefix is a renaming option that registers a renaming rule that
// trims a prefix for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameTrimPrefix(inPrefix, outPrefix string) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimSuffix is a renaming option that registers a renaming rule that
// trims a suffix for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameTrimSuffix(inSuffix, outSuffix string) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimCommonPrefix is a renaming option that registers a renaming rule
// trims the longest common prefix for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameTrimCommonPrefix(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimCommonSuffix is a renaming option that registers a renaming rule
// trims the longest common suffix for matching names.
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr],
// [Enum], and [EnumErr].
func RenameTrimCommonSuffix(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimCommonWordPrefix is a renaming option that registers a renaming rule that
// trims the longest common prefix for matching names based on word boundaries
func RenameTrimCommonWordPrefix(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameTrimCommonWordSuffix is a renaming option that registers a renaming rule that
// trims the longest common suffix for matching names based on word boundaries.
func RenameTrimCommonWordSuffix(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// RenameReset is a renaming option that clears all the renaming rules
// registered so far.
func RenameReset(inCancel, outCancel bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// Path parameters indicate a specific struct field, nested field, union
// implementation, or enum member. Struct fields can be indicated by using
// struct literals. Nested fields can be indicated by chaining field selections:
//
//	User{}
//	User{}.ID
//	User{}.Address.City
//
// The pointer type also can be indicated:
//
//	(&User{}).SetName
//	(*User)(nil).SetName
type Path = any

// Match manually maps a specific pair.
//
// For [Struct] and [StructErr], it specifies fields:
//
//	// source:
//	convgen.Struct[User, api.User](nil,
//		convgen.Match(User{}.Name, api.User{}.Username),
//	)
//
// For [Union] and [UnionErr], it specifies implementations:
//
//	// source:
//	convgen.Union[Event, api.Event](nil,
//		convgen.Match(ClickEvent{}, &api.ClickEvt{}),
//	)
//
// For [Enum] and [EnumErr], it specifies enum members:
//
//	// source:
//	convgen.Enum[Status, api.Status](nil, api.STATUS_UNSPECIFIED,
//		convgen.Match(StatusActive, api.STATUS_ACTIVE),
//	)
func Match(inPath, outPath Path) Option[no, no, yes, yes, yes] {
	panic("convgen: not generated")
}

// MatchFunc is a variant of [Match] that specifies a custom conversion function
// for the pair:
//
//	// source:
//	convgen.Struct[User, api.User](mod,
//		convgen.MatchFunc(User{}.Name, api.User{}.DisplayName, renderName),
//	)
//
//	// generated (simplified):
//	func convUser(in User) (out api.User) {
//		out.DisplayName = renderName(in.Name)
//		return
//	}
//
// To use a function that returns an error, use [MatchFuncErr] instead.
func MatchFunc[In, Out Path](inPath In, outPath Out, fn func(In) Out) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// MatchFuncErr is the error-returning variant of [MatchFunc]. It specifies a
// custom conversion function for the given pair.
func MatchFuncErr[In, Out Path](inPath In, outPath Out, fn func(In) (Out, error)) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// MatchSkip skips a specific pair so that Convgen does not attempt to match
// them automatically. The pair must otherwise be matchable by Convgen;
// otherwise, it reports an error at generation time.
//
//	// source:
//	convgen.Struct[User, api.User](nil,
//		convgen.MatchSkip(User{}.PasswordHash, api.User{}.PasswordHash),
//	)
//
// A common use case is to ignore missing conversions using nil:
//
//	// source:
//	convgen.Struct[User, api.User](nil,
//		convgen.MatchSkip(nil, api.User{}.LastAPIVersion), // ignore missing conversion
//	)
func MatchSkip(inPath, outPath Path) Option[no, no, yes, yes, yes] {
	panic("convgen: not generated")
}

// DiscoverBySample enables Convgen to discover matching items from the package
// of the given sample value.
//
// When implementations or constants are declared in a package different from
// where the type itself is defined, this option allows Convgen to locate and
// match them. When enabled, Convgen ignores items declared in the package of
// the type and instead searches in the package of the sample value:
//
//	// source:
//	var convAnimal = convgen.Union[Animal, api.Animal](mod,
//		convgen.DiscoverBySample(impls.Cat{}, nil), // discover Animal implementations in impls package
//	)
//
// At least one argument must be non-nil; a nil argument means to keep the
// default discovery behavior for that corresponding type.
//
// When this option is specified multiple times, the last one takes effect.
func DiscoverBySample(inSample, outSample Path) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// DiscoverUnexported enables discovery of unexported fields, implementations,
// or enum members when the type is defined in the same package as the
// converter:
//
//	type A struct{ n int }
//	type B struct{ n int }
//
//	// source:
//	var AtoB = convgen.Struct[A, B](nil,
//		convgen.DiscoverUnexported(true, true),
//	)
//
//	// generated (simplified):
//	func AtoB(in A) (out B) {
//		out.n = in.n
//		return
//	}
//
// Passing false disables discovery of unexported items, allowing previous
// settings to be overridden:
//
//	var mod = convgen.Module(convgen.DiscoverUnexported(true, true))
//	var conv1 = convgen.Struct[A, B](mod) // discovers unexported fields of A and B
//	var conv2 = convgen.Struct[C, D](mod,
//		convgen.DiscoverUnexported(false, false), // disables unexported discovery
//	)
//
// When this option is specified multiple times, the last one takes effect.
func DiscoverUnexported(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// DiscoverGetters enables discovery of getter methods for reading fields of an
// input struct. A getter method has one of the following forms:
//
//	func (T) PrefixFieldNameSuffix() FieldType
//	func (T) PrefixFieldNameSuffix() (FieldType, error)
//
// The prefix and suffix parameter control how getter names are formed, and the
// empty string is allowed for either. When matching to output fields, Convgen
// trims the prefix and suffix from the method name:
//
//	// source:
//	type User struct{ name string }
//	func (u User) GetName() string { return u.name }
//	var convUser = convgen.Struct[User, api.User](nil,
//		convgen.DiscoverGetters("Get", ""),
//	)
//
//	// generated (simplified):
//	func convUser(in User) (out api.User) {
//		out.Name = in.GetName()
//		return
//	}
//
// This option can also be set at the module level to apply to all struct
// converters within the module, which is useful when most structs follow a
// getter naming convention:
//
//	// source:
//	var mod = convgen.Module(convgen.DiscoverGetters("Get", ""))
//	var convUser = convgen.Struct[User, api.User](mod)
//
// In that case, use [DiscoverFieldsOnly] to disable getter discovery for
// specific struct converters that do not have getters.
//
// When this option is specified multiple times, the last one takes effect.
func DiscoverGetters(prefix, suffix string) Option[yes, yes, yes, no, no] {
	panic("convgen: not generated")
}

// DiscoverSetters enables discovery of setter methods for writing fields of an
// output struct. A setter method has one of the following forms:
//
//	func (*T) PrefixFieldNameSuffix(v FieldType)
//	func (*T) PrefixFieldNameSuffix(v FieldType) error
//
// The prefix and suffix parameter control how setter names are formed, and the
// empty string is allowed for either. When matching to input fields, Convgen
// trims the prefix and suffix from the method name:
//
//	// source:
//	type User struct{ name string }
//	func (u *User) SetName(v string) { u.name = v }
//	var convUser = convgen.Struct[api.User, User](nil,
//		convgen.DiscoverSetters("Set", ""),
//	)
//
//	// generated (simplified):
//	func convUser(in api.User) (out User) {
//		out.SetName(in.Name)
//		return
//	}
//
// This option can also be set at the module level to apply to all struct
// converters within the module, which is useful when most structs follow a
// setter naming convention:
//
//	// source:
//	var mod = convgen.Module(convgen.DiscoverSetters("Set", ""))
//	var convUser = convgen.Struct[api.User, User](mod)
//
// In that case, use [DiscoverFieldsOnly] to disable setter discovery for
// specific struct converters that do not have setters.
//
// When this option is specified multiple times, the last one takes effect.
func DiscoverSetters(prefix, suffix string) Option[yes, yes, yes, no, no] {
	panic("convgen: not generated")
}

// DiscoverFieldsOnly disables previously registered [DiscoverGetters] and
// [DiscoverSetters] options so that Convgen discovers only struct fields.
//
// Because getter and setter discovery can be enabled at the module level, this
// option is useful for overriding them for specific struct converters.
//
//	// source:
//	var (
//		mod      = convgen.Module(convgen.DiscoverGetters("Get", ""))
//		convUser = convgen.Struct[User, api.User](mod)
//	)
//
//	// Address does not have getters unlike other structs
//	var convAddress = convgen.Struct[Address, api.Address](mod,
//		convgen.DiscoverFieldsOnly(true, false),
//	)
//
// When this option is specified multiple times, the last one takes effect.
func DiscoverFieldsOnly(inEnable, outEnable bool) Option[yes, yes, yes, no, no] {
	panic("convgen: not generated")
}

// DiscoverNested is a converter-level option which ...
//
//	convgen.Struct[Person, FlatPerson](mod,
//		convgen.DiscoverNested(Person{}.Address, FlatPerson{}),
//	)
//
// Accepted by [Struct] and [StructErr].
func DiscoverNested(inPath, outPath Path) Option[no, no, yes, no, no] {
	panic("convgen: not generated")
}

// FieldGetter wraps an errorless getter function (func() y) in [MatchFunc] or
// [MatchFuncErr]. This helps to resolve type errors.
//
//	convgen.Struct[X, Y](mod,
//		convgen.WireMapFunc(convgen.FieldGetter(X{}.Name), Y{}.Name, rename),
//	)
func FieldGetter[In any](fn func() In) In { return *new(In) }

// FieldGetterErr wraps an errorful getter function (func() (y, error)) in
// [MatchFunc] or [MatchFuncErr]. This helps to resolve type errors.
//
//	convgen.StructErr[X, Y](mod,
//		convgen.WireMapFunc(convgen.FieldGetterErr(X{}.Name), Y{}.Name, rename),
//	)
func FieldGetterErr[In any](fn func() (In, error)) In { return *new(In) }

// FieldSetter wraps an errorless setter function (func(x)) in [MatchFunc] or
// [MatchFuncErr]. This helps to resolve type errors.
//
//	convgen.Struct[X, Y](mod,
//		convgen.WireMapFunc(X{}.Name, convgen.FieldSetter((*Y)(nil).SetName), rename),
//	)
func FieldSetter[Out any](fn func(Out)) Out { return *new(Out) }

// FieldSetterErr wraps an errorful setter function (func(x) error) in [MatchFunc]
// or [MatchFuncErr]. This helps to resolve type errors.
//
//	convgen.StructErr[X, Y](mod,
//		convgen.WireMapFunc(X{}.Name, convgen.FieldSetterErr((*Y)(nil).SetName), rename),
//	)
func FieldSetterErr[Out any](fn func(Out) error) Out { return *new(Out) }
