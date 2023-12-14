// Copyright (c) Gabriel de Quadros Ligneul
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

// This package contains the main that executes the nonodo command.
package main

import (
	"os"

	"github.com/gligneul/nonodo/cmd/nonodo/cmd"
)

func main() {
	err := cmd.Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
