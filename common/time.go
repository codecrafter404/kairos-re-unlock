package common

import (
	"time"

	"github.com/beevik/ntp"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
)

func QueryOffset(config config.Config) time.Duration {
	time, err := ntp.Query(config.NTPServer)
	if err != nil {
		panic(err)
	}
	return time.ClockOffset
}

func GetCurrentTime(offset time.Duration) time.Time {
	return time.Now().Add(offset)
}
