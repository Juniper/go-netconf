ARG GO_VER=1.20
ARG GOLANGCI_VER=1.51

FROM golang:${GO_VER} as base
WORKDIR /src
COPY go.mod go.sum /src/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

FROM golangci/golangci-lint:v${GOLANGCI_VER}-alpine AS lint-base

FROM base AS lint
RUN --mount=target=. \
    --mount=from=lint-base,src=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run --timeout 10m0s ./...

FROM base as unittest
RUN --mount=target=/src \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir /out && \
    go test -v -race -coverprofile=/out/cover.out ./... | tee /out/test.stdout

FROM base AS inttest
RUN apt update && apt install -y \
    openssh-client \
    sshpass \
    libxml-xpath-perl
RUN --mount=target=/src \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir /out && \
    cd inttest && \
    go test --tags=inttest -c -o /out/inttest.test
WORKDIR /out
COPY inttest/wait-for-hello.sh .
CMD ./inttest.test -test.v -test.race


FROM scratch AS unittest-coverage
COPY --from=unittest /out /