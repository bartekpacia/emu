package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	emulator "github.com/bartekpacia/emu"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(0)
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
				cmd := exec.Command("adb", "shell", "cmd", "uimode", "night", "no")
				err := cmd.Run()
				if err != nil {
					return fmt.Errorf("failed to run %s: %v", cmd, err)
				}
				return nil
			},
		},
		{
			Name:  "dark",
			Usage: "Enables dark theme",
			Action: func(c *cli.Context) error {
				cmd := exec.Command("adb", "shell", "cmd", "uimode", "night", "yes")
				err := cmd.Run()
				if err != nil {
					return fmt.Errorf("failed to run %s: %v", cmd, err)
				}
				return nil
			},
		},
		{
			Name:  "toggle",
			Usage: "Toggles between light and dark theme",
			Action: func(c *cli.Context) error {
				cmd := exec.Command("adb", "shell", "cmd", "uimode", "night")
				out, err := cmd.Output()
				if err != nil {
					return fmt.Errorf("failed to run and read stdout: %v", err)
				}
				output := string(out)

				targetMode := "yes"
				if output == "Night mode: yes\n" {
					targetMode = "no"
				} else if output == "Night mode: no\n" {
					targetMode = "yes"
				} else {
					return fmt.Errorf("unknown output: %s", output)
				}

				cmd = exec.Command("adb", "shell", "cmd", "uimode", "night", targetMode)
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("failed to run %s: %v", cmd, err)
				}

				return nil
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
				return setFontSize("0.85")
			},
		},
		{
			Name:  "default",
			Usage: "Sets font scale to 1.0",
			Action: func(c *cli.Context) error {
				return setFontSize("1.0")
			},
		},
		{
			Name:  "large",
			Usage: "Sets font scale to 1.15",
			Action: func(c *cli.Context) error {
				return setFontSize("1.15")
			},
		},
		{
			Name:  "largest",
			Usage: "Sets font scale to 1.30",
			Action: func(c *cli.Context) error {
				return setFontSize("1.30")
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
				return setDisplaySize(0.85)
			},
		},
		{
			// e.g. 160
			Name:  "default",
			Usage: "Sets display size to default",
			Action: func(c *cli.Context) error {
				return setDisplaySize(1.0)
			},
		},
		{
			// e.g. 186
			Name:  "large",
			Usage: "Sets display size to default * 1.1625",
			Action: func(c *cli.Context) error {
				return setDisplaySize(1.1625)
			},
		},
		{
			// e.g. 212
			Name:  "largest",
			Usage: "Sets display size to default * 1.325",
			Action: func(c *cli.Context) error {
				return setDisplaySize(1.325)
			},
		},
		{
			// e.g. 240
			Name:  "ultra",
			Usage: "Sets font scale to default * 1.5",
			Action: func(c *cli.Context) error {
				return setDisplaySize(1.5)
			},
		},
	},
}

func setFontSize(value string) error {
	return adbShell("settings", "put", "system", "font_scale", value)
}

func getDensity() (int, error) {
	cmd := exec.Command("adb", "shell", "wm", "density")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run: %v", err)
	}
	output := string(out)

	var density int
	_, err = fmt.Sscanf(output, "Physical density: %d", &density)
	if err != nil {
		return 0, fmt.Errorf("failed to parse density: %v", err)
	}

	return density, nil
}

func setDisplaySize(value float32) error {
	density, err := getDensity()
	if err != nil {
		return fmt.Errorf("failed to get density: %v", err)
	}

	return adbShell("wm", "density", fmt.Sprintf("%d", int(float32(density)*value)))
}

func adbShell(cmd ...string) error {
	args := []string{"shell"}
	args = append(args, cmd...)

	var stderr bytes.Buffer

	adbCmd := exec.Command("adb", args...)
	log.Println(adbCmd.String())
	adbCmd.Stderr = &stderr
	err := adbCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run %s: %v, %v", cmd, err, stderr.String())
	}
	return nil
}
