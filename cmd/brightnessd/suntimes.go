package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SunTimes struct {
	TwilightBegin CustomTime `json:"civil_twilight_begin"`
	Sunrise       CustomTime `json:"sunrise"`
	DayLength     string     `json:"day_length"`
	SolarNoon     CustomTime `json:"solar_noon"`
	Sunset        CustomTime `json:"sunset"`
	TwilightEnd   CustomTime `json:"civil_twilight_end"`
}

type sunTimeWrapper struct {
	Results SunTimes
	Status  string
}

func newSunTimes(lat string, long string) (*SunTimes, error) {
	b := sunTimeWrapper{}
	fmt.Println("Get sunrise and sunset times")
	rep, err := http.Get("https://api.sunrise-sunset.org/json?lat=" + lat + "&lng=" + long)
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(rep.Body)
	if err := dec.Decode(&b); err != nil {
		panic(err)
	}
	return &b.Results, nil
}

const ctLayout = "3:04:05 PM"

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ctLayout, s)
	t := time.Now().UTC()
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	ct.Time = t.Add(ct.Time.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)))
	return
}
