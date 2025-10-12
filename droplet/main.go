package main

import (
	"fmt"
	"os"

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

	log.Logger = zerolog.New(file).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	config, err := droplet.UnmarshalConfig()
	checkErr(err)

	if len(os.Args) >= 2 && bus.IsEventDefined(os.Args[1]) {
		checkErr(droplet.Start(config))
		os.Exit(0)
	}

	fmt.Printf("EdgeVPN Token: %s\n", config.EdgeVPNToken)
	fmt.Printf("EdgeVPN PrivateKey: %s\n", config.PrivateKey)
	fmt.Printf("EdgeVPN PublicKey: %s\n", config.PublicKey)

}

func checkErr(err error) {
	if err != nil {
		log.Err(err).Msg("Error occured")
		fmt.Println(err)
		os.Exit(1)
	}
}
