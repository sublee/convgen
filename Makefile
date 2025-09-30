.PHONY: test cover lint golangci-lint-convgen

test:
	go test

cover:
	go test -cover -coverpkg .,./internal/... -coverprofile cover.prof
	go tool cover -html=cover.prof -o cover.html

lint:
	golangci-lint run . ./internal/... ./testdata/...

golangci-lint-convgen:
	golangci-lint custom
