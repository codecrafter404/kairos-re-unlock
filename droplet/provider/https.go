package provider

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/codecrafter404/kairos-re-unlock/common"
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
	"github.com/rs/zerolog/log"
)

func getAsyncHttpsResponse(config config.Config, channel chan<- pluggable.EventResponse) {
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

		password, err := passwd.GetPassword(config)

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

	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		privKey, err := common.ParsePublicKey(config.PublicKey)
		if err != nil {
			w.Write([]byte("Failed parse public key"))
			w.WriteHeader(http.StatusInternalServerError)
		}
		logs, err := os.ReadFile("/tmp/kcrypt-kairos-re-unlock.log")
		if err != nil {
			w.Write([]byte("Failed read logs"))
			w.WriteHeader(http.StatusInternalServerError)
		}
		encrypted, err := rsa.EncryptOAEP(crypto.SHA512.New(), rand.Reader, &privKey, logs, nil)
		if err != nil {
			w.Write([]byte("Failed to encrypt logs"))
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(base64.StdEncoding.EncodeToString(encrypted)))
		log.Info().Str("receiver", r.RemoteAddr).Msg("Send log")
	})

	log.Info().Msg("Listening on :505")
	err := http.ListenAndServe(":505", nil)
	if err != nil {
		log.Err(err).Msg("Failed to listen and serve")
	}
}
