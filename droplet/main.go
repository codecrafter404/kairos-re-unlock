package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/codecrafter404/kairos-re-unlock/droplet/droplet"
	"github.com/kairos-io/kcrypt/pkg/bus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LOGFILE = "/tmp/kcrypt-kairos-re-unlock.log"

func main() {
	if err := os.RemoveAll(LOGFILE); err != nil {
		checkErr(fmt.Errorf("removing the logfile: %w", err))
	}
	file, err := os.OpenFile(
		LOGFILE,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	checkErr(err)
	defer file.Close()

	config, err := config.UnmarshalConfig()
	checkErr(err)

	if config.NTPServer == "" {
		config.NTPServer = "de.pool.ntp.org"
	}

	ntp_offset := common.QueryOffset(config)

	log_level := zerolog.ErrorLevel
	log_level = zerolog.Level(config.DebugConfig.LogLevel)

	zerolog.TimestampFunc = func() time.Time {
		return common.GetCurrentTime(ntp_offset)
	}

	log.Logger = zerolog.New(file).
		Level(log_level).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Info().Msg("Start")

	if len(os.Args) >= 2 && bus.IsEventDefined(os.Args[1]) {
		checkErr(droplet.Start(config, ntp_offset))
		os.Exit(0)
	}

	fmt.Printf("EdgeVPN Token: %s\n", config.EdgeVPNToken)
	fmt.Printf("EdgeVPN PrivateKey: %s\n", config.PrivateKey)
	fmt.Printf("EdgeVPN PublicKey: %s\n", config.PublicKey)
	fmt.Printf("Version: %s\n", common.GetVersionInformation())
	fmt.Printf("IsDebugEnabled: %+v\n", config.DebugConfig)
	fmt.Printf("Timeserver: %s\n", config.NTPServer)

}

func checkErr(err error) {
	if err != nil {
		log.Err(err).Msg("Error occured")
		fmt.Println(err)
		os.Exit(1)
	}
}
