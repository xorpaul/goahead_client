package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	h "github.com/xorpaul/gohelper"
)

func getPayloadFqdn() string {
	if len(config.Fqdn) > 0 {
		return config.Fqdn
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			h.Fatalf("Error while getting hostname Error: " + err.Error())
		}
		return hostname
	}
}

func getPayloadUptime() string {
	dat, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		h.Fatalf("Error while trying to open /proc/uptime. Error: " + err.Error())
	}
	times := strings.Fields(string(dat))
	uptimeSeconds := strings.Split(times[0], ".")[0]
	if err != nil {
		h.Fatalf("Error while trying to ParseFloat uptime. Error: " + err.Error())
	}
	uptime, err := time.ParseDuration(uptimeSeconds + "s")
	if err != nil {
		h.Fatalf("Error while trying to parse uptime to Duration. Error: " + err.Error())
	}

	return uptime.String()
}

func secondsToTime(time int) (int, int, int) {
	seconds := time % 60
	totalMinute := (time - seconds) / 60
	minutes := totalMinute % 60
	hours := (totalMinute - minutes) / 60

	return hours, minutes, seconds
}
