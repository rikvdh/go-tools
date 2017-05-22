// +build windows

// Package brightness controls brightness for backlight-enabled systems
package brightness

import (
	"errors"
	"fmt"

	"github.com/StackExchange/wmi"
)

type wmiBrightnessData struct {
	CurrentBrightness int32
	Levels            int32
	InstanceName      string
}

func (b *Brightness) init() error {
	var dst []wmiBrightnessData

	err := wmi.QueryNamespace(`SELECT CurrentBrightness,Levels,InstanceName FROM WmiMonitorBrightness`, &dst, `root\WMI`)
	if err != nil {
		return err
	}
	if len(dst) < 1 {
		return errors.New("No brightness data received")
	}
	b.maxValue = float64(dst[0].Levels)
	return nil
}

func (b *Brightness) Get() float64 {
	var dst []wmiBrightnessData
	err := wmi.QueryNamespace(`SELECT CurrentBrightness,Levels,InstanceName FROM WmiMonitorBrightness`, &dst, `root\WMI`)
	if err != nil {
		panic(err)
	}
	if len(dst) < 1 {
		panic(errors.New("No brightness data received"))
	}
	return (float64(dst[0].CurrentBrightness) / b.maxValue) * 100
}

func (b *Brightness) Set(newBrightness float64) {
	fmt.Printf("Setting brightness to %.2f\n", newBrightness)
	panic(errors.New("Cant set brightness on windows (yet)"))
}
