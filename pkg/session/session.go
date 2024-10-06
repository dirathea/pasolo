package session

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/dirathea/pasolo/pkg/config"
	"github.com/go-webauthn/webauthn/webauthn"
)

type ResponseSession struct {
	Response struct {
		ClientDataJSON string `json:"clientDataJSON"`
	} `json:"response"`
}

func (r *ResponseSession) GetSessionChallenge() (string, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(r.Response.ClientDataJSON)
	if err != nil {
		return "", err
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(decoded, &session); err != nil {
		return "", err
	}

	return session.Challenge, nil
}

type sessionMap map[string]*webauthn.SessionData

func sessionPath() string {
	config := config.LoadConfig()
	return path.Join(config.Store.DataDir, "session.json")
}

func load() sessionMap {
	file, err := os.Open(sessionPath())
	if err != nil {
		return make(sessionMap)
	}
	defer file.Close()

	var sessions sessionMap
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessions); err != nil {
		return make(sessionMap)
	}

	return sessions
}

func Store(sessionData *webauthn.SessionData) error {
	var sessions sessionMap

	if _, err := os.Stat(sessionPath()); os.IsNotExist(err) {
		sessions = make(sessionMap)
	} else {
		sessions = load()
	}

	sessions[sessionData.Challenge] = sessionData

	file, err := os.Create(sessionPath())
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(sessions); err != nil {
		return err
	}

	return nil
}

func Load(challenge string) (*webauthn.SessionData, error) {
	sessions := load()
	sessionData, ok := sessions[challenge]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	return sessionData, nil
}

func Delete(challenge string) error {
	sessions := load()
	delete(sessions, challenge)

	file, err := os.Create(sessionPath())
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(sessions); err != nil {
		return err
	}

	return nil
}
