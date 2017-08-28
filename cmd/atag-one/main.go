package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	influxDb       = "energy"
	influxUser     = "energy"
	influxPassword = "dsmr"
)

func findAtagOne() (string, error) {
	con, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 11000})
	if err != nil {
		return "", err
	}

	// TODO: timeout
	for {
		pkg := make([]byte, 37)
		n, addr, err := con.ReadFromUDP(pkg[:])
		if err != nil {
			return "", err
		}
		fmt.Printf("Received %d bytes: %s\n", n, pkg)
		if strings.HasPrefix(string(pkg), "ONE ") {
			return addr.IP.String(), nil
		}
	}
}

type dataReply struct {
	RetrieveReply struct {
		Status struct {
			DeviceID         string `json:"device_id"`
			DeviceStatus     int    `json:"device_status"`
			ConnectionStatus int    `json:"connection_status"`
			DateTime         int64  `json:"date_time"`
		} `json:"status"`
		Report struct {
			ReportTime     int64   `json:"report_time"`
			BurningHours   float64 `json:"burning_hours"`
			RoomTemp       float64 `json:"room_temp"`
			OutsideTemp    float64 `json:"outside_temp"`
			DbgOutsideTemp float64 `json:"dbg_outside_temp"`
			PcbTemp        float64 `json:"pcb_temp"`
			ChSetpoint     float64 `json:"ch_set_point"`
			DhwWaterTemp   float64 `json:"dhw_water_temp"`
			ChWaterTemp    float64 `json:"ch_water_temp"`
			DhwWaterPres   float64 `json:"dhw_water_pres"`
			ChWaterPress   float64 `json:"ch_water_pres"`
			ChReturnTemp   float64 `json:"ch_return_temp"`
			BoilerStatus   int64   `json:"boiler_status"`
			BoilerConfig   int64   `json:"boiler_config"`
			ShownSetTemp   float64 `json:"shown_set_temp"`
		} `json:"report"`
	} `json:"retrieve_reply"`
}

func main() {
	location, err := findAtagOne()
	if err != nil {
		panic(err)
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     os.Args[1],
		Username: influxUser,
		Password: influxPassword,
	})
	if err != nil {
		panic(err)
	}
	for {
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDb,
			Precision: "ms",
		})
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		buf := bytes.NewBufferString(`{"retrieve_message":{
			"seqnr":0,"account_auth":{
				"user_account":"foo@bar.com","mac_address":"13:37:CA:FE:BA:BE"
			},"info":15}}`)
		rep, err := http.Post("http://"+location+":10000/retrieve", "application/json", buf)
		if err != nil {
			fmt.Printf("http request error: %v\n", err)
			continue
		}
		dat := dataReply{}

		err = json.NewDecoder(rep.Body).Decode(&dat)
		rep.Body.Close()
		if err != nil {
			fmt.Printf("json decode failed: %v\n", err)
			continue
		}

		p, err := client.NewPoint("thermostat", map[string]string{
			"ThermostatID": dat.RetrieveReply.Status.DeviceID,
		}, map[string]interface{}{
			"ReportTime":     dat.RetrieveReply.Report.ReportTime,
			"BurningHours":   dat.RetrieveReply.Report.BurningHours,
			"RoomTemp":       dat.RetrieveReply.Report.RoomTemp,
			"OutsideTemp":    dat.RetrieveReply.Report.OutsideTemp,
			"DbgOutsideTemp": dat.RetrieveReply.Report.DbgOutsideTemp,
			"PcbTemp":        dat.RetrieveReply.Report.PcbTemp,
			"ChSetpoint":     dat.RetrieveReply.Report.ChSetpoint,
			"DhwWaterTemp":   dat.RetrieveReply.Report.DhwWaterTemp,
			"ChWaterTemp":    dat.RetrieveReply.Report.ChWaterTemp,
			"DhwWaterPres":   dat.RetrieveReply.Report.DhwWaterPres,
			"ChWaterPress":   dat.RetrieveReply.Report.ChWaterPress,
			"ChReturnTemp":   dat.RetrieveReply.Report.ChReturnTemp,
			"BoilerStatus":   dat.RetrieveReply.Report.BoilerStatus,
			"BoilerConfig":   dat.RetrieveReply.Report.BoilerConfig,
			"ShownSetTemp":   dat.RetrieveReply.Report.ShownSetTemp,
		}, time.Now().UTC())
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		bp.AddPoint(p)
		fmt.Println("writing to influx...")
		if err := c.Write(bp); err != nil {
			log.Printf("error writing to influx: %v", err)
		}
		time.Sleep(time.Second * 30)
	}
}
