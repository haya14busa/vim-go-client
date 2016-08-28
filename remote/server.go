package remote

import (
	"os/exec"
	"strings"
)

func ServerList() ([]string, error) {
	out, err := exec.Command("vim", "--serverlist").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Trim(string(out), "\n"), "\n"), nil
}

func IsServed(servername string) bool {
	servers, err := ServerList()
	if err != nil {
		return false
	}
	for _, s := range servers {
		if s == servername {
			return true
		}
	}
	return false
}
