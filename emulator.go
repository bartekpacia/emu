// Package emulator provides common functionality to manage Android emulators.
package emulator

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// AVD represents an Android Virtual Device.
//
// It assumes that no 2 instances of the same AVD run at the same time.
type AVD struct {
	Name    string
	Running bool

	// PID of the emulator process. Equals 0 if Running is false.
	Pid int
}

func (a AVD) Describe() string {
	suffix := ""
	if a.Running {
		suffix = " RUNNING"
	}

	return fmt.Sprintf("%s%s", a.Name, suffix)
}

// List returns a list of available AVDs and whether they're running or not.
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

	// map avds to AVD struct
	avds := make([]AVD, len(avdsStr))
	for i, avd := range avdsStr {
		avds[i] = AVD{Name: avd}
	}

	cmd = exec.Command(
		"ps",
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
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("start avd %s: %v", name, err)
			}

			return nil
		}
	}

	return fmt.Errorf("avd %s not found", name)
}

// emuInPID returns the name of the AVD that is running in process.
//
// Returns an empty string if the process isn't an emulator process.
func emuInPID(pid int) string {
	cmd := exec.Command(
		"ps",
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
