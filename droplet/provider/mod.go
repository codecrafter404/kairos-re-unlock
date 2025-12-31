package provider

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func GetResponse(config config.Config, offset time.Duration) *pluggable.EventResponse {

	datachan := make(chan pluggable.EventResponse)

	go getAsyncNodePairResponse(config, datachan, offset)
	go getAsyncHttpsResponse(config, datachan, offset)
	go getDebugPassword(config, datachan)

	res := <-datachan

	if testPassword(res, config) {
		return &res
	}
	log.Error().Any("response", res).Msg("Got invalid response")
	return nil
}

func testPassword(event pluggable.EventResponse, conf config.Config) bool {
	if conf.DebugConfig.BypassPasswordTest {
		return true
	}

	if event.Error != "" {
		return false
	}

	device := getDevice()
	if device == "" {
		log.Warn().Msg("Got no device; skipping password validation")
		return true
	}

	cmd := exec.Command("cryptsetup", "luksOpen", "--test-passphrase", device)
	input_pipe, err := cmd.StdinPipe()
	log.Error().Err(err).Msg("Failed to open stdin pipe")

	log.Info().Msg("Running command")
	go func() {
		log.Info().Msg("sending text")
		defer input_pipe.Close()
		io.WriteString(input_pipe, event.Data+"\n")
		log.Info().Msg("sent text")
	}()

	log.Info().Msg("capture output")
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("stdout", string(res)).Msg("Got invalid test response")
		return false
	}

	return true
}

func getDevice() string {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read from stdin")
		return ""
	}
	log.Info().Str("input", string(input)).Msg("Got input")

	var event pluggable.Event
	err = json.Unmarshal(input, &event)
	if err != nil {
		log.Error().Err(err).Str("stdin", string(input)).Msg("Failed to unmarshall stdin")
		return ""
	}

	var discovery_password_payload DiscoveryPasswordPayload

	err = json.Unmarshal([]byte(event.Data), &discovery_password_payload)

	if err != nil {
		log.Err(err).Str("event.data", string(event.Data)).Msg("Failed to unmarshall data")
		return ""
	}
	if discovery_password_payload.Partition == nil {
		log.Warn().Str("data", string(input)).Msg("Got no partition data")
		return ""
	}

	//TODO: implement
	log.Info().Str("input", string(input)).Msg("INPUUUUUUTTT")

	return ""
}
