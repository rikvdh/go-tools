package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

const suntimesTest = `{"sunrise":"7:00:00 AM","sunset":"7:00:00 PM",
	"solar_noon":"11:36:34 AM","day_length":"15:48:52",
	"civil_twilight_begin":"3:00:00 AM","civil_twilight_end":"11:00:00 PM",
	"nautical_twilight_begin":"2:01:49 AM","nautical_twilight_end":"9:11:19 PM",
	"astronomical_twilight_begin":"12:25:52 AM",
	"astronomical_twilight_end":"10:47:16 PM"}`

func TestBrightnesSetter(t *testing.T) {
	var max = 100.0
	var min = 0.0

	maxBrightness = &max
	minBrightness = &min

	st := SunTimes{}
	err := json.Unmarshal([]byte(suntimesTest), &st)
	if err != nil {
		t.Error(err)
	}
	tm := time.Now()
	tm = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.UTC)
	endtm := time.Date(tm.Year(), tm.Month(), tm.Day(), 23, 59, 59, 0, time.UTC)
	for {
		f := brightnessCalculator(tm, &st)
		fmt.Printf("%s\t", tm.Format(time.Stamp))
		fmt.Printf("%.2f\n", f)
		tm = tm.Add(time.Minute * 10)
		if endtm.Sub(tm).Seconds() < 0 {
			break
		}
	}
}
