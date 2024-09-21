package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dirathea/passkey-backend/pkg/user"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func main() {
	e := echo.New()

	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",                     // Display Name for your site
		RPID:          "localhost",                       // Generally the FQDN for your site
		RPOrigins:     []string{"http://localhost:8080"}, // The origin URLs allowed for WebAuthn requests
	}
	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Println(err)
	}

	var staticUser *user.User

	// Static user
	keyBytes := [32]byte{}
	copy(keyBytes[:], Key)
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

		credential, err := webAuthn.FinishRegistration(staticUser, *sessionData, c.Request())
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

		return c.JSON(200, credentials)
	})

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "frontend/build/client",
		HTML5: true,
	}))

	e.Logger.Fatal(e.Start(":8080"))
}
