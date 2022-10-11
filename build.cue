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

		test: {
			"1.17": _
			"1.18": _
			"1.19": _

			[v=string]: go.#Test & {
				source:  _code.output
				package: "./..."
			}
		}
	}
}
