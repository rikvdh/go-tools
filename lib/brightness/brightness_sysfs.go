// +build linux

// Package brightness controls brightness for backlight-enabled systems
package brightness

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (b *Brightness) init() error {
	err := filepath.Walk("/sys/class/backlight/",
		func(path string, info os.FileInfo, err error) error {
			if path != "/sys/class/backlight/" {
				b.maxFile = path + `/max_brightness`
				b.brightnessFile = path + `/brightness`
			}
			return nil
		})
	if err != nil {
		return err
	}
	if len(b.maxFile) == 0 {
		return fmt.Errorf("Need brightness-file")
	}
	max, err := ioutil.ReadFile(b.maxFile)
	if err != nil {
		return err
	}
	b.maxValue, err = strconv.ParseFloat(strings.TrimSpace(string(max)), 64)
	return err
}

// Get returns the brightness for sysfs-enabled systems (i.e. Linux)
func (b *Brightness) Get() float64 {
	br, err := ioutil.ReadFile(b.brightnessFile)
	if err != nil {
		panic(err)
	}

	brf, err := strconv.ParseFloat(strings.TrimSpace(string(br)), 64)
	if err != nil {
		panic(err)
	}
	return (brf / b.maxValue) * 100
}

// Set sets the brightness via Sysfs
func (b *Brightness) Set(newBrightness float64) {
	fmt.Printf("Setting brightness to %.2f\n", newBrightness)
	newRaw := (b.maxValue / 100 * newBrightness)
	rawStr := strconv.FormatUint(uint64(newRaw), 10)
	err := ioutil.WriteFile(b.brightnessFile, []byte(rawStr), 0)
	if err != nil {
		fmt.Printf("Error setting brightness: %v\n", err)
	}
}
