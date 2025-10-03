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
//	var EncodeUser = convgen.Struct[model.User, pb.User](nil)
//
//	// generated: (simplified)
//	func EncodeUser(in model.User) (out pb.User) {
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
// diagnostics. For example, if our model.User has an ID field but pb.User has
// Id (with a lower case "d") instead, so they don't match exactly:
//
//	main.go:10:10: invalid match between model.User and pb.User
//		FAIL: ID -> ?  // missing
//		FAIL: ?  -> Id // missing
//
// Renaming rules can be applied to resolve those mismatches. In this case, we
// can solve with just [RenameToLower]. It renames model.User.ID and pb.User.Id
// both to become "id":
//
//	// source:
//	var EncodeUser = convgen.Struct[model.User, pb.User](nil,
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
//	var EncodeUser = convgen.Struct[model.User, pb.User](nil,
//		convgen.Match(model.User{}.ID, pb.User{}.Id),
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
//		EncodeUser = convgen.Struct[model.User, pb.User](enc)
//		EncodeRole = convgen.Enum[model.Role, pb.Role](enc, pb.Role_ROLE_UNSPECIFIED, convgen.RenameTrimCommonPrefix(true, true))
//	)
//
//	// generated: (simplified)
//	func EncodeUser(in model.User) (out pb.User) {
//		out.Name = in.Name
//		out.Email = in.Email
//		out.Role = EncodeRole(in.Role)
//		           ^^^^^^^^^^^^^^^^^^^ // reused converter in the same module
//		return
//	}
//	func EncodeRole(in model.Role) (out pb.Role) {
//		switch in {
//		case model.RoleAdmin:
//			return pb.Role_ROLE_ADMIN
//		case model.RoleMember:
//			return pb.Role_ROLE_MEMBER
//		default:
//			return pb.Role_ROLE_UNSPECIFIED
//		}
//	}
//
// Conversions are often more than field-to-field copies. If Convgen cannot
// determine how to convert a type to another type, it reports an error.
// However, a custom function can be used for any type conversion. A custom
// function can be imported into a module by [ImportFunc].
//
// For example, when model.User.ID is int but pb.User.ID is string, a func(int)
// string like strconv.Itoa can be imported to convert them:
//
//	// source:
//	var (
//		enc = convgen.Module(convgen.ImportFunc(strconv.Itoa))
//		EncodeUser = convgen.Struct[model.User, pb.User](enc)
//	)
//
//	// generated: (simplified)
//	func EncodeUser(in model.User) (out pb.User) {
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
// In the above example, we could encode model.User to pb.User without any
// error. But the reverse is not possible because string to int conversion
// (pb.User.ID to model.User.ID) may fail at runtime. So, we need to use the Err
// variants:
//
//	// source:
//	var (
//		dec = convgen.Module(convgen.ImportFuncErr(strconv.Atoi))
//		DecodeUser = convgen.StructErr[pb.User, model.User](dec)
//	)
//
//	// generated: (simplified)
//	func DecodeUser(in pb.User) (out model.User, err error) {
//	                                             ^^^^^^^^^ // may return error
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

func ForStruct(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

func ForUnion(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

func ForEnum(opts ...forOption) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// Struct directive generates a converter function between two struct types
// without error.
//
//	// source:
//	var convUser = convgen.Struct[User, pb.User](nil)
//
// The input and output types are declared as type parameters. The variable that
// holds the directive is rewritten to the actual function when Convgen
// generates code:
//
//	// generated: (simplified)
//	func convUser(in User) (out pb.User) {
//		out.Name = in.Name
//		out.Email = in.Email
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
// converter function that returns (Out, error) instead of just Out. Unlike
// [Struct], StructErr allows the generated converter to call other functions
// within the same [Module] that may themselves return an error.
func StructErr[In, Out any](mod module, opts ...structOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

// Union marks a converter function between two interface types. It finds
// implementations from the same package of each interface type or the sample
// implementation specified by [DiscoverBySample]. The converter functions for
// implementations have to be discoverable in the module.
func Union[In, Out any](mod module, opts ...unionOption) func(In) Out {
	panic("convgen: not generated")
}

func UnionErr[In, Out any](mod module, opts ...unionOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

func Enum[In, Out any](mod module, default_ Out, opts ...enumOption) func(In) Out {
	panic("convgen: not generated")
}

func EnumErr[In, Out any](mod module, default_ Out, opts ...enumOption) func(In) (Out, error) {
	panic("convgen: not generated")
}

type Option[Module, For, Struct, Union, Enum canUseFor] interface {
	moduleOption() Module
	forOption() For
	structOption() Struct
	unionOption() Union
	enumOption() Enum
}

// ImportFunc is an option which registers a custom errorless converter function
// (func(x) y) in the module. Converters in the module may use it to convert
// inner types.
//
//	mod := convgen.New(convgen.ImportFunc(strconv.Itoa))
//
// Accepted by [Module] only.
func ImportFunc[In, Out any](fn func(In) Out) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

// ImportFuncErr is a module-level option which registers a custom errorful
// converter function (func(x) (y, error)) in the module. Converters in the
// module may use it to convert inner types.
//
//	mod := convgen.New(convgen.ImportFuncErr(strconv.Atoi))
//
// Accepted by [Module] only.
func ImportFuncErr[In, Out any](fn func(In) (Out, error)) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

func ImportErrWrap(fn func(error) error) Option[yes, no, no, no, no] {
	panic("convgen: not generated")
}

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

// Path parameters indicate a specific type or struct field.
//
// A type itself, a field of a struct, or a nested field can be indicated as a
// Path:
//
//	Order{}
//	Order{}.ID
//	Order{}.Address.City
//
// The pointer type also can be indicated:
//
//	(*Order)(nil).SetID
type Path = any

// Match is a converter-level option which matches a pair manually.
//
//	convgen.Struct[X, Y](mod,
//		convgen.Match(X{}.ID, Y{}.Identifier),
//	)
//
// Accepted by [Struct], [StructErr], [Union], [UnionErr], [Enum], and
// [EnumErr].
func Match(inPath, outPath Path) Option[no, no, yes, yes, yes] {
	panic("convgen: not generated")
}

// MatchFunc is a converter-level option which matches a pair manually. It also
// specifies a custom errorless converter function for converting them.
//
//	convgen.Struct[X, Y](mod,
//		convgen.MatchFunc(X{}.Name, Y{}.DisplayName, renderName),
//	)
//
// Accepted by [Struct], [StructErr], [Union], and [UnionErr].
func MatchFunc[In, Out Path](inPath In, outPath Out, fn func(In) Out) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// MatchFuncErr is a converter-level option which matches a pair manually. It also
// specifies a custom errorful converter function for converting them.
//
//	convgen.StructErr[X, Y](mod,
//		convgen.MatchFuncErr(X{}.UUID, Y{}.ID, parseUUID),
//	)
//
// Accepted by [StructErr] and [UnionErr].
func MatchFuncErr[In, Out Path](inPath In, outPath Out, fn func(In) (Out, error)) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// MatchSkip is a converter-level option which marks a specific matched pair to
// be ignored. To suppress errors on missing items, mark them by this option
// with [Missing].
//
//	convgen.Struct[X, Y](mod,
//		convgen.MatchSkip(X{}.Metadata, Y{}.Metadata),
//		convgen.MatchSkip(convgen.Missing, Y{}.Extra),
//	)
//
// Accepted by [Struct], [StructErr], [Union], [UnionErr], [Enum], and
// [EnumErr].
func MatchSkip(inPath, outPath Path) Option[no, no, yes, yes, yes] {
	panic("convgen: not generated")
}

// DiscoverBySample is a converter-level option which specifies a sample of
// items. Convgen will look up other items from the same package of the given
// sample.
//
// Accepted by [Union], [UnionErr], [Enum], and [EnumErr].
func DiscoverBySample(inSample, outSample Path) Option[no, no, yes, yes, no] {
	panic("convgen: not generated")
}

// DiscoverUnexported enables to discover unexported fields, methods,
// implmentations, and values in the same package.
//
//	type A struct { n int }
//	type B struct { n int }
//
//	// Generate: b.n = a.n
//	convgen.Struct[A, B](mod, convgen.DiscoverUnexported(true))
//
// Accepted by [Module], [Struct], [StructErr], [Union], [UnionErr], [Enum], and
// [EnumErr].
func DiscoverUnexported(inEnable, outEnable bool) Option[yes, yes, yes, yes, yes] {
	panic("convgen: not generated")
}

// DiscoverGetters enables to match getter methods to access fields in a struct.
//
//	// Generate: b.ID = a.GetID()
//	convgen.Struct[A, B](mod, convgen.DiscoverGetters(true, "Get", ""))
//
// Accepted by [Module], [Struct], and [StructErr].
func DiscoverGetters(enable bool, prefix, suffix string) Option[yes, yes, yes, no, no] {
	panic("convgen: not generated")
}

// DiscoverSetters enables to match setter methods to set fields in a struct.
//
//	// Generate: b.SetID(a.ID)
//	convgen.Struct[A, B](mod, convgen.DiscoverSetters(true, "Set", ""))
//
// Accepted by [Module], [Struct], and [StructErr].
func DiscoverSetters(enable bool, prefix, suffix string) Option[yes, yes, yes, no, no] {
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
