module github.com/sublee/convgen/cmd/golangci-lint-convgen

go 1.25.1

require (
	github.com/golangci/plugin-module-register v0.1.2
	github.com/sublee/convgen v0.0.0
	golang.org/x/tools v0.32.0
)

replace github.com/sublee/convgen => ../..

require (
	github.com/emirpasic/gods v1.18.1 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)
