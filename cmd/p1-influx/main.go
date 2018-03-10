package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/rikvdh/dsmr4p1"
	"github.com/tarm/serial"
)

const (
	gpioP1Request = "/sys/class/gpio/gpio2/value"

	influxDb       = "energy"
	influxUser     = "energy"
	influxPassword = "dsmr"
)

func enableP1Messages() error {
	return ioutil.WriteFile(gpioP1Request, []byte{'1'}, 0777)
}
func disableP1Messages() error {
	return ioutil.WriteFile(gpioP1Request, []byte{'0'}, 0777)
}

func main() {
	baud := flag.Int("baud", 115200, "P1 BAUD rate")
	flag.Parse()
	ser, err := serial.OpenPort(&serial.Config{Name: flag.Arg(0), Baud: *baud, ReadTimeout: time.Second * 15})
	if err != nil {
		panic(err)
	}
	defer ser.Close()

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     flag.Arg(1),
		Username: influxUser,
		Password: influxPassword,
	})
	if err != nil {
		panic(err)
	}

	if err := enableP1Messages(); err != nil {
		panic(err)
	}

	ch := dsmr4p1.Poll(ser)
	for t := range ch {
		fmt.Printf("Received telegram: %+v\n", t)
		_ = t
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  influxDb,
			Precision: "ms",
		})
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		p, err := client.NewPoint("energy", map[string]string{
			"ElectricityEquipmentID": t.ElectricityEquipmentID,
			"DsmrVersion":            fmt.Sprintf("%d", t.DsmrVersion.Val),
		}, map[string]interface{}{
			"ElectricityDelivered1":         t.ElectricityDelivered1.Val,
			"ElectricityReturned1":          t.ElectricityReturned1.Val,
			"ElectricityDelivered2":         t.ElectricityDelivered2.Val,
			"ElectricityReturned2":          t.ElectricityReturned2.Val,
			"ElectricityCurrentlyDelivered": t.ElectricityCurrentlyDelivered.Val,
			"ElectricityCurrentlyReturned":  t.ElectricityCurrentlyReturned.Val,
			"PhaseCurrentlyDeliveredL1":     t.PhaseCurrentlyDeliveredL1.Val,
			"PhaseCurrentlyDeliveredL2":     t.PhaseCurrentlyDeliveredL2.Val,
			"PhaseCurrentlyDeliveredL3":     t.PhaseCurrentlyDeliveredL3.Val,
			"DsmrVersion":                   t.DsmrVersion.Val,
			"ElectricityTariff":             t.ElectricityTariff.Val,
			"PowerFailureCount":             t.PowerFailureCount.Val,
			"LongPowerFailureCount":         t.LongPowerFailureCount.Val,
			"InstantaneousCurrentL1":        t.InstantaneousCurrentL1.Val,
			"InstantaneousCurrentL2":        t.InstantaneousCurrentL2.Val,
			"InstantaneousCurrentL3":        t.InstantaneousCurrentL3.Val,
			"InstantaneousActivePowerL1":    t.InstantaneousActivePowerL1.Val,
			"InstantaneousActivePowerL2":    t.InstantaneousActivePowerL2.Val,
			"InstantaneousActivePowerL3":    t.InstantaneousActivePowerL3.Val,
			"VoltageSagCountL1":             t.VoltageSagCountL1.Val,
			"VoltageSagCountL2":             t.VoltageSagCountL2.Val,
			"VoltageSagCountL3":             t.VoltageSagCountL3.Val,
			"VoltageSwellCountL1":           t.VoltageSwellCountL1.Val,
			"VoltageSwellCountL2":           t.VoltageSwellCountL2.Val,
			"VoltageSwellCountL3":           t.VoltageSwellCountL3.Val,
		}, time.Time(t.Timestamp).UTC())
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		bp.AddPoint(p)

		pn, err := client.NewPoint("gas", map[string]string{
			"GasEquipmentID": t.GasEquipmentID,
		}, map[string]interface{}{
			"GasValue": t.GasTimeValue.Value,
		}, time.Time(t.GasTimeValue.Timestamp).UTC())
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		bp.AddPoint(pn)
		if err := c.Write(bp); err != nil {
			log.Printf("error writing to influx: %v", err)
		}
	}

	rdr := bufio.NewReader(ser)
	for {
		l, _, err := rdr.ReadLine()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(l))
	}

	disableP1Messages()
}
