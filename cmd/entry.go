package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

var cmds []*cli.Command

// RegisterCommand registers a command with dici
func RegisterCommand(c *cli.Command) {
	cmds = append(cmds, c)
}

// Entry is the entrypoint of the dici command line app
func Entry() {
	app := cli.App{
		Name:     "dici",
		Usage:    "Manage packages",
		Commands: cmds,
	}
	err := app.Run(os.Args)
	if err != nil {
		Error(err.Error())
	}
}
