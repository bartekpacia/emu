package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	emulator "github.com/bartekpacia/emu"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

const (
	categoryManage  = "Manage AVDs"
	categoryControl = "Control a running AVD"
)

// This is set by GoReleaser, see https://goreleaser.com/cookbooks/using-main.version
var version = "dev"

func main() {
	log.SetFlags(0)
	emulator.PrintInvocations = true

	root := &cli.Command{
		Name:                  "emu",
		Usage:                 "Manage android emulators with ease",
		Authors:               []any{"Bartek Pacia <barpac02@gmail.com>"},
		Version:               version,
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "do not print invocations of subprocesses",
				Action: func(ctx context.Context, c *cli.Command, value bool) error {
					emulator.PrintInvocations = !value
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			// control
			&themeCommand,
			&fontsizeCommand,
			&displaysizeCommand,
			&animationsCommand,
			// manage
			&runCommand,
			&listCommand,
			&killCommand,
			// docs
			&printDocsCommand,
		},
		CommandNotFound: func(ctx context.Context, c *cli.Command, command string) {
			log.Printf("invalid command '%s'. See 'emu --help'\n", command)
		},
	}

	err := root.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

var runCommand = cli.Command{
	Name:      "run",
	Usage:     "Boot AVD",
	ArgsUsage: "<avd>",
	Category:  categoryManage,
	Action: func(ctx context.Context, c *cli.Command) error {
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
	ShellComplete: func(ctx context.Context, c *cli.Command) {
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
	Name:            "list",
	Aliases:         []string{"ls"},
	Usage:           "List all AVDs",
	Category:        categoryManage,
	HideHelpCommand: true,
	Action: func(ctx context.Context, c *cli.Command) error {
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
	Name:     "kill",
	Usage:    "Kill running AVDs",
	Category: categoryManage,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "kill all emulators",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
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

				err := syscall.Kill(avd.Pid, syscall.SIGKILL)
				if err != nil {
					return fmt.Errorf("failed to kill avd %#v: %v", avdName, err)
				}
				return nil
			}
		}

		return fmt.Errorf("avd '%s' not found", avdName)
	},
	ShellComplete: func(ctx context.Context, c *cli.Command) {
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
	Usage:           "Switch between light and dark mode",
	Category:        categoryControl,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&serialFlag,
	},
	Commands: []*cli.Command{
		{
			Name:  "light",
			Usage: "Enables light theme",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.DisableDarkTheme()
			},
		},
		{
			Name:  "dark",
			Usage: "Enables dark theme",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.EnableDarkTheme()
			},
		},
		{
			Name:  "toggle",
			Usage: "Toggles between light and dark theme",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.ToggleDarkTheme()
			},
		},
	},
}

var fontsizeCommand = cli.Command{
	Name:            "fontsize",
	Usage:           "Make text bigger or smaller",
	Category:        categoryControl,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&serialFlag,
	},
	Commands: []*cli.Command{
		{
			Name:  "small",
			Usage: "Sets font scale to 0.85",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetFontSize("0.85")
			},
		},
		{
			Name:  "default",
			Usage: "Sets font scale to 1.0",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetFontSize("1.0")
			},
		},
		{
			Name:  "large",
			Usage: "Sets font scale to 1.15",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetFontSize("1.15")
			},
		},
		{
			Name:  "largest",
			Usage: "Sets font scale to 1.30",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetFontSize("1.30")
			},
		},
	},
}

var displaysizeCommand = cli.Command{
	Name:            "displaysize",
	Usage:           "Make everything bigger or smaller",
	Category:        categoryControl,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&serialFlag,
	},
	Commands: []*cli.Command{
		{
			// e.g. 136
			Name:  "small",
			Usage: "Sets display size to default * 0.85",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetDisplaySize(0.85)
			},
		},
		{
			// e.g. 160
			Name:  "default",
			Usage: "Sets display size to default",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetDisplaySize(1.0)
			},
		},
		{
			// e.g. 186
			Name:  "large",
			Usage: "Sets display size to default * 1.1625",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetDisplaySize(1.1625)
			},
		},
		{
			// e.g. 212
			Name:  "largest",
			Usage: "Sets display size to default * 1.325",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetDisplaySize(1.325)
			},
		},
		{
			// e.g. 240
			Name:  "ultra",
			Usage: "Sets font scale to default * 1.5",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.SetDisplaySize(1.5)
			},
		},
	},
}

var animationsCommand = cli.Command{
	Name:            "animations",
	Usage:           "Enable or disable animations",
	Category:        categoryControl,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&serialFlag,
	},
	Commands: []*cli.Command{
		{
			Name:  "off",
			Usage: "Disables animations",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.DisableAnimations()
			},
		},
		{
			Name:  "on",
			Usage: "Enables animation",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.EnableAnimations()
			},
		},
		{
			Name:  "toggle",
			Usage: "Toggles between light and dark theme",
			Action: func(ctx context.Context, c *cli.Command) error {
				return emulator.ToggleAnimations()
			},
		},
	},
}

var printDocsCommand = cli.Command{
	Name:   "docs",
	Usage:  "Print documentation in various formats",
	Hidden: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:   "format",
			Usage:  "output format [markdown, man, or man-with-section]",
			Hidden: true,
			Value:  "markdown",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		format := cmd.String("format")
		switch format {
		case "", "markdown":
			content, err := docs.ToMarkdown(cmd.Root())
			if err != nil {
				return fmt.Errorf("generate documentation in markdown: %v", err)
			}
			fmt.Println(content)
		case "man":
			content, err := docs.ToMan(cmd.Root())
			if err != nil {
				return fmt.Errorf("generate documentation in man: %v", err)
			}
			fmt.Println(content)
		case "man-with-section":
			content, err := docs.ToManWithSection(cmd.Root(), 1)
			if err != nil {
				return fmt.Errorf("generate documentation in man with section 1: %v", err)
			}
			fmt.Println(content)
		default:
			return fmt.Errorf("invalid documentation format %#v", format)
		}
		return nil
	},
}
