package convgeninternal

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var Version string

// Main is the main entry point for Convgen. It is used by the command-line tool
// directly.
//
// ctx is the context for loading packages. If the loading is too slow, ctx can
// cancel the operation. wd is the path of the working directory. env is the
// environment variables to use when running the tool. tags is the build tags to
// use when loading packages. tests indicates whether to include test files.
// outFile is the name of the output file to generate in each package. And
// patterns are the package patterns to process.
//
// It returns a map of output file paths to their contents. If any error occurs,
// it returns a non-nil error.
func Main(ctx context.Context, wd string, env []string, tags string, tests bool, outFile string, patterns []string) (map[string][]byte, error) {
	pkgs, err := load(ctx, wd, env, tags, tests, patterns)
	if err != nil {
		return nil, err
	}

	outs := make(map[string][]byte)
	var errs error

	for _, pkg := range pkgs {
		if len(pkg.Errors) != 0 {
			err := fmt.Errorf("pkg %q has errors", pkg.Name)
			errs = errors.Join(errs, err)
			continue
		}

		cg, err := New(pkg)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if err := cg.Build(); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		code := cg.Generate()
		if len(code) == 0 {
			continue
		}

		outDir := filepath.Dir(pkg.GoFiles[0])
		if rel, err := filepath.Rel(wd, outDir); err == nil {
			outDir = rel
		}
		out := filepath.Join(outDir, outFile)
		outs[out] = code
	}
	if errs != nil {
		// errs already contains comprehensive error messages. So we don't need
		// to attach another error message.
		return nil, reorderErrors(errs)
	}

	return outs, nil
}

// load loads packages.
func load(ctx context.Context, wd string, env []string, tags string, tests bool, patterns []string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedDeps | packages.NeedFiles | packages.NeedImports | packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Context:    ctx,
		Dir:        wd,
		Env:        env,
		BuildFlags: []string{"-tags=convgen"},
		Tests:      tests,
	}
	if tags != "" {
		cfg.BuildFlags[0] += "," + tags
	}

	// Load the packages based on the provided patterns.
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found: %v", patterns)
	}

	// Check for errors in the loaded packages.
	var errs error
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			if err.Pos == "" {
				errs = errors.Join(errs, errors.New(err.Msg))
				continue
			}

			path, rowcol, _ := strings.Cut(err.Pos, ":")
			if rel, relErr := filepath.Rel(wd, path); relErr == nil {
				err.Pos = rel + ":" + rowcol
			}
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return nil, errs
	}

	return pkgs, nil
}

func reorderErrors(errs error) error {
	if errs == nil {
		return nil
	}

	// Flatten nested errors
	list := []error{errs}
	for i := 0; i < len(list); i++ {
		if u, ok := list[i].(interface{ Unwrap() []error }); ok {
			// errors.Join collapses errors with a single error having Unwrap()
			// []error method. The underlying errors could be retrieved using
			// the Unwrap() method.
			list = append(list, u.Unwrap()...)

			// The underlying errors are appended to the list. So the original
			// error can be removed.
			list[i] = nil
			continue
		}
	}
	list = slices.DeleteFunc(list, func(err error) bool {
		return err == nil
	})

	// Sort errors by message
	sort.Slice(list, func(i, j int) bool {
		return list[i].Error() < list[j].Error()
	})
	return errors.Join(list...)
}
