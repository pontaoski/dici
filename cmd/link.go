package cmd

import (
	"dici/linker"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
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
				Name:        "daemon",
				Usage:       "Start the dici daemon",
				Description: "This command launches a daemon that monitors dici packages",
				ArgsUsage:   "[packages] [mount] [output]",
				Action:      Daemon,
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

// Daemon handles dici link daemon
func Daemon(c *cli.Context) error {
	if c.Args().Len() < 3 {
		Error("The following arguments are required: [packages] [mount] [output]")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Error(err.Error())
	}
	if err := watcher.Add(c.Args().Get(0)); err != nil {
		Error(err.Error())
	}
	defer watcher.Close()
	defer linker.Unlink(c.Args().Get(1), c.Args().Get(2))

	linker.Link(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
	if err := watcher.Add(c.Args().Get(0)); err != nil {
		Error(err.Error())
	}
	go func() {
		for {
			select {
			case <-watcher.Events:
				err := linker.Link(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
				if err != nil {
					Warning(err.Error())
				}
			case err := <-watcher.Errors:
				Warning(err.Error())
			}
		}
	}()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
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
