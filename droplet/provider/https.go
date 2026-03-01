package provider

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func getAsyncHttpsResponse(config config.Config, channel chan<- pluggable.EventResponse, offset time.Duration, ctx context.Context) {
	srv := http.Server{Addr: ":505", BaseContext: func(l net.Listener) context.Context { return ctx }}

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

	errCh := make(chan error, 1)
	go func() {
		log.Info().Msg("Listening on :505")
		res := srv.ListenAndServe()
		errCh <- res
		log.Info().Msg("End")
	}()

	select {
	case res := <-errCh:
		if res != http.ErrServerClosed {
			log.Info().Msg("Context done")
			log.Error().Err(res).Msg("Server start failed")
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Server shutdown failed")
		}
		http.DefaultServeMux = http.NewServeMux()
		log.Info().Msg("Http server shutdown")
	}
}
