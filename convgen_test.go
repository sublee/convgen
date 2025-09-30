package convgen_test

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	convgeninternal "github.com/sublee/convgen/internal/convgen"
	"github.com/sublee/convgen/pkg/convgenanalysis"
)

// TestAnalysis tests parsing and building errors using the Go analysis
// protocol. In this test, Convgen errors will be reported as analysis errors.
// "// want `REGEXP`" comments in the fixture source files are used to check for
// expected analysis errors.
//
// The directory structure of testdata for subtests is as follows:
//
//	testdata/
//	└── analysis/
//	    ├── pkg1/
//	    │   └── *.go // with want comments
//	    └── pkg2/
//	        └── *.go // with want comments
func TestAnalysis(t *testing.T) {
	ents, err := os.ReadDir(filepath.FromSlash("testdata/analysis"))
	require.NoError(t, err)

	t.Setenv("GOFLAGS", "-tags=convgen")

	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}

		t.Run(ent.Name(), func(t *testing.T) {
			t.Parallel()

			defer func() {
				if t.Failed() {
					t.Logf("\n\tReproduce:\tgo run ./cmd/convgen ./testdata/analysis/%s", ent.Name())
				}
			}()

			analysistest.Run(t, "", convgenanalysis.Analyzer, "./testdata/analysis/"+ent.Name())
		})
	}
}

// TestPrograms tests programs in the testdata directory.
//
// The directory structure of testdata for subtests is as follows:
//
//	testdata/
//	└── program/
//	    ├── program1/
//	    │   ├── main_pkg.txt --- If main_pkg.txt is not present, "main" will be used as the default package name.
//	    │   ├── main/
//	    │   │   └── main.go
//	    │   └── want/
//	    │       └── program_output.txt
//	    └── program2/
//	        ├── main_pkg.txt
//	        ├── foo/
//	        │   └── foo.go
//	        ├── bar/
//	        │   └── bar.go
//	        └── want/
//	            └── convgen_error.txt
func TestPrograms(t *testing.T) {
	// NOTE: Code snippets were stolen from Wire.
	ents, err := os.ReadDir(filepath.FromSlash("testdata/program"))
	require.NoError(t, err)

	convgenGo, err := os.ReadFile("convgen.go")
	require.NoError(t, err)
	convgenErrorsGo, err := os.ReadFile(filepath.FromSlash("pkg/convgenerrors/errors.go"))
	require.NoError(t, err)

	var tests []*programTest
	for _, ent := range ents {
		name := ent.Name()
		if !ent.IsDir() || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}

		test, err := newProgramTest(name, convgenGo, convgenErrorsGo)
		if err != nil {
			t.Error(err)
			continue
		}

		tests = append(tests, test)
	}

	for _, test := range tests {
		t.Run(test.Name(), test.Test())
	}
}

// programTest is a test case for a program. It executes Convgen for the program
// and runs the program with generated code to check the output.
type programTest struct {
	name    string
	mainPkg string
	files   map[string][]byte
	want    struct {
		ProgramOutput string
		ConvgenError  string
	}
}

func (test *programTest) Name() string {
	return test.name
}

func (test *programTest) PkgPath() string {
	return fmt.Sprintf("example.com/%s", test.name)
}

func (test *programTest) ProgramPath() string {
	return fmt.Sprintf("%s/%s", test.PkgPath(), test.mainPkg)
}

// newProgramTest creates a new program test case.
func newProgramTest(name string, convgenGo, convgenErrorsGo []byte) (*programTest, error) {
	root := filepath.Join(filepath.FromSlash("testdata/program"), name)
	test := programTest{
		name:  name,
		files: make(map[string][]byte),
	}

	// mainPkg
	mainPkg, err := os.ReadFile(filepath.Join(root, "main_pkg.txt"))
	if errors.Is(err, os.ErrNotExist) {
		mainPkg = []byte("main")
	} else if err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}
	test.mainPkg = string(bytes.TrimSpace(mainPkg))

	// want
	programOutput, _ := os.ReadFile(filepath.Join(root, "want", "program_output.txt"))
	convgenError, _ := os.ReadFile(filepath.Join(root, "want", "convgen_error.txt"))
	test.want.ProgramOutput = string(bytes.TrimSpace(programOutput))
	test.want.ConvgenError = string(bytes.TrimSpace(convgenError))

	if test.want.ProgramOutput == "" && test.want.ConvgenError == "" {
		return nil, fmt.Errorf("load test case %s: does not want anything", name)
	}

	// files
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Bubble up I/O errors
			return err
		}

		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Skip directories
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			panic(err)
		}

		if !info.Mode().IsRegular() || filepath.Ext(path) != ".go" {
			// Skip non-Go files
			return nil
		}

		if filepath.Base(path) == "convgen_gen.go" {
			// Skip generated Convgen files, they might be existed for debugging
			// purposes.
			return nil
		}

		goCode, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		test.files[test.PkgPath()+"/"+filepath.ToSlash(rel)] = goCode
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}

	test.files["github.com/sublee/convgen/convgen.go"] = convgenGo
	test.files["github.com/sublee/convgen/pkg/convgenerrors/errors.go"] = convgenErrorsGo
	return &test, nil
}

// materialize copies the program code and convgen.go into the given GOPATH.
func (test *programTest) materialize(gopath string) error {
	// NOTE: Code snippets were stolen from Wire.
	for name, content := range test.files {
		dst := filepath.Join(gopath, "src", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0o777); err != nil {
			return fmt.Errorf("mkdir %s: %w", name, err)
		}
		if err := os.WriteFile(dst, content, 0o666); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	// Write go.mod file for github.com/sublee/convgen
	convgenGomodPath := filepath.Join(gopath, "src", "github.com", "sublee", "convgen", "go.mod")
	convgenGomod := `
	module github.com/sublee/convgen
	go 1.25.0`
	if err := os.WriteFile(convgenGomodPath, []byte(convgenGomod), 0o666); err != nil {
		return fmt.Errorf("write github.com/sublee/convgen/go.mod: %w", err)
	}

	// Write go.mod file for example.com/NAME
	testGomodPath := filepath.Join(gopath, "src", filepath.FromSlash(test.PkgPath()), "go.mod")
	testGomod := fmt.Sprintf(`
	module %s
	go 1.25.0
	require github.com/sublee/convgen v0.0.0
	replace github.com/sublee/convgen => %s
	`, test.PkgPath(), filepath.Join(gopath, filepath.FromSlash("src/github.com/sublee/convgen")))
	if err := os.WriteFile(testGomodPath, []byte(testGomod), 0o666); err != nil {
		return fmt.Errorf("write %s/go.mod: %w", test.PkgPath(), err)
	}

	return nil
}

// Test returns a test function for the program test. It runs Convgen for the
// program and then checks its error or output messages.
func (test *programTest) Test() func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		defer func() {
			if t.Failed() {
				t.Logf("\n\tReproduce:\tgo run ./cmd/convgen ./testdata/program/%s/%s", test.Name(), test.mainPkg)
			}
		}()

		// Materialize in a temporary directory
		// gopath := t.TempDir()
		gopath := os.TempDir() + "/convgen_test_" + test.Name()
		require.NoError(t, test.materialize(gopath), "Materialization failed")

		// Run Convgen
		wd := filepath.Join(gopath, "src", filepath.FromSlash(test.PkgPath()))
		env := append(os.Environ(), "GOPATH="+gopath)
		generated, convgenErr := convgeninternal.Main(t.Context(), wd, env, "", false, "convgen_gen.go", []string{"pattern=./" + test.mainPkg})

		// Check for the Convgen error
		if convgenErr != nil {
			convgenErr = errors.New(relPathInString(convgenErr.Error(), wd))
			if test.want.ConvgenError != "" {
				want := normalizeWhitespace(test.want.ConvgenError)
				have := normalizeWhitespace(convgenErr.Error())
				assert.Equal(t, want, have)
			} else {
				require.NoError(t, convgenErr, "Convgen exited with errors unexpectedly")
			}
			return
		}

		if test.want.ConvgenError != "" {
			require.Error(t, convgenErr, "Convgen should have exited with an error")
		}

		// Write generated files
		for name, content := range generated {
			err := os.WriteFile(filepath.Join(wd, name), content, 0o666)
			require.NoError(t, err, "Failed to write a generated file")
		}

		// Run the program
		goCmd := filepath.Join(build.Default.GOROOT, "bin", "go")
		cmd := exec.Command(goCmd, "run", test.ProgramPath())
		cmd.Dir = wd
		progOut, err := cmd.CombinedOutput()
		require.NoError(t, err, string(progOut))

		// Test
		if test.want.ProgramOutput != "" {
			assert.Equal(t, test.want.ProgramOutput, strings.TrimSpace(string(progOut)))
		}
	}
}

// relPathInString replaces paths in the given string to their relative paths to
// the new working directory.
func relPathInString(s, wd string) string {
	realWD, err := os.Getwd()
	if err != nil {
		return s
	}

	rel, err := filepath.Rel(realWD, wd)
	if err != nil {
		return s
	}

	s = strings.ReplaceAll(s, rel+"/", "")
	s = strings.ReplaceAll(s, rel, "")
	return s
}

// normalizeWhitespaces normalizes whitespace in the given string for consistent
// comparison regardless of whitespace style.
func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\t", "    ")
	return s
}
