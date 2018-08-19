package main

// #cgo LDFLAGS: -lX11 -lasound
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"log"
	"time"

	psu "github.com/KimNorgaard/gopsu"
)

var dpy = C.XOpenDisplay(nil)

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func main() {
	statusRuneMap := map[string]rune{
		"Unknown":      '?',
		"Full":         '∞',
		"Not Charging": '!',
		"Charging":     '+',
		"Discharging":  '-',
	}

	if dpy == nil {
		log.Fatal("Can't open display")
	}
	for {
		t := time.Now().Format("Jan _2 15:04")
		psus, _ := psu.GetPowerSupplies()
		st := psus.GetBatteriesStatus()
		statusShort := string(statusRuneMap[st])
		output := fmt.Sprintf("%.0f%%%s", psus.GetBatteriesCapacityPercent(), statusShort)
		if st == "Charging" || st == "Discharging" {
			h, m, _ := psus.GetBatteriesCapacityTime()
			output = fmt.Sprintf("%s [%dh:%dm]", output, h, m)
		}
		output = fmt.Sprintf("%s  %s", output, t)
		fmt.Println(output)
		s := C.CString(output)
		setStatus(s)
		time.Sleep(time.Minute)
	}
}
