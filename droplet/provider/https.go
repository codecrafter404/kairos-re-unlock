package provider

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func getAsyncHttpsResponse(config config.Config, channel chan<- pluggable.EventResponse, offset time.Duration) {
	srv := http.Server{Addr: ":505"}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthy"))
	})

	http.HandleFunc("/unlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var passwd common.Payload

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&passwd)
		if err != nil {
			log.Err(err).Msg("Failed to decode json body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		password, err := passwd.GetPassword(config, offset)

		if err != nil {
			log.Err(err).Msg("Failed to get password")
			w.WriteHeader(http.StatusInternalServerError)
			return

		}

		channel <- pluggable.EventResponse{
			Data:  password,
			Error: "",
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("received"))
		go func() {
			log.Info().Msg("Shutting down http server")
			srv.Shutdown(r.Context())
			http.DefaultServeMux = http.NewServeMux()
		}()
	})

	if config.IsDebugEnabled() {
		http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
			logs, err := os.ReadFile("/tmp/kcrypt-kairos-re-unlock.log")
			if err != nil {
				w.Write([]byte("Failed read logs"))
				log.Err(err).Msg("Failed to read logs")
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(logs)
			log.Info().Str("receiver", r.RemoteAddr).Msg("Send log")
		})
	}

	log.Info().Msg("Listening on :505")
	err := srv.ListenAndServe()
	if err != nil {
		log.Err(err).Msg("Failed to listen and serve")
	}
}
