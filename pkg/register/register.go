package register

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/dirathea/pasolo/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

func passwordPath() string {
	config := config.LoadConfig()
	return path.Join(config.Store.DataDir, "password.txt")
}

func Verify(password string) bool {
	file, err := os.Open(passwordPath())
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	hashedPassword := make([]byte, 60)
	if _, err := file.Read(hashedPassword); err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		fmt.Println("Error comparing password:", err)
		return false
	}

	return true
}

func generateRandomPassword(i int) string {
	b := make([]byte, i)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error generating random password:", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

func Init() {
	if _, err := os.Stat(passwordPath()); os.IsNotExist(err) {
		// generate password and hash it
		// store it in a file
		password := generateRandomPassword(12)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("Error hashing password:", err)
			return
		}

		file, err := os.Create(passwordPath())
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		if _, err := file.Write(hashedPassword); err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		if err := file.Sync(); err != nil {
			fmt.Println("Error syncing file:", err)
			return
		}
		fmt.Println("Store this password securely, and use it to register your passkey. This password will not be displayed again.")
		fmt.Println(password)
	}
}
