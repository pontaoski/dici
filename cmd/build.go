package cmd

import (
	"dici/builder"
	"dici/pkg"

	"github.com/urfave/cli/v2"
)

func init() {
	RegisterCommand(&cli.Command{
		Name:  "build",
		Usage: "Build dici packages",
		Subcommands: []*cli.Command{
			{
				Name:        "compress",
				Usage:       "Compress a directory into a package",
				Description: "This command allows you to compress any directory into a dici-format package.",
				ArgsUsage:   "[directory] [output]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "description",
						Aliases:  []string{"d"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "version",
						Aliases:  []string{"v"},
						Required: true,
					},
				},
				Action: Squash,
			},
		},
	})
}

// Squash handles dici build compress
func Squash(c *cli.Context) error {
	if c.Args().Len() < 2 {
		Error("The following arguments are required: [directory] [output]")
	}
	data := pkg.PackageMetadata{
		Name:        c.Value("name").(string),
		Description: c.Value("description").(string),
		Version:     c.Value("version").(string),
	}
	err := builder.SquashPackage(c.Args().Get(0), c.Args().Get(1), data)
	if err != nil {
		Error(err.Error())
	}
	return nil
}
