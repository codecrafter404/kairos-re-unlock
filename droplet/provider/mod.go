package provider

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func GetResponse(config config.Config, offset time.Duration, stdin []byte, ctx context.Context) *pluggable.EventResponse {

	datachan := make(chan pluggable.EventResponse)

	go getAsyncNodePairResponse(config, datachan, offset, ctx)
	go getAsyncHttpsResponse(config, datachan, offset, ctx)
	go getDebugPassword(config, datachan)

	res := <-datachan

	if validatePassword(res, config, stdin) {
		return &res
	}
	log.Error().Any("response", res).Msg("Got invalid response")
	return nil
}

func validatePassword(event pluggable.EventResponse, conf config.Config, stdin []byte) bool {
	if conf.DebugConfig.BypassPasswordTest {
		return true
	}

	if event.Error != "" {
		return false
	}

	device := getDevice(stdin)
	if device == "" {
		log.Warn().Msg("Got no device; skipping password validation")
		return true
	}

	cmd := exec.Command("cryptsetup", "luksOpen", "--test-passphrase", device)
	input_pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open stdin pipe")

	}

	go func() {
		defer input_pipe.Close()
		io.WriteString(input_pipe, event.Data)
	}()

	res, err := cmd.CombinedOutput()
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			log.Error().Err(err).Msg("Exec error")
			return true
		}
		code := exit.ExitCode()
		if code == 2 {
			log.Error().Msg("Got negative test response")
			return false
		}
		log.Error().Err(err).Str("stdout", string(res)).Msg("Got execution error; allowing password")
		return true
	}

	return true
}

func getDevice(stdin []byte) string {
	log.Info().Str("input", string(stdin)).Msg("Got input")

	var event pluggable.Event
	err := json.Unmarshal(stdin, &event)
	if err != nil {
		log.Error().Err(err).Str("stdin", string(stdin)).Msg("Failed to unmarshall stdin")
		return ""
	}

	var partition Partition

	err = json.Unmarshal([]byte(event.Data), &partition)

	if err != nil {
		log.Err(err).Str("event.data", string(event.Data)).Msg("Failed to unmarshall data")
		return ""
	}
	res := "/dev/" + partition.Name
	log.Info().Str("device", res).Msg("Found device")

	return res
}
