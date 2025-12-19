package provider

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
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
			log.Err(err).Msg("Failed to read logs")
			w.WriteHeader(http.StatusInternalServerError)
		}
		aesKey := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
			log.Err(err).Msg("Failed to generate AES key")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		block, err := aes.NewCipher(aesKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		nonce := make([]byte, aesGCM.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ciphertext := aesGCM.Seal(nonce, nonce, logs, nil)

		encrypted, err := rsa.EncryptOAEP(crypto.SHA512.New(), rand.Reader, &privKey, aesKey, nil)
		if err != nil {
			w.Write([]byte("Failed to encrypt logs"))
			log.Err(err).Msg("Failed to encrypt logs")
			w.WriteHeader(http.StatusInternalServerError)
		}

		res := common.LogsPayload{
			Nonce: base64.StdEncoding.EncodeToString(nonce),
			Data:  base64.RawStdEncoding.EncodeToString(ciphertext),
			Key:   base64.RawStdEncoding.EncodeToString(encrypted),
		}
		res_b, err := json.Marshal(res)
		if err != nil {
			w.Write([]byte("Failed to marhsal response"))
		}
		w.Write(res_b)
		log.Info().Str("receiver", r.RemoteAddr).Msg("Send log")
	})

	log.Info().Msg("Listening on :505")
	err := http.ListenAndServe(":505", nil)
	if err != nil {
		log.Err(err).Msg("Failed to listen and serve")
	}
}
