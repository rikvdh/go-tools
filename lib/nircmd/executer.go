package nircmd

import (
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var m sync.Mutex

var nircmdLocation = os.TempDir() + string(os.PathSeparator) + "nircmd.exe"

func prepare() {
	m.Lock()
	if _, err := os.Stat(nircmdLocation); os.IsNotExist(err) {
		RestoreAsset(os.TempDir(), "nircmd.exe")
	}
	m.Unlock()
}

// SetBrightness via Nircmd
func SetBrightness(val uint64) error {
	prepare()
	cmd := exec.Command(nircmdLocation, "setbrightness", strconv.FormatUint(val, 10))
	return cmd.Run()
}
