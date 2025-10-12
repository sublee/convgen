.PHONY: lint test cover golangci-lint-convgen

lint:
	golangci-lint run

test:
	go test ./...

cover:
	go test -cover -coverpkg .,./internal/... -coverprofile cover.prof
	go tool cover -html=cover.prof -o cover.html

golangci-lint-convgen:
	golangci-lint custom
