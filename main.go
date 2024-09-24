package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/dirathea/passkey-backend/pkg/config"
	"github.com/dirathea/passkey-backend/pkg/cookie"
	"github.com/dirathea/passkey-backend/pkg/user"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error
)

const (
	// SessionDataFile is the file where the session data is stored
	SessionDataFile = "sessionData.json"
	PersistFile     = "user.json"
	Key             = "12345678901234567890123456789012"
)

func storeSessionData(sessionData *webauthn.SessionData) error {
	file, err := os.Create(SessionDataFile)
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

func loadSessionData() (*webauthn.SessionData, error) {
	file, err := os.Open(SessionDataFile)
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

func deleteSessionData() error {
	return os.Remove(SessionDataFile)
}

func initRegisterPassword() {
	// generate password and hash it
	// store it in a file
	password := generateRandomPassword(12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}

	file, err := os.Create("password.txt")
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

func generateRandomPassword(i int) string {
	b := make([]byte, i)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error generating random password:", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

func checkPassword(password string) bool {
	file, err := os.Open("password.txt")
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

func main() {

	if _, err := os.Stat("password.txt"); os.IsNotExist(err) {
		initRegisterPassword()
	}

	config := config.LoadConfig()

	e := echo.New()

	origin := fmt.Sprintf("%s://%s", config.Server.Protocol, config.Server.Domain)
	if config.Server.Port != "" {
		origin = fmt.Sprintf("%s:%s", origin, config.Server.Port)
	}

	wconfig := &webauthn.Config{
		RPDisplayName: config.Passkey.DisplayName, // Display Name for your site
		RPID:          config.Server.Domain,       // Generally the FQDN for your site
		RPOrigins: []string{
			origin,
		}, // The origin URLs allowed for WebAuthn requests
	}
	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Println(err)
	}

	var staticUser *user.User

	// Static user
	keyBytes := [32]byte{}
	copy(keyBytes[:], config.EncyptionKey)
	e.Logger.Debugf("Key: %v", config.EncyptionKey)
	if staticUser, err = user.LoadUser(PersistFile, keyBytes); err != nil {
		fmt.Println("User does not exist, creating a new one")
		newUser := user.GetUser()
		staticUser = newUser.(*user.User)
		staticUser.Persist(PersistFile, [32]byte(keyBytes))
	}

	e.GET("/auth/register", func(c echo.Context) error {
		options, sessionData, err := webAuthn.BeginRegistration(staticUser)
		if err != nil {
			return c.JSON(500, err)
		}

		if err := storeSessionData(sessionData); err != nil {
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/register", func(c echo.Context) error {
		e.Logger.Print("POST /auth/register")
		e.Logger.Print(c.Request().Body)
		e.Logger.Print("Loading session data")
		sessionData, err := loadSessionData()
		if err != nil {
			return c.JSON(500, err)
		}
		e.Logger.Print("Session data loaded", sessionData)

		// clone request body to different request variable
		authRequest := c.Request().Clone(c.Request().Context())
		var body map[string]interface{}
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			e.Logger.Print("Failed to decode request body", err)
			return c.JSON(400, err)
		}

		// check if password is provided
		pass, ok := body["password"]
		if !ok {
			return c.JSON(400, "missing password attribute")
		}

		// check if password is correct
		if checkPassword(pass.(string)) == false {
			return c.JSON(400, "incorrect password")
		}

		authCred, ok := body["credential"]
		if !ok {
			return c.JSON(400, "missing credential attribute")
		}
		credentialBytes, err := json.Marshal(authCred)
		if err != nil {
			e.Logger.Print("Failed to marshal credential", err)
			return c.JSON(400, err)
		}
		authRequest.Body = io.NopCloser(bytes.NewReader(credentialBytes))

		credential, err := webAuthn.FinishRegistration(staticUser, *sessionData, authRequest)
		if err != nil {
			e.Logger.Print("Registration Failed", err)
			return c.JSON(500, err)
		}

		if err := deleteSessionData(); err != nil {
			print(err)
		}

		staticUser.AddCredential(*credential)
		if err := staticUser.Persist(PersistFile, keyBytes); err != nil {
			return c.JSON(500, err)
		}

		return c.JSON(200, credential)
	})

	e.GET("/auth/login", func(c echo.Context) error {
		options, sessionData, err := webAuthn.BeginLogin(staticUser)
		if err != nil {
			return c.JSON(500, err)
		}

		if err := storeSessionData(sessionData); err != nil {
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		e.Logger.Print("POST /auth/login")
		e.Logger.Print(c.Request().Body)
		e.Logger.Print("Loading session data")
		sessionData, err := loadSessionData()
		if err != nil {
			return c.JSON(500, err)
		}
		e.Logger.Print("Session data loaded", sessionData)

		credentials, err := webAuthn.FinishLogin(staticUser, *sessionData, c.Request())
		if err != nil {
			e.Logger.Print("Login Failed", err)
			return c.JSON(500, err)
		}
		if err := deleteSessionData(); err != nil {
			print(err)
		}

		cookie.SetCookie(c, staticUser)

		return c.JSON(200, credentials)
	})

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "frontend/build/client",
		HTML5: true,
	}))

	e.Logger.Fatal(e.Start(":8080"))
}
