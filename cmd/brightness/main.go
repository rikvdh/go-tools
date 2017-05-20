package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var maxFile string
var brightnessFile string

func scanBrightnessFiles() {
	filepath.Walk("/sys/class/backlight/", func(path string, info os.FileInfo, err error) error {
		if path != "/sys/class/backlight/" {
			maxFile = path + `/max_brightness`
			brightnessFile = path + `/brightness`
		}
		return nil
	})
}

func main() {
	scanBrightnessFiles()
	if len(maxFile) == 0 {
		log.Fatal("Need brightness-file")
	}
	newBrightness := flag.Float64("s", -1, "Brightness in %")
	flag.Parse()
	max, err := ioutil.ReadFile(maxFile)
	if err != nil {
		panic(err)
	}
	maxf, err := strconv.ParseFloat(strings.TrimSpace(string(max)), 64)
	if err != nil {
		panic(err)
	}

	if *newBrightness < 0 {
		br, err := ioutil.ReadFile(brightnessFile)
		if err != nil {
			panic(err)
		}

		brf, err := strconv.ParseFloat(strings.TrimSpace(string(br)), 64)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Brightness is: %.1f\n", (brf/maxf)*100)
	} else {
		newRaw := (maxf / 100 * *newBrightness)
		rawStr := strconv.FormatUint(uint64(newRaw), 10)
		ioutil.WriteFile(brightnessFile, []byte(rawStr), 0)
	}
}
