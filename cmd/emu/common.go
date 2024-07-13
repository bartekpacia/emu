package main

import (
	"context"

	emulator "github.com/bartekpacia/emu"
	"github.com/urfave/cli/v3"
)

var serialFlag = cli.StringFlag{
	Name:    "serial",
	Aliases: []string{"s"},
	Usage:   "use device with given serial",
	Action: func(ctx context.Context, c *cli.Command, value string) error {
		emulator.Serial = value
		return nil
	},
}
