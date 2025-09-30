package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/sys/unix"

	convgeninternal "github.com/sublee/convgen/internal/convgen"
)

var Version = "dev"

var (
	bFlag = flag.String("b", "", "comma-separated build tags")
	tFlag = flag.Bool("t", false, "include tests")
	oFlag = flag.String("o", "convgen_gen.go", "output file name")
	cFlag = flag.String("c", "auto", "colorize (auto|always|never)")
)

func init() {
	convgeninternal.Version = Version
}

func main() {
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	color := false
	switch *cFlag {
	case "auto":
		color = isatty()
	case "always":
		color = true
	case "never":
		color = false
	default:
		fmt.Fprintln(os.Stderr, "invalid -c value:", *cFlag)
		os.Exit(1)
	}

	outs, err := convgeninternal.Main(context.Background(), wd, os.Environ(), *bFlag, *tFlag, *oFlag, flag.Args())
	if err != nil {
		message := err.Error()
		if color {
			message = colorize(message)
		}
		fmt.Fprintln(os.Stderr, message)
		os.Exit(1)
	}

	for out, code := range outs {
		if err := os.WriteFile(out, code, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if relOut, err := filepath.Rel(wd, out); err == nil {
			out = relOut
		}
		fmt.Println("Generated:", out)
	}
}

// isatty reports whether the program is running in a terminal. If it is true,
// we can use ANSI color codes.
func isatty() bool {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	return err == nil
}

var (
	reTab  = regexp.MustCompile(`(?m)^\t.+`)
	reFail = regexp.MustCompile(`^\tFAIL:.+`)
)

// colorize adds ANSI color codes to the message.
func colorize(message string) string {
	const (
		red   = "\033[31m"
		dim   = "\033[2m"
		reset = "\033[0m"
	)
	m := []byte(message)
	m = reTab.ReplaceAllFunc(m, func(b []byte) []byte {
		if reFail.Match(b) {
			return []byte(red + string(b) + reset)
		}
		return []byte(dim + string(b) + reset)
	})
	return string(m)
}
