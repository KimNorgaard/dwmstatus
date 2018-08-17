package main

// #cgo LDFLAGS: -lX11 -lasound
// #include <X11/Xlib.h>
import "C"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var dpy = C.XOpenDisplay(nil)

const PowerSupplyPath string = "/sys/class/power_supply/"

type Battery struct {
	Name        string
	StatusText  string
	CapacityPct float64
}

func NewBattery(name string) *Battery {
	b := &Battery{
		Name: name,
	}
	b.Update()
	return b
}

func (b *Battery) IsBattery() bool {
	supplyTypeFile := filepath.Join(PowerSupplyPath, b.Name, "type")

	if str, err := StringFromFile(supplyTypeFile); err == nil && strings.Contains(str, "Battery") {
		return true
	}

	return false
}

func (b *Battery) GetStatusText() string {
	status, _ := StringFromFile(filepath.Join(PowerSupplyPath, b.Name, "status"))
	return status
}

func (b *Battery) GetChargeNow() int64 {
	chargeNow, _ := StringFromFile(filepath.Join(PowerSupplyPath, b.Name, "energy_now"))
	i, _ := strconv.ParseInt(chargeNow, 10, 64)
	return i
}

func (b *Battery) GetChargeFull() int64 {
	chargeFull, _ := StringFromFile(filepath.Join(PowerSupplyPath, b.Name, "energy_full"))
	i, _ := strconv.ParseInt(chargeFull, 10, 64)
	return i
}

func (b *Battery) GetCapacityPercent() float64 {
	return float64(100) * float64(b.GetChargeNow()) / float64(b.GetChargeFull())
}

func (b *Battery) Update() {
	b.StatusText = b.GetStatusText()
	b.CapacityPct = b.GetCapacityPercent()
}

func GetBatteries() []*Battery {
	batteries := make([]*Battery, 0)

	paths, err := filepath.Glob(PowerSupplyPath + "*")
	if err != nil {
		return batteries
	}

	for _, path := range paths {
		supply := filepath.Base(path)
		battery := NewBattery(supply)
		if battery.IsBattery() {
			batteries = append(batteries, battery)
		}
	}

	return batteries
}

func GetBatteriesChargeNow(batteries []*Battery) int64 {
	var nowSum int64 = 0
	for _, b := range batteries {
		nowSum = nowSum + b.GetChargeNow()
	}
	return nowSum
}

func GetBatteriesChargeFull(batteries []*Battery) int64 {
	var fullSum int64 = 0
	for _, b := range batteries {
		fullSum = fullSum + b.GetChargeFull()
	}
	return fullSum
}

func GetBatteriesCapacity(batteries []*Battery) float64 {
	return (float64(100) * float64(GetBatteriesChargeNow(batteries)) / float64(GetBatteriesChargeFull(batteries)))
}

func GetBatteriesStatus(batteries []*Battery) string {
	var maxStatus string
	var maxStatusValue int

	statusMap := map[string]int{
		"Unknown":      0,
		"Full":         1,
		"Not Charging": 2,
		"Charging":     3,
		"Discharging":  4,
	}

	for _, b := range batteries {
		status := b.GetStatusText()

		statusValue := statusMap[status]
		if statusValue > maxStatusValue {
			maxStatusValue = statusValue
		}
	}

	for k, v := range statusMap {
		if v == maxStatusValue {
			maxStatus = k
		}
	}

	return maxStatus
}

type Batteries struct {
	StatusText  string
	CapacityPct float64
}

func GetAllBatteries() *Batteries {
	batteries := GetBatteries()
	bts := &Batteries{
		StatusText:  GetBatteriesStatus(batteries),
		CapacityPct: GetBatteriesCapacity(batteries),
	}
	return bts
}

func StringFromFile(fileName string) (string, error) {
	fh, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(bytes.NewBuffer(b).String(), " \n"), nil
}

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func main() {
	if dpy == nil {
		log.Fatal("Can't open display")
	}
	for {
		t := time.Now().Format("Jan _2 15:04")
		bats := GetAllBatteries()
		s := C.CString(fmt.Sprintf("%.0f%% (%s)  %s", bats.CapacityPct, bats.StatusText, t))
		setStatus(s)
		time.Sleep(time.Minute)
	}
}
