package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"
	"universe.dagger.io/go"
	"universe.dagger.io/alpha/go/golangci"
)

dagger.#Plan & {
	actions: {
		_code: core.#Source & {
			path: "."
		}

		lint: golangci.#Lint & {
			source: _code.output
		}

		test: go.#Test & {
			source:  _code.output
			package: "./..."
		}
	}
}
