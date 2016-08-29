// Package remote provides utility of Vim's clientserver feature (:h remote.txt).
package remote

import (
	"os/exec"
	"strings"
)

// ServerList return Vim serverlist (:h --serverlist).
func ServerList() ([]string, error) {
	out, err := exec.Command("vim", "--serverlist").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Trim(string(out), "\n"), "\n"), nil
}

// IsServed returns true if given servername Vim server exists.
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
