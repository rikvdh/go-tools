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
	br := 5.0

	if (now.After(s.TwilightBegin.Time) && now.Before(s.Sunrise.Time)) || now.Equal(s.Sunrise.Time) {
		period := s.Sunrise.Sub(s.TwilightBegin.Time).Seconds()
		timeSince := now.Sub(s.TwilightBegin.Time).Seconds()
		pos := (timeSince / period) * math.Pi
		curveVal := (math.Cos(pos+math.Pi) + 1) / 2
		br = (*maxBrightness-*minBrightness)*curveVal + *minBrightness
	} else if (now.After(s.Sunset.Time) && now.Before(s.TwilightEnd.Time)) || now.Equal(s.Sunset.Time) {
		period := s.TwilightEnd.Sub(s.Sunset.Time).Seconds()
		timeSince := now.Sub(s.Sunset.Time).Seconds()
		pos := (timeSince / period) * math.Pi
		curveVal := (math.Cos(pos) + 1) / 2
		br = (*maxBrightness-*minBrightness)*curveVal + *minBrightness
	} else if now.After(s.Sunrise.Time) && now.Before(s.Sunset.Time) {
		br = *maxBrightness
	} else {
		br = *minBrightness
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

	interval := time.NewTicker(time.Second * 10)
	b.Set(brightnessCalculator(time.Now(), s))
	lastBrightness := time.Now()
	lastSuntimes := time.Now()
	for {
		select {
		case <-interval.C:
			if time.Since(lastSuntimes) > time.Hour {
				sNew, err := newSunTimes(*lat, *long)
				if err == nil {
					s = sNew
					lastSuntimes = time.Now()
				} else {
					fmt.Printf("problem retrieving suntimes: %v", err)
				}
			}
			if time.Since(lastBrightness) > time.Minute {
				b.Set(brightnessCalculator(time.Now(), s))
				lastBrightness = time.Now()
			}
		}
	}
}
