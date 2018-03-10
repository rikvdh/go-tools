package main

import (
	"flag"
	"math"
	"time"

	"github.com/rikvdh/go-tools/lib/brightness"
	"github.com/sirupsen/logrus"
)

var (
	lat           = flag.String("lat", "51.697816", "Latitude for the user location")
	long          = flag.String("long", "5.303675", "Longitude for the user location")
	minBrightness = flag.Float64("min", 4, "minimal brightness")
	maxBrightness = flag.Float64("max", 80, "maximum brightness")
)

const (
	brightnessInterval = time.Second * 20
	sunTimesInterval   = time.Minute * 30
)

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
	flag.Parse()

	b, err := brightness.New()
	if err != nil {
		panic(err)
	}

	lastSuntimes := time.Now()
	s, err := newSunTimes(*lat, *long)
	if err != nil {
		lastSuntimes = lastSuntimes.Add(-brightnessInterval)
		logrus.Warnf("problem retrieving suntimes: %v", err)
	}

	interval := time.NewTicker(time.Second * 1)
	b.Set(brightnessCalculator(time.Now(), s))
	lastBrightness := time.Now()
	for range interval.C {
		if time.Since(lastSuntimes) > sunTimesInterval {
			sNew, err := newSunTimes(*lat, *long)
			if err == nil {
				s = sNew
				b.Set(brightnessCalculator(time.Now(), s))
				lastSuntimes = time.Now()
				lastBrightness = time.Now()
			} else {
				logrus.Warnf("problem retrieving suntimes: %v\n", err)
			}
		}
		if time.Since(lastBrightness) > brightnessInterval {
			b.Set(brightnessCalculator(time.Now(), s))
			lastBrightness = time.Now()
		}
	}
}
