export DOCKER_BUILDKIT := "1"

pwd := justfile_directory()

@list:
    just --list

golangci-lint_version := "1.51"
lint:
    docker run \
        -v {{ pwd }}:/src:ro \
        -w /src \
        golangci/golangci-lint:v{{ golangci-lint_version }} golangci-lint run

go_version := env_var_or_default("GO_VERSION", "1.20")

test:
    docker run \
        -v {{ pwd }}:/src:ro \
        -w /src \
        -e CGO_ENABLED=1 \
        golang:{{ go_version }} \
        go test -v -race ./...

inttest:
    just inttest/all