export DOCKER_BUILDKIT := "1"

pwd := justfile_directory()

@list:
    just --list

golangci_version := "1.51"
go_version := env_var_or_default("GO_VERSION", "1.20")

test *args:
    docker buildx build \
        --build-arg GO_VER={{ go_version }} \
        --output out/test \
        --target unittest-coverage . \
        {{ args }}
    cat out/test/test.stdout

lint *args:
    @docker buildx build \
        --build-arg GO_VER={{ go_version }} \
        --build-arg GOLANGCI_VER={{ golangci_version }} \
        --target lint . \
        {{ args }}

inttest:
    just inttest/all

clean:
    @rm -rf ./out