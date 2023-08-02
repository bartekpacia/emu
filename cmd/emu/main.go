package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	emulator "github.com/bartekpacia/emu"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(0)
	emulator.PrintInvocations = true

	app := &cli.App{
		Name:                 "emu",
		Usage:                "Manage android emulators with ease",
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Commands: []*cli.Command{
			&runCommand,
			&listCommand,
			&killCommand,
			&themeCommand,
			&fontsizeCommand,
			&displaysizeCommand,
		},
		CommandNotFound: func(c *cli.Context, command string) {
			log.Printf("invalid command '%s'. See 'emu --help'\n", command)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

var runCommand = cli.Command{
	Name:      "run",
	Usage:     "Boot AVD",
	ArgsUsage: "<avd>",
	Action: func(c *cli.Context) error {
		avd := c.Args().First()
		if avd == "" {
			return fmt.Errorf("avd not specified")
		}

		err := emulator.Start(avd)
		if err != nil {
			return fmt.Errorf("failed to start emulator: %v", err)
		}

		return nil
	},
	BashComplete: func(c *cli.Context) {
		avds, err := emulator.List()
		if err != nil {
			return
		}

		for _, avd := range avds {
			if avd.Running {
				continue
			}

			fmt.Println(avd.Name)
		}
	},
}

var listCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List all AVDs",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "region-id",
			Aliases: []string{"id"},
			Value:   "",
			Usage:   "region whose generated directory will be compressed",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Value:   false,
			Usage:   "print extensive logs",
		},
	},
	Action: func(c *cli.Context) error {
		avds, err := emulator.List()
		if err != nil {
			return fmt.Errorf("failed to list avds: %v", err)
		}

		for _, avd := range avds {
			fmt.Println(avd.Describe())
		}

		return nil
	},
}

var killCommand = cli.Command{
	Name:  "kill",
	Usage: "Kill running AVDs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "kill all emulators",
		},
	},
	Action: func(c *cli.Context) error {
		avdName := c.Args().First()

		if avdName == "" {
			return fmt.Errorf("avd not specified")
		}

		avds, err := emulator.List()
		if err != nil {
			return fmt.Errorf("failed to list avds: %v", err)
		}

		for _, avd := range avds {
			if avd.Name == avdName {
				if !avd.Running {
					return fmt.Errorf("avd '%s' is not running", avdName)
				}

				syscall.Kill(avd.Pid, syscall.SIGKILL)
				return nil
			}
		}

		return fmt.Errorf("avd '%s' not found", avdName)
	},
	BashComplete: func(c *cli.Context) {
		avds, err := emulator.List()
		if err != nil {
			return
		}

		for _, avd := range avds {
			if !avd.Running {
				continue
			}

			fmt.Println(avd.Name)
		}
	},
}

var themeCommand = cli.Command{
	Name:            "theme",
	HideHelpCommand: true,
	Usage:           "Switch between light and dark mode",
	Subcommands: []*cli.Command{
		{
			Name:  "light",
			Usage: "Enables light theme",
			Action: func(c *cli.Context) error {
				return emulator.DisableDarkTheme()
			},
		},
		{
			Name:  "dark",
			Usage: "Enables dark theme",
			Action: func(c *cli.Context) error {
				return emulator.EnableDarkTheme()
			},
		},
		{
			Name:  "toggle",
			Usage: "Toggles between light and dark theme",
			Action: func(c *cli.Context) error {
				return emulator.ToggleDarkTheme()
			},
		},
	},
}

var fontsizeCommand = cli.Command{
	Name:            "fontsize",
	Usage:           "Make text bigger or smaller",
	HideHelpCommand: true,
	Subcommands: []*cli.Command{
		{
			Name:  "small",
			Usage: "Sets font scale to 0.85",
			Action: func(c *cli.Context) error {
				return emulator.SetFontSize("0.85")
			},
		},
		{
			Name:  "default",
			Usage: "Sets font scale to 1.0",
			Action: func(c *cli.Context) error {
				return emulator.SetFontSize("1.0")
			},
		},
		{
			Name:  "large",
			Usage: "Sets font scale to 1.15",
			Action: func(c *cli.Context) error {
				return emulator.SetFontSize("1.15")
			},
		},
		{
			Name:  "largest",
			Usage: "Sets font scale to 1.30",
			Action: func(c *cli.Context) error {
				return emulator.SetFontSize("1.30")
			},
		},
	},
}

var displaysizeCommand = cli.Command{
	Name:            "displaysize",
	Usage:           "Make everything bigger or smaller",
	HideHelpCommand: true,
	Subcommands: []*cli.Command{
		{
			// e.g. 136
			Name:  "small",
			Usage: "Sets display size to default * 0.85",
			Action: func(c *cli.Context) error {
				return emulator.SetDisplaySize(0.85)
			},
		},
		{
			// e.g. 160
			Name:  "default",
			Usage: "Sets display size to default",
			Action: func(c *cli.Context) error {
				return emulator.SetDisplaySize(1.0)
			},
		},
		{
			// e.g. 186
			Name:  "large",
			Usage: "Sets display size to default * 1.1625",
			Action: func(c *cli.Context) error {
				return emulator.SetDisplaySize(1.1625)
			},
		},
		{
			// e.g. 212
			Name:  "largest",
			Usage: "Sets display size to default * 1.325",
			Action: func(c *cli.Context) error {
				return emulator.SetDisplaySize(1.325)
			},
		},
		{
			// e.g. 240
			Name:  "ultra",
			Usage: "Sets font scale to default * 1.5",
			Action: func(c *cli.Context) error {
				return emulator.SetDisplaySize(1.5)
			},
		},
	},
}
