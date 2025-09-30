// golangcilintconvgen package provides a plugin for golangci-lint to integrate
// the Convgen analyzer. To build a custom golangci-lint binary with this
// plugin, use the following command at this package's directory:
//
//	golangci-lint custom
//
// Now you will have a golangci-lint-convgen binary that you can use to lint
// your Go code with the Convgen analyzer.
package golangcilintconvgen

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/sublee/convgen/pkg/convgenanalysis"
)

func init() {
	register.Plugin("convgen", New)
}

func New(settings any) (register.LinterPlugin, error) {
	return ConvgenLinter{}, nil
}

type ConvgenLinter struct{}

func (ConvgenLinter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{convgenanalysis.Analyzer}, nil
}

func (ConvgenLinter) GetLoadMode() string {
	return register.LoadModeSyntax
}
