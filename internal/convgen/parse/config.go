package parse

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/types/typeutil"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/lcs"
	"github.com/sublee/convgen/internal/typeinfo"
)

type Path struct {
	StructField []types.Object
	UnionImpl   types.Type
	EnumMember  *types.Const
	Pos         token.Pos
}

func (p Path) IsValid() bool {
	return len(p.StructField) > 0 || p.UnionImpl != nil || p.EnumMember != nil
}

type Config struct {
	Funcs     []typeinfo.Func
	FuncExprs []ast.Expr
	ErrWraps  []typeinfo.Func

	RenamersX      []func(string, string) string
	RenamersY      []func(string, string) string
	CommonFindersX []func([]string) string
	CommonFindersY []func([]string) string

	Match      [][2]Path
	MatchAt    []token.Pos
	MatchFuncs map[[2]token.Pos]typeinfo.Func

	MatchSkip   [][2]Path
	MatchSkipAt []token.Pos

	DiscoverBySampleEnabled bool
	DiscoverBySamplePkgX    *types.Package
	DiscoverBySamplePkgY    *types.Package

	DiscoverUnexportedEnabled bool
	DiscoverUnexportedX       bool
	DiscoverUnexportedY       bool

	DiscoverGettersEnabled bool
	DiscoverGettersPrefix  string
	DiscoverGettersSuffix  string

	DiscoverSettersEnabled bool
	DiscoverSettersPrefix  string
	DiscoverSettersSuffix  string

	DiscoverNestedX []Path
	DiscoverNestedY []Path

	ForStruct *Config
	ForUnion  *Config
	ForEnum   *Config
}

func (cfg Config) Fork() Config {
	// Fork rename options to allow overriding
	cfg.RenamersX = slices.Clone(cfg.RenamersX)
	cfg.RenamersY = slices.Clone(cfg.RenamersY)
	cfg.CommonFindersX = slices.Clone(cfg.CommonFindersX)
	cfg.CommonFindersY = slices.Clone(cfg.CommonFindersY)

	// Reset match options
	cfg.Match = nil
	cfg.MatchFuncs = make(map[[2]token.Pos]typeinfo.Func)
	cfg.MatchAt = nil
	cfg.MatchSkip = nil
	cfg.MatchSkipAt = nil

	// Reset discover sample options
	cfg.DiscoverBySampleEnabled = false
	cfg.DiscoverBySamplePkgX = nil
	cfg.DiscoverBySamplePkgY = nil
	cfg.DiscoverNestedX = nil
	cfg.DiscoverNestedY = nil
	return cfg
}

func (cfg *Config) Update(other Config) {
	// Merge Import and Rename options
	cfg.Funcs = append(cfg.Funcs, other.Funcs...)
	cfg.FuncExprs = append(cfg.FuncExprs, other.FuncExprs...)
	cfg.ErrWraps = append(cfg.ErrWraps, other.ErrWraps...)

	cfg.RenamersX = append(cfg.RenamersX, other.RenamersX...)
	cfg.RenamersY = append(cfg.RenamersY, other.RenamersY...)
	cfg.CommonFindersX = append(cfg.CommonFindersX, other.CommonFindersX...)
	cfg.CommonFindersY = append(cfg.CommonFindersY, other.CommonFindersY...)

	// Follow Match options
	cfg.Match = slices.Clone(other.Match)
	cfg.MatchFuncs = maps.Clone(other.MatchFuncs)
	cfg.MatchSkip = slices.Clone(other.MatchSkip)
	cfg.MatchSkipAt = slices.Clone(other.MatchSkipAt)
	cfg.DiscoverNestedX = slices.Clone(other.DiscoverNestedX)

	// Follow Discover options if enabled
	if other.DiscoverBySampleEnabled {
		cfg.DiscoverBySampleEnabled = true
		cfg.DiscoverBySamplePkgX = other.DiscoverBySamplePkgX
		cfg.DiscoverBySamplePkgY = other.DiscoverBySamplePkgY
	}

	if other.DiscoverUnexportedEnabled {
		cfg.DiscoverUnexportedEnabled = true
		cfg.DiscoverUnexportedX = other.DiscoverUnexportedX
		cfg.DiscoverUnexportedY = other.DiscoverUnexportedY
	}

	if other.DiscoverGettersEnabled {
		cfg.DiscoverGettersEnabled = true
		cfg.DiscoverGettersPrefix = other.DiscoverGettersPrefix
		cfg.DiscoverGettersSuffix = other.DiscoverGettersSuffix
	}

	if other.DiscoverSettersEnabled {
		cfg.DiscoverSettersEnabled = true
		cfg.DiscoverSettersPrefix = other.DiscoverSettersPrefix
		cfg.DiscoverSettersSuffix = other.DiscoverSettersSuffix
	}
}

func (cfg Config) ForkForStruct() Config {
	c := cfg.Fork()
	if cfg.ForStruct != nil {
		c.Update(*cfg.ForStruct)
	}
	return c
}

func (cfg Config) ForkForUnion() Config {
	c := cfg.Fork()
	if cfg.ForUnion != nil {
		c.Update(*cfg.ForUnion)
	}
	return c
}

func (cfg Config) ForkForEnum() Config {
	c := cfg.Fork()
	if cfg.ForEnum != nil {
		c.Update(*cfg.ForEnum)
	}
	return c
}

type parsers interface {
	ParsePathX(p *Parser, expr ast.Expr) (*Path, error)
	ParsePathY(p *Parser, expr ast.Expr) (*Path, error)
	ValidatePath(p *Parser, path Path, pos token.Pos) error

	ParsePkgX(p *Parser, expr ast.Expr) (*types.Package, error)
	ParsePkgY(p *Parser, expr ast.Expr) (*types.Package, error)
}

func (p *Parser) ParseConfig(cfg *Config, args []ast.Expr, parsers parsers) error {
	var errs error
	for _, arg := range args {
		if _, ok := arg.(*ast.Ident); ok {
			err := codefmt.Errorf(p, arg, "option must be inlined, not assigned to variable")
			errs = errors.Join(errs, err)
			continue
		}

		call, ok := ast.Unparen(arg).(*ast.CallExpr)
		if !ok {
			// Probably, this case is unreachable because every option type is
			// unexported. The only way to create a valid option is to call a
			// option directive function, or assign it to a variable. The latter
			// one is caught above.
			err := codefmt.Errorf(p, arg, "cannot use %c as option", arg)
			errs = errors.Join(errs, err)
			continue
		}

		// Dispatch configuration qualifiers
		switch {
		case p.IsDirective(call, "ForStruct"):
			if cfg.ForStruct == nil {
				cfg.ForStruct = &Config{}
			}
			if err := p.ParseConfig(cfg.ForStruct, call.Args, parsers); err != nil {
				errs = errors.Join(errs, err)
			}
			continue
		case p.IsDirective(call, "ForUnion"):
			if cfg.ForUnion == nil {
				cfg.ForUnion = &Config{}
			}
			if err := p.ParseConfig(cfg.ForUnion, call.Args, parsers); err != nil {
				errs = errors.Join(errs, err)
			}
			continue
		case p.IsDirective(call, "ForEnum"):
			if cfg.ForEnum == nil {
				cfg.ForEnum = &Config{}
			}
			if err := p.ParseConfig(cfg.ForEnum, call.Args, parsers); err != nil {
				errs = errors.Join(errs, err)
			}
			continue
		}

		if err := p.ParseOption(cfg, call, parsers); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (p *Parser) ParseOption(cfg *Config, call *ast.CallExpr, ps parsers) error { // nolint: gocyclo
	callee := typeutil.Callee(p.Pkg().TypesInfo, call)
	if callee == nil || !IsConvgenImport(callee.Pkg().Path()) {
		return codefmt.Errorf(p, call, "option must be convgen directive")
	}

	name := callee.Name()
	switch name {
	case "ImportFunc":
		return p.ParseOptionImportFunc(cfg, call, false)
	case "ImportFuncErr":
		return p.ParseOptionImportFunc(cfg, call, true)
	case "ImportErrWrap":
		return p.ParseOptionImportErrWrap(cfg, call)
	case "ImportErrWrapReset":
		return p.ParseOptionImportErrWrapReset(cfg, call)

	case "RenameReplace":
		return p.ParseOptionRenameReplace(cfg, call)
	case "RenameReplaceRegexp":
		return p.ParseOptionRenameReplaceRegexp(cfg, call)
	case "RenameToLower":
		return p.ParseOptionRenameBool(cfg, call, strings.ToLower)
	case "RenameToUpper":
		return p.ParseOptionRenameBool(cfg, call, strings.ToUpper)
	case "RenameTrimPrefix":
		return p.ParseOptionRenameString(cfg, call, strings.TrimPrefix)
	case "RenameTrimSuffix":
		return p.ParseOptionRenameString(cfg, call, strings.TrimSuffix)
	case "RenameTrimCommonPrefix":
		return p.ParseOptionRenameCommon(cfg, call, lcs.CommonPrefix, strings.TrimPrefix)
	case "RenameTrimCommonSuffix":
		return p.ParseOptionRenameCommon(cfg, call, lcs.CommonSuffix, strings.TrimSuffix)
	case "RenameTrimCommonWordPrefix":
		return p.ParseOptionRenameCommon(cfg, call, lcs.CommonWordPrefix, strings.TrimPrefix)
	case "RenameTrimCommonWordSuffix":
		return p.ParseOptionRenameCommon(cfg, call, lcs.CommonWordSuffix, strings.TrimSuffix)
	case "RenameReset":
		return p.ParseOptionRenameReset(cfg, call)

	case "Match":
		return p.ParseOptionMatch(cfg, call, ps, false, false)
	case "MatchFunc":
		return p.ParseOptionMatch(cfg, call, ps, true, false)
	case "MatchFuncErr":
		return p.ParseOptionMatch(cfg, call, ps, true, true)
	case "MatchSkip":
		return p.ParseOptionMatchSkip(cfg, call, ps)

	case "DiscoverBySample":
		return p.ParseOptionDiscoverBySample(cfg, call, ps)
	case "DiscoverUnexported":
		return p.ParseOptionDiscoverUnexported(cfg, call)
	case "DiscoverGetters":
		return p.ParseOptionDiscoverGetters(cfg, call)
	case "DiscoverSetters":
		return p.ParseOptionDiscoverSetters(cfg, call)
	case "DiscoverFieldsOnly":
		return p.ParseOptionDiscoverFieldsOnly(cfg, call)
	case "DiscoverNested":
		return p.ParseOptionDiscoverNested(cfg, call, ps)
	}

	return codefmt.Errorf(p, call.Fun, "%s is not supported option", name)
}

func (p *Parser) ParseOptionImportFunc(c *Config, call *ast.CallExpr, hasErr bool) error {
	expr, err := needArgs1(p, call)
	if err != nil {
		return err
	}

	fn, err := p.ParseFunc(expr, hasErr)
	if err != nil {
		return err
	}

	c.Funcs = append(c.Funcs, fn.WithPos(call.Pos()))
	c.FuncExprs = append(c.FuncExprs, call)
	return nil
}

func (p *Parser) ParseOptionImportErrWrap(c *Config, call *ast.CallExpr) error {
	expr, err := needArgs1(p, call)
	if err != nil {
		return err
	}

	fn, err := p.ParseErrWrap(expr)
	if err != nil {
		return err
	}

	c.ErrWraps = append(c.ErrWraps, fn.WithPos(call.Pos()))
	return nil
}

func (p *Parser) ParseOptionImportErrWrapReset(c *Config, call *ast.CallExpr) error {
	err := needArgs0(p, call)
	if err != nil {
		return err
	}

	c.ErrWraps = nil
	return nil
}

func (p *Parser) ParseOptionRenameBool(c *Config, call *ast.CallExpr, rename func(string) string) error {
	x, y, err := parseArgs2[bool, bool](p, call)
	if err != nil {
		return err
	}

	if !x && !y {
		return codefmt.Errorf(p, call, "at least one parameter must be true")
	}

	if x {
		c.RenamersX = append(c.RenamersX, func(s, _ string) string { return rename(s) })
		c.CommonFindersX = append(c.CommonFindersX, nil)
	}
	if y {
		c.RenamersY = append(c.RenamersY, func(s, _ string) string { return rename(s) })
		c.CommonFindersY = append(c.CommonFindersY, nil)
	}
	return nil
}

func (p *Parser) ParseOptionRenameString(c *Config, call *ast.CallExpr, rename func(string, string) string) error {
	x, y, err := parseArgs2[string, string](p, call)
	if err != nil {
		return err
	}

	if x != "" {
		c.RenamersX = append(c.RenamersX, func(s, _ string) string { return rename(s, x) })
		c.CommonFindersX = append(c.CommonFindersX, nil)
	}
	if y != "" {
		c.RenamersY = append(c.RenamersY, func(s, _ string) string { return rename(s, y) })
		c.CommonFindersY = append(c.CommonFindersY, nil)
	}
	return nil
}

func (p *Parser) ParseOptionRenameCommon(c *Config, call *ast.CallExpr, find func([]string) string, rename func(string, string) string) error {
	x, y, err := parseArgs2[bool, bool](p, call)
	if err != nil {
		return err
	}

	if x {
		c.RenamersX = append(c.RenamersX, rename)
		c.CommonFindersX = append(c.CommonFindersX, find)
	}
	if y {
		c.RenamersY = append(c.RenamersY, rename)
		c.CommonFindersY = append(c.CommonFindersY, find)
	}
	return nil
}

func (p *Parser) ParseOptionRenameReplace(c *Config, call *ast.CallExpr) error {
	oldX, newX, oldY, newY, err := parseArgs4[string, string, string, string](p, call)
	if err != nil {
		return err
	}

	c.RenamersX = append(c.RenamersX, func(s, _ string) string { return strings.ReplaceAll(s, oldX, newX) })
	c.RenamersY = append(c.RenamersY, func(s, _ string) string { return strings.ReplaceAll(s, oldY, newY) })
	c.CommonFindersX = append(c.CommonFindersX, nil)
	c.CommonFindersY = append(c.CommonFindersY, nil)
	return nil
}

func (p *Parser) ParseOptionRenameReplaceRegexp(c *Config, call *ast.CallExpr) error {
	regexpPatternX, replX, regexpPatternY, replY, err := parseArgs4[string, string, string, string](p, call)
	if err != nil {
		return err
	}

	var errs error
	regexpX, err := regexp.Compile(regexpPatternX)
	if err != nil {
		errs = errors.Join(errs, codefmt.Errorf(p, call, "invalid regexp pattern: %s", regexpPatternX))
	}
	regexpY, err := regexp.Compile(regexpPatternY)
	if err != nil {
		errs = errors.Join(errs, codefmt.Errorf(p, call, "invalid regexp pattern: %s", regexpPatternY))
	}
	if errs != nil {
		return errs
	}

	c.RenamersX = append(c.RenamersX, func(s, _ string) string { return regexpX.ReplaceAllString(s, replX) })
	c.RenamersY = append(c.RenamersY, func(s, _ string) string { return regexpY.ReplaceAllString(s, replY) })
	c.CommonFindersX = append(c.CommonFindersX, nil)
	c.CommonFindersY = append(c.CommonFindersY, nil)
	return nil
}

func (p *Parser) ParseOptionRenameReset(c *Config, call *ast.CallExpr) error {
	x, y, err := parseArgs2[bool, bool](p, call)
	if err != nil {
		return err
	}

	if x {
		c.RenamersX = nil
		c.CommonFindersX = nil
	}
	if y {
		c.RenamersY = nil
		c.CommonFindersY = nil
	}
	return nil
}

func (p *Parser) ParseOptionDiscoverUnexported(c *Config, call *ast.CallExpr) error {
	x, y, err := parseArgs2[bool, bool](p, call)
	if err != nil {
		return err
	}

	c.DiscoverUnexportedEnabled = true
	c.DiscoverUnexportedX = x
	c.DiscoverUnexportedY = y
	return nil
}

func (p *Parser) ParseOptionDiscoverGetters(c *Config, call *ast.CallExpr) error {
	prefix, suffix, err := parseArgs2[string, string](p, call)
	if err != nil {
		return err
	}

	c.DiscoverGettersEnabled = true
	c.DiscoverGettersPrefix = prefix
	c.DiscoverGettersSuffix = suffix
	return nil
}

func (p *Parser) ParseOptionDiscoverSetters(c *Config, call *ast.CallExpr) error {
	prefix, suffix, err := parseArgs2[string, string](p, call)
	if err != nil {
		return err
	}

	c.DiscoverSettersEnabled = true
	c.DiscoverSettersPrefix = prefix
	c.DiscoverSettersSuffix = suffix
	return nil
}

func (p *Parser) ParseOptionDiscoverFieldsOnly(c *Config, call *ast.CallExpr) error {
	x, y, err := parseArgs2[bool, bool](p, call)
	if err != nil {
		return err
	}

	if !x && !y {
		return codefmt.Errorf(p, call, "at least one parameter must be true")
	}

	if x {
		c.DiscoverGettersEnabled = false
		c.DiscoverGettersPrefix = ""
		c.DiscoverGettersSuffix = ""
	}
	if y {
		c.DiscoverSettersEnabled = false
		c.DiscoverSettersPrefix = ""
		c.DiscoverSettersSuffix = ""
	}
	return nil
}

func (p *Parser) ParseOptionMatch(c *Config, call *ast.CallExpr, ps parsers, withFn bool, hasErr bool) error {
	var elemX, elemY, fnExpr ast.Expr
	if withFn {
		var err error
		elemX, elemY, fnExpr, err = needArgs3(p, call)
		if err != nil {
			return err
		}
	} else {
		var err error
		elemX, elemY, err = needArgs2(p, call)
		if err != nil {
			return err
		}
	}

	if p.IsNil(elemX) {
		return codefmt.Errorf(p, elemX, "cannot use nil for %c", call.Fun)
	}
	if p.IsNil(elemY) {
		return codefmt.Errorf(p, elemY, "cannot use nil for %c", call.Fun)
	}

	var errs error
	pathX, err := ps.ParsePathX(p, elemX)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	pathY, err := ps.ParsePathY(p, elemY)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	var fn typeinfo.Func
	if withFn {
		fn, err = p.ParseFunc(fnExpr, hasErr)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return errs
	}

	if err := ps.ValidatePath(p, *pathX, elemX.Pos()); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := ps.ValidatePath(p, *pathY, elemY.Pos()); err != nil {
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return errs
	}

	c.Match = append(c.Match, [2]Path{*pathX, *pathY})
	c.MatchAt = append(c.MatchAt, call.Pos())
	if fn != nil {
		c.MatchFuncs[[2]token.Pos{pathX.Pos, pathY.Pos}] = fn
	}
	return nil
}

func (p *Parser) ParseOptionMatchSkip(c *Config, call *ast.CallExpr, ps parsers) error {
	elemX, elemY, err := needArgs2(p, call)
	if err != nil {
		return err
	}

	nilX := p.IsNil(elemX)
	nilY := p.IsNil(elemY)

	if nilX && nilY {
		return codefmt.Errorf(p, call, "cannot use nil for both parameters")
	}

	var errs error
	var pathX, pathY Path

	if !nilX {
		pathPtrX, err := ps.ParsePathX(p, elemX)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			pathX = *pathPtrX
		}
	}
	if !nilY {
		pathPtrY, err := ps.ParsePathY(p, elemY)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			pathY = *pathPtrY
		}
	}
	if errs != nil {
		return errs
	}

	if err := ps.ValidatePath(p, pathX, elemX.Pos()); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := ps.ValidatePath(p, pathY, elemY.Pos()); err != nil {
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return errs
	}

	c.MatchSkip = append(c.MatchSkip, [2]Path{pathX, pathY})
	c.MatchSkipAt = append(c.MatchSkipAt, call.Pos())
	return nil
}

func (p *Parser) ParseOptionDiscoverBySample(c *Config, call *ast.CallExpr, ps parsers) error {
	elemX, elemY, err := needArgs2(p, call)
	if err != nil {
		return err
	}

	var errs error
	if c.DiscoverBySampleEnabled {
		errs = errors.Join(errs, codefmt.Errorf(p, call, "convgen.DiscoverBySample already configured"))
	}

	var pkgX, pkgY *types.Package
	if !p.IsNil(elemX) {
		pkg, err := ps.ParsePkgX(p, elemX)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			pkgX = pkg
		}
	}
	if !p.IsNil(elemY) {
		pkg, err := ps.ParsePkgY(p, elemY)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			pkgY = pkg
		}
	}
	if errs != nil {
		return errs
	}

	if pkgX == nil && pkgY == nil {
		return codefmt.Errorf(p, call, "at least one parameter must be non-nil")
	}

	c.DiscoverBySampleEnabled = true
	if pkgX != nil {
		c.DiscoverBySamplePkgX = pkgX
	}
	if pkgY != nil {
		c.DiscoverBySamplePkgY = pkgY
	}
	return nil
}

func (p *Parser) ParseOptionDiscoverNested(c *Config, call *ast.CallExpr, ps parsers) error {
	elemX, elemY, err := needArgs2(p, call)
	if err != nil {
		return err
	}

	var errs error
	var pathX, pathY Path

	if !p.IsNil(elemX) {
		pathPtrX, err := ps.ParsePathX(p, elemX)
		pathX = *pathPtrX
		errs = errors.Join(errs, err)
	}
	if !p.IsNil(elemY) {
		pathPtrY, err := ps.ParsePathY(p, elemY)
		pathY = *pathPtrY
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return errs
	}

	if !pathX.IsValid() && !pathY.IsValid() {
		return codefmt.Errorf(p, call, "at least one parameter must be non-nil")
	}

	if pathX.IsValid() {
		if err := ps.ValidatePath(p, pathX, elemX.Pos()); err != nil {
			errs = errors.Join(errs, err)
		}
		if !typeinfo.TypeOf(pathX.StructField[len(pathX.StructField)-1].Type()).Deref().IsStruct() {
			err := codefmt.Errorf(p, elemX, "must be struct or *struct for nested discovery")
			errs = errors.Join(errs, err)
		}
	}
	if pathY.IsValid() {
		if err := ps.ValidatePath(p, pathY, elemY.Pos()); err != nil {
			errs = errors.Join(errs, err)
		}
		if !typeinfo.TypeOf(pathY.StructField[len(pathY.StructField)-1].Type()).Deref().IsStruct() {
			err := codefmt.Errorf(p, elemY, "must be struct or *struct for nested discovery")
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return errs
	}

	if pathX.IsValid() {
		c.DiscoverNestedX = append(c.DiscoverNestedX, pathX)
	}
	if pathY.IsValid() {
		c.DiscoverNestedY = append(c.DiscoverNestedY, pathY)
	}
	return nil
}
