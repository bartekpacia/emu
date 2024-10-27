// Package emulator provides common functionality to manage Android emulators.
package emulator

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// PrintInvocations controls whether to print invocations of subprocesses.
var PrintInvocations bool

// Serial is the serial port of the running AVD, for example "emulator-5554".
var Serial string

// AVD represents an Android Virtual Device.
//
// It assumes that no 2 instances of the same AVD run at the same time.
type AVD struct {
	// Name of the AVD, for example "Pixel_7_API_33".
	Name    string
	Running bool

	// PID of the emulator process. Equals 0 if Running is false.
	Pid int

	// ADB
	Serial string
}

func (a AVD) Describe() string {
	suffix := ""
	if a.Running {
		suffix = " RUNNING"
	}

	return fmt.Sprintf("%s%s", a.Name, suffix)
}

// List returns a list of available AVDs and whether they're running or not.
//
// It merges information from "emulator -list-avds" and "adb devices".
//
// If device appears in
func List() ([]AVD, error) {
	cmd := exec.Command("emulator", "-list-avds")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	avdsStr := strings.Split(string(data), "\n")
	// remove empty strings
	for i := len(avdsStr) - 1; i >= 0; i-- {
		if avdsStr[i] == "" {
			avdsStr = append(avdsStr[:i], avdsStr[i+1:]...)
		}
	}

	// Workaround for a bug in emulator v34
	for i, avd := range avdsStr {
		if strings.Contains(avd, "Storing crashdata") {
			avdsStr = slices.Delete(avdsStr, i, i+1)
		}
	}

	// map avds to AVD struct
	avds := make([]AVD, len(avdsStr))
	for i, avd := range avdsStr {
		avds[i] = AVD{Name: avd}
	}

	cmd = exec.Command(
		"ps",
		"-e",
		"-ww", // don't truncate output
		"-o", "pid=,comm=",
	)
	data, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %v", err)
	}

	// parse output of ps
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "qemu-system") {
			continue
		}

		// remove leading and trailing spaces
		line = strings.Trim(line, " ")

		fields := strings.Split(line, " ")
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse pid: %v", err)
		}

		avdName := emuInPID(pid)
		for i, avd := range avds {
			if avd.Name == avdName {
				avds[i].Running = true
				avds[i].Pid = pid
			}
		}
	}

	return avds, nil
}

// Start starts the AVD with the given name.
func Start(name string) error {
	avds, err := List()
	if err != nil {
		return fmt.Errorf("list avds: %v", err)
	}

	for _, avd := range avds {
		if avd.Name == name {
			if avd.Running {
				return fmt.Errorf("avd %s is already running", name)
			}

			args := []string{fmt.Sprintf("@%s", name), "-no-boot-anim", "-no-audio"}
			cmd := exec.Command("emulator", args...)
			printInvocation(cmd)
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("start avd %s: %v", name, err)
			}

			return nil
		}
	}

	return fmt.Errorf("avd %s not found", name)
}

func EnableDarkTheme() error {
	return adbShell("cmd", "uimode", "night", "yes")
}

func DisableDarkTheme() error {
	return adbShell("cmd", "uimode", "night", "no")
}

func ToggleDarkTheme() error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("adb", "shell", "cmd", "uimode", "night")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	printInvocation(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run and read stdout: %v: stderr: %v", err, stderr.String())
	}
	output := stdout.String()

	targetMode := "yes"
	if output == "Night mode: yes\n" {
		targetMode = "no"
	}

	return adbShell("cmd", "uimode", "night", targetMode)
}

func SetFontSize(value string) error {
	return adbShell("settings", "put", "system", "font_scale", value)
}

func SetDisplaySize(value float32) error {
	density, err := getDensity()
	if err != nil {
		return fmt.Errorf("failed to get density: %v", err)
	}

	return adbShell("wm", "density", fmt.Sprintf("%d", int(float32(density)*value)))
}

func DisableAnimations() error {
	var err error
	err = adbShell("settings", "put", "global", "window_animation_scale", "0")
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "transition_animation_scale", "0")
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "animator_duration_scale", "0")
	return err
}

func EnableAnimations() error {
	var err error
	err = adbShell("settings", "put", "global", "window_animation_scale", "1")
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "transition_animation_scale", "1")
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "animator_duration_scale", "1")
	return err
}

func ToggleAnimations() error {
	// the 3 values are always in sync, so I think it's enough to get just a single one
	cmd := exec.Command("adb", "shell", "settings", "get", "global", "window_animation_scale")
	printInvocation(cmd)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run and read stdout: %v", err)
	}
	output := string(out)

	// basic validation
	if output != "0\n" && output != "1\n" {
		return fmt.Errorf("unexpected output: %s", output)
	}

	targetScale := "1"
	if output == "1\n" {
		targetScale = "0"
	}

	err = adbShell("settings", "put", "global", "window_animation_scale", targetScale)
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "transition_animation_scale", targetScale)
	if err != nil {
		return err
	}
	err = adbShell("settings", "put", "global", "animator_duration_scale", targetScale)
	return err
}

func getDensity() (int, error) {
	cmd := exec.Command("adb", "shell", "wm", "density")
	printInvocation(cmd)
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

// emuInPID returns the name of the AVD that is running in process.
//
// Returns an empty string if the process isn't an emulator process.
func emuInPID(pid int) string {
	cmd := exec.Command(
		"ps",
		"-e",
		"-ww", // don't truncate output
		"-o", "args=",
		"-p", strconv.Itoa(pid),
	)
	data, err := cmd.Output()
	if err != nil {
		return ""
	}

	args := strings.Split(string(data), " ")
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			return strings.TrimPrefix(arg, "@")
		}
	}

	return ""
}

func useSerial() (string, error) {
	devices, err := adbDevices()
	if err != nil {
		return "", fmt.Errorf("failed to get devices: %v", err)
	}

	if Serial == "" {
		if err != nil {
			return "", fmt.Errorf("failed to find default serial: %v", err)
		}

		if len(devices) == 0 {
			return "", fmt.Errorf("no devices connected")
		}

		// if len(devices) > 1 {
		// 	log.Printf("More than 1 device connected, using the first one (%s)\n", devices[0])
		// }

		return devices[0], nil
	} else {
		for _, device := range devices {
			if device == Serial {
				return Serial, nil
			}
		}

		return "", fmt.Errorf("serial %s not found in \"adb devices\" output", Serial)

	}
}

// adbShell runs a command on device using "adb shell".
func adbShell(cmd ...string) error {
	serial, err := useSerial()
	if err != nil {
		return fmt.Errorf("failed to get serial: %v", err)
	}

	args := []string{"-s", serial, "shell"}
	args = append(args, cmd...)

	var stderr bytes.Buffer

	adbCmd := exec.Command("adb", args...)
	printInvocation(adbCmd)
	adbCmd.Stderr = &stderr
	err = adbCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run %s: %v, %v", cmd, err, stderr.String())
	}
	return nil
}

func adbDevices() ([]string, error) {
	cmd := exec.Command("adb", "devices")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run and read stdout: %v", err)
	}
	output := string(out)

	lines := strings.Split(output, "\n")
	devices := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "List of devices attached") {
			continue
		}

		if strings.Contains(line, "device") {
			fields := strings.Fields(line)
			devices = append(devices, fields[0])
		}
	}

	return devices, nil
}

func printInvocation(cmd *exec.Cmd) {
	if PrintInvocations {
		log.Println(cmd.String())
	}
}
