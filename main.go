package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const emulatorOkMsg = "emulator: INFO: Found systemPath"

func main() {
	app := &cli.App{
		Name:  "emu",
		Usage: "manage android emulators with ease",
		Commands: []*cli.Command{
			&runCommand,
			&listCommand,
			&killCommand,
			&themeCommand,
			&fontCommand,
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

func read(r io.Reader, resultStream chan bool) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, emulatorOkMsg) {
			resultStream <- true
		} else if strings.HasPrefix(line, "ERROR") {
			resultStream <- false
		}
	}
}

var runCommand = cli.Command{
	Name:      "run",
	Usage:     "boot avd",
	ArgsUsage: "<avd>",
	Action: func(c *cli.Context) error {
		avd := c.Args().First()
		if avd == "" {
			return fmt.Errorf("avd not specified")
		}

		args := []string{fmt.Sprintf("@%s", avd), "-no-boot-anim", "-no-audio"}
		cmd := exec.Command("emulator", args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %v", err)
		}

		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to run emulator: %v", err)
		}

		resultStream := make(chan bool)
		go read(stdout, resultStream)

		select {
		case result := <-resultStream:
			if !result {
				log.Fatalf("failed to start emulator: ")
			} else {
				fmt.Println("started emulator")
			}
		case <-time.After(5 * time.Second):
			log.Fatalf("failed to start emulator: timed out")
		}

		return nil
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list emulators",

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
		return nil
	},
}

var killCommand = cli.Command{
	Name:  "kill",
	Usage: "kill emulators",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "kill all emulators",
		},
	},
	Action: func(c *cli.Context) error {
		return nil
	},
}

var themeCommand = cli.Command{
	Name:  "theme",
	Usage: "switch between light and dark mode",
	Subcommands: []*cli.Command{
		{
			Name: "light",
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
			Name: "dark",
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
			Name: "toggle",
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

var fontCommand = cli.Command{
	Name:  "font",
	Usage: "switch fonts",
	Subcommands: []*cli.Command{
		{
			Name: "small",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name: "default",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name: "large",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name: "largest",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name: "reset",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	},
}

// func execAdb(args ...string) (string, error) {
// 	arg
// 	for _, a := range args {

// 	}

// 	cmd := exec.Command("adb", "shell",)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to run %s: %v", cmd, err)
// 	}

// 	return string(output), nil
// }
