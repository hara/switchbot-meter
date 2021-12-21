package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/cmd"
	"github.com/pkg/errors"
)

const (
	UUID_SCANRSP      = "cba20d00224d11e69fb80002a5d5c51b"
	UUID_SERVICE_DATA = "0d00"
)

const (
	DEVICE_TYPE_METER = 0x54
)

var (
	Version  = "unknown"
	Revision = "unknown"
)

var (
	addrFlag        string
	timeoutFlag     uint
	daemonFlag      bool
	showVersionFlag bool
)

var cancel context.CancelFunc

func init() {
	flag.StringVar(&addrFlag, "a", "", "MAC address of the meter device.")
	flag.UintVar(&timeoutFlag, "t", 10, "Specify a timeout in seconds before exits. This option has no effect when used with '-d'")
	flag.BoolVar(&daemonFlag, "d", false, "Run as daemon.")
	flag.BoolVar(&showVersionFlag, "v", false, "Show version.")
}

func main() {
	flag.Parse()

	if showVersionFlag {
		fmt.Printf("switch-meter version %v (revision %v)\n", Version, Revision)
		return
	}

	scanp := ble.OptScanParams(cmd.LESetScanParameters{
		LEScanType:           0x01,   // 0x00: passive, 0x01: active
		LEScanInterval:       0x4000, // 0x0004 - 0x4000; N * 0.625msec
		LEScanWindow:         0x4000, // 0x0004 - 0x4000; N * 0.625msec
		OwnAddressType:       0x01,   // 0x00: public, 0x01: random
		ScanningFilterPolicy: 0x00,   // 0x00: accept all, 0x01: ignore non-white-listed.
	})
	dev, err := linux.NewDevice(scanp)
	if err != nil {
		log.Fatalf("could not open device: %v", err)
	}

	ble.SetDefaultDevice(dev)

	var ctx context.Context
	if daemonFlag {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutFlag))
	}
	handleError(ble.Scan(ctx, true, advHandler, nil))
}

func handleError(err error) {
	switch errors.Cause(err) {
	case context.DeadlineExceeded:
	case context.Canceled:
	default:
		log.Fatalf(err.Error())
	}
}

func advHandler(a ble.Advertisement) {
	for _, uuid := range a.Services() {
		if uuid.String() != UUID_SCANRSP {
			break
		}

		if addrFlag != "" && !strings.EqualFold(addrFlag, a.Addr().String()) {
			break
		}

		for _, data := range a.ServiceData() {
			if !isMeterServiceData(data) {
				break
			}

			metric, _ := MetricFromServiceData(data)
			metric.Address = a.Addr().String()

			out, err := json.Marshal(&metric)
			if err != nil {
				log.Fatalf("could not marshal metric: %v", err)
			}
			fmt.Println(string(out))

			if !daemonFlag {
				cancel()
			}
		}
	}
}

func isMeterServiceData(data ble.ServiceData) bool {
	if data.UUID.String() != UUID_SERVICE_DATA {
		return false
	}

	if dtype := data.Data[0] & 0x7f; dtype != DEVICE_TYPE_METER {
		return false
	}

	return true
}

type Metric struct {
	Address     string  `json:"addr"`
	Battery     byte    `json:"bat"`
	Temperature float64 `json:"temp"`
	Humidity    byte    `json:"humi"`
	Timestamp   int64   `json:"ts"`
}

func MetricFromServiceData(d ble.ServiceData) (*Metric, error) {
	// ref: https://github.com/OpenWonderLabs/python-host/wiki/Meter-BLE-open-API#new-broadcast-message
	bat := d.Data[2] & 0x7f
	temp := float64(d.Data[4] & 0x7f)
	temp += float64(d.Data[3]&0x0f) / 10
	if d.Data[4]&0x80 == 0 {
		temp *= -1
	}
	humi := d.Data[5] & 0x7f

	return &Metric{
		Battery:     bat,
		Temperature: temp,
		Humidity:    humi,
		Timestamp:   time.Now().UnixMilli(),
	}, nil
}
