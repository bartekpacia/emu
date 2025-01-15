package emulator

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CreateAVD creates a new Android Virtual Device and returns its name and path.
//
// It wraps the avdmanager tool from Android SDK.Example AVD manager invocation:
//
//	avdmanager create avd \
//	  --sdcard '8192M' \
//	  --package "system-images;android-34;google_apis_playstore;arm64-v8a" \
//	  --name "Pixel_7_API_34" \
//	  --device "pixel_7"
//
// In addition, it also automatically enables keyboard input.
func CreateAVD(osimage SystemImage, skin string, sdcardMB int) (string, string, error) {
	avdName := cases.Title(language.English, cases.NoLower).String(skin)
	avdName = fmt.Sprint(avdName, "_API_", osimage.ApiLevel())
	args := []string{"create", "avd"}
	args = append(args, "--sdcard", strconv.Itoa(sdcardMB)+"M")
	args = append(args, "--package", string(osimage))
	args = append(args, "--name", avdName)
	args = append(args, "--device", skin)

	var stderr bytes.Buffer
	cmd := exec.Command("avdmanager", args...)
	printInvocation(cmd)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("failed to run %s: %v, %v", cmd, err, stderr.String())
	}

	avdPath := filepath.Join(os.Getenv("ANDROID_USER_HOME"), "avd", avdName+".avd")
	err = updateConfig(avdPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to update config %s: %v", avdPath, err)
	}

	return avdName, avdPath, nil
}

func DeleteAVD(avdName string) error {
	avdIniPath := filepath.Join(os.Getenv("ANDROID_USER_HOME"), "avd", avdName+".ini")
	err := os.Remove(avdIniPath)
	if err != nil {
		return fmt.Errorf("delete AVD ini file: %v", err)
	}

	avdDirPath := filepath.Join(os.Getenv("ANDROID_USER_HOME"), "avd", avdName+".avd")
	err = os.RemoveAll(avdDirPath)
	if err != nil {
		return fmt.Errorf("delete AVD directory: %v", err)
	}

	return nil
}

func Skins() ([]string, error) {
	var directories []string

	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		return nil, fmt.Errorf("ANDROID_HOME environment variable not set")
	}

	skinsPath := filepath.Join(androidHome, "skins")

	entries, err := os.ReadDir(skinsPath)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %v", skinsPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

// updateConfig updates the config.ini file to enable keyboard support.
func updateConfig(avdDir string) error {
	configIniPath := filepath.Join(avdDir, "config.ini")

	file, err := os.OpenFile(configIniPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open config.ini file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "hw.keyboard=no") {
			lines = append(lines, "hw.keyboard=yes")
		} else if strings.HasPrefix(line, "vm.heapsize=") {
			lines = append(lines, "vm.heapsize=1024M")
		} else {
			lines = append(lines, line)
		}
	}

	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("scanning %s: %v", configIniPath, err)
	}

	err = os.Truncate(configIniPath, 0)
	if err != nil {
		return fmt.Errorf("truncating %s: %v", configIniPath, err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seeking %s: %v", configIniPath, err)
	}

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("writing %s: %v", configIniPath, err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flushing %s: %v", configIniPath, err)
	}

	return nil
}
