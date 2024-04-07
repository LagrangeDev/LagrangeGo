package platform

import (
	"os/exec"
	"regexp"
)

func GetVersion() string {
	out, _ := exec.Command("cmd", "/c", "ver").CombinedOutput()
	match := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`).FindStringSubmatch(string(out))
	return match[0]
}
