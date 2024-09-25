package session

import (
	"encoding/json"
	"os"
	"path"

	"github.com/dirathea/pasolo/pkg/config"
	"github.com/go-webauthn/webauthn/webauthn"
)

func sessionPath() string {
	config := config.LoadConfig()
	return path.Join(config.Store.DataDir, "session.json")
}

func Store(sessionData *webauthn.SessionData) error {
	file, err := os.Create(sessionPath())
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(sessionData); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}

func Load() (*webauthn.SessionData, error) {
	file, err := os.Open(sessionPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sessionData *webauthn.SessionData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionData); err != nil {
		return nil, err
	}

	return sessionData, nil
}

func Delete() error {
	return os.Remove(sessionPath())
}
