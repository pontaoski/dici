package cmd

import (
	"dici/linker"

	"github.com/urfave/cli/v2"
)

func init() {
	RegisterCommand(&cli.Command{
		Name:  "link",
		Usage: "Link packages into a directory",
		Subcommands: []*cli.Command{
			{
				Name:        "activate",
				Usage:       "Activate packages",
				Description: "This command allows you to activate dici packages",
				ArgsUsage:   "[packages] [mount] [output]",
				Action:      Activate,
			},
			{
				Name:        "deactivate",
				Usage:       "Deactivate packages",
				Description: "This command allows you to deactivate dici packages",
				ArgsUsage:   "[mount] [output]",
				Action:      Deactivate,
			},
		},
	})
}

// Activate handles dici link activate
func Activate(c *cli.Context) error {
	if c.Args().Len() < 3 {
		Error("The following arguments are required: [packages] [mount] [output]")
	}
	err := linker.Link(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
	if err != nil {
		Error(err.Error())
	}
	return nil
}

// Deactivate handles dici link activate
func Deactivate(c *cli.Context) error {
	if c.Args().Len() < 2 {
		Error("The following arguments are required: [mount] [output]")
	}
	err := linker.Unlink(c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		Error(err.Error())
	}
	return nil
}
