// Copyright (c) Gabriel de Quadros Ligneul
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

// This package contains the nonodo run function.
// This is separate from the main package to facilitate testing.
package nonodo

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gligneul/nonodo/internal/devnet"
	"github.com/gligneul/nonodo/internal/echoapp"
	"github.com/gligneul/nonodo/internal/inputter"
	"github.com/gligneul/nonodo/internal/inspect"
	"github.com/gligneul/nonodo/internal/model"
	"github.com/gligneul/nonodo/internal/reader"
	"github.com/gligneul/nonodo/internal/rollup"
	"github.com/gligneul/nonodo/internal/supervisor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var ApplicationConflictErr = errors.New("can't use built-in echo with custom application")

// Options to nonodo.
type NonodoOpts struct {
	AnvilPort    int
	AnvilVerbose bool

	HttpAddress string
	HttpPort    int

	InputBoxAddress    string
	InputBoxBlock      uint64
	ApplicationAddress string

	// If RpcUrl is set, connect to it instead of anvil.
	RpcUrl string

	// If set, start echo dapp.
	EnableEcho bool

	// If set, start application.
	ApplicationArgs []string
}

// Create the options struct with default values.
func NewNonodoOpts() NonodoOpts {
	return NonodoOpts{
		AnvilPort:          devnet.AnvilDefaultPort,
		AnvilVerbose:       false,
		HttpAddress:        "127.0.0.1",
		HttpPort:           8080,
		InputBoxAddress:    devnet.InputBoxAddress,
		InputBoxBlock:      0,
		ApplicationAddress: devnet.ApplicationAddress,
		RpcUrl:             "",
		EnableEcho:         false,
		ApplicationArgs:    nil,
	}
}

// Start nonodo.
func NewNonodoWorker(opts NonodoOpts) (w supervisor.SupervisorWorker, err error) {
	if opts.EnableEcho && len(opts.ApplicationArgs) > 0 {
		return w, ApplicationConflictErr
	}

	w.Name = "nonodo"

	model := model.NewNonodoModel()
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	rollup.Register(e, model)
	inspect.Register(e, model)
	reader.Register(e, model)

	if opts.RpcUrl == "" {
		w.Workers = append(w.Workers, devnet.AnvilWorker{
			Port:    opts.AnvilPort,
			Verbose: opts.AnvilVerbose,
		})
		opts.RpcUrl = fmt.Sprintf("ws://127.0.0.1:%v", opts.AnvilPort)
	}
	w.Workers = append(w.Workers, inputter.InputterWorker{
		Model:              model,
		Provider:           opts.RpcUrl,
		InputBoxAddress:    common.HexToAddress(opts.InputBoxAddress),
		InputBoxBlock:      opts.InputBoxBlock,
		ApplicationAddress: common.HexToAddress(opts.ApplicationAddress),
	})
	w.Workers = append(w.Workers, supervisor.HttpWorker{
		Address: fmt.Sprintf("%v:%v", opts.HttpAddress, opts.HttpPort),
		Handler: e,
	})
	if len(opts.ApplicationArgs) > 0 {
		w.Workers = append(w.Workers, supervisor.CommandWorker{
			Name:    "app",
			Command: opts.ApplicationArgs[0],
			Args:    opts.ApplicationArgs[1:],
		})
	} else if opts.EnableEcho {
		w.Workers = append(w.Workers, echoapp.EchoAppWorker{
			RollupEndpoint: fmt.Sprintf("http://127.0.0.1:%v/rollup", opts.HttpPort),
		})
	}

	return w, nil
}
