package doctor

import "os/exec"

// CommandExists reports whether a command can be found in PATH.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
