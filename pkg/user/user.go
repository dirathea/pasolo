package user

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dirathea/passkey-backend/pkg/config"
	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/crypto/nacl/secretbox"
)

type User struct {
	ID          string
	DisplayName string
	Credentials []webauthn.Credential
	Name        string
}

func (u *User) AddCredential(credential webauthn.Credential) {
	u.Credentials = append(u.Credentials, credential)
}

func (u *User) Persist(filePath string, key [32]byte) error {
	// Marshal the User struct to JSON
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}

	// Create a random nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	// Encrypt the data
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)

	// Write the encrypted data to the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(encrypted); err != nil {
		return err
	}

	return nil
}

func LoadUser(filePath string, key [32]byte) (*User, error) {
	// Read the encrypted data from the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encrypted, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt data")
	}

	// Unmarshal the decrypted data into a User struct
	var user User
	if err := json.Unmarshal(decrypted, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// WebAuthnCredentials implements webauthn.User.
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// WebAuthnDisplayName implements webauthn.User.
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnID implements webauthn.User.
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

// WebAuthnName implements webauthn.User.
func (u *User) WebAuthnName() string {
	return u.Name
}

func GetUser() webauthn.User {

	config := config.LoadConfig()

	return &User{
		ID:          config.User.ID,
		DisplayName: config.User.DisplayName,
		Name:        config.User.Name,
		Credentials: []webauthn.Credential{},
	}
}
