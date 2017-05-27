package main

import (
	"flag"
	"fmt"
	"math"
	"time"

	"github.com/rikvdh/go-tools/lib/brightness"
)

var minBrightness *float64
var maxBrightness *float64

func brightnessCalculator(now time.Time, s *SunTimes) float64 {
	br := *minBrightness
	endDim := s.Sunrise.Time
	startDim := s.Sunset.Time
	if now.After(endDim) && now.Before(startDim) {
		period := startDim.Sub(endDim).Seconds()
		timeSince := now.Sub(endDim).Seconds()
		pos := (timeSince / period) * (math.Pi * 2)
		curveVal := (math.Cos(pos+math.Pi) + 1) / 2
		br = (*maxBrightness-*minBrightness)*curveVal + *minBrightness
	}
	return br
}

func main() {
	lat := flag.String("lat", "51.697816", "Latitude for the user location")
	long := flag.String("long", "5.303675", "Longitude for the user location")
	minBrightness = flag.Float64("min", 4, "minimal brightness")
	maxBrightness = flag.Float64("max", 80, "maximum brightness")
	flag.Parse()

	b, err := brightness.New()
	if err != nil {
		panic(err)
	}
	s, err := newSunTimes(*lat, *long)
	if err != nil {
		panic(err)
	}

	interval := time.NewTicker(time.Second * 5)
	b.Set(brightnessCalculator(time.Now(), s))
	lastBrightness := time.Now()
	lastSuntimes := time.Now()
	for range interval.C {
		if time.Since(lastSuntimes) > time.Hour {
			sNew, err := newSunTimes(*lat, *long)
			if err == nil {
				s = sNew
				b.Set(brightnessCalculator(time.Now(), s))
				lastSuntimes = time.Now()
				lastBrightness = time.Now()
			} else {
				fmt.Printf("problem retrieving suntimes: %v\n", err)
			}
		}
		if time.Since(lastBrightness) > time.Second*20 {
			b.Set(brightnessCalculator(time.Now(), s))
			lastBrightness = time.Now()
		}
	}
}
