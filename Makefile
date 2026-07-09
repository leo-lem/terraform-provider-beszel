.PHONY: build test testacc generate lint

build:
	go build ./...

test:
	go vet ./...
	go test ./...

# Requires: eval "$(./scripts/testacc-up.sh)" first, and ./scripts/testacc-down.sh after.
testacc:
	TF_ACC=1 go test ./... -v -timeout 20m

generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.0 generate --provider-name beszel

lint:
	test -z "$$(gofmt -l .)"
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run
