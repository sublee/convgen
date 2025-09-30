package convgenanalysis

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	convgeninternal "github.com/sublee/convgen/internal/convgen"
)

// Analyzer validates the usage of Convgen in the package.
var Analyzer = &analysis.Analyzer{
	Name: "convgen",
	Doc:  "linter for convgen usage",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkg := &packages.Package{
		Name:      pass.Pkg.Name(),
		PkgPath:   pass.Pkg.Path(),
		Types:     pass.Pkg,
		Fset:      pass.Fset,
		Syntax:    pass.Files,
		TypesInfo: pass.TypesInfo,
	}

	cg, err := convgeninternal.New(pkg)
	if err != nil {
		return nil, err
	}

	if err := cg.Build(); err != nil {
		// Unroll all errors and report them
		errs := []error{err}
		for len(errs) != 0 {
			err := errs[0]
			errs = errs[1:]

			if codeErr, ok := err.(*codefmt.CodeError); ok {
				pass.Report(analysis.Diagnostic{
					Pos:     codeErr.Pos(),
					End:     codeErr.End(),
					Message: codeErr.Unwrap().Error(),
				})
				continue
			}

			if u, ok := err.(interface{ Unwrap() []error }); ok {
				errs = append(errs, u.Unwrap()...)
			}
		}
	}

	return nil, nil
}
