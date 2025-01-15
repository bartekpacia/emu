package emulator

import (
	"fmt"
	"os/exec"
	"strings"
)

// SystemImage is a unique identifier of an Android OS image .Examples:
//   - system-images;android-34;google_apis_playstore;arm64-v8a
//   - system-images;android-35;google_apis;x86_64
type SystemImage string

// ApiLevel returns the API level of this system image.
// Error can be returned when e.g. the new Android version still has only a codename, not a number.
func (s SystemImage) ApiLevel() string {
	substrings := strings.Split(string(s), ";")
	if len(substrings) != 4 {
		panic("invalid output")
	}

	// E.g. android-35, android-Baklava, or android-34-ext9
	androidPart := substrings[1]
	substrings = strings.Split(androidPart, "-")
	if len(substrings) == 0 {
		panic("invalid output")
	}

	return substrings[1]
}

// SystemImages returns installed Android system images.
func SystemImages() ([]SystemImage, error) {
	systemImages := make([]SystemImage, 0)

	cmd := exec.Command("sdkmanager", "--list_installed")
	printInvocation(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run sdkmanager: %v", err)
	}

	// Sample output:
	// system-images;android-33;google_apis;arm64-v8a           | 17            | Google APIs ARM 64 v8a System Image         | system-images/android-33/google_apis/arm64-v8a
	// system-images;android-33;google_apis_playstore;arm64-v8a | 9             | Google Play ARM 64 v8a System Image         | system-images/android-33/google_apis_playstore/arm64-v8a
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line := strings.Split(line, " ")[0]
		if strings.HasPrefix(line, "system-images;") {
			systemImages = append(systemImages, SystemImage(line))
		}
	}

	return systemImages, nil
}
