package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	_ "github.com/joho/godotenv/autoload"

	"github.com/dirathea/pasolo/pkg/config"
	"github.com/dirathea/pasolo/pkg/cookie"
	"github.com/dirathea/pasolo/pkg/register"
	"github.com/dirathea/pasolo/pkg/session"
	"github.com/dirathea/pasolo/pkg/user"
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
	PersistFile = "user.json"
	Key         = "12345678901234567890123456789012"
)

func main() {

	register.Init()

	config := config.LoadConfig()

	e := echo.New()

	wconfig := &webauthn.Config{
		RPDisplayName: config.Passkey.DisplayName, // Display Name for your site
		RPID:          config.Server.Domain,       // Generally the FQDN for your site
		RPOrigins: []string{
			config.Passkey.Origin,
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

		if err := session.Store(sessionData); err != nil {
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/register", func(c echo.Context) error {
		e.Logger.Print("POST /auth/register")
		e.Logger.Print(c.Request().Body)
		e.Logger.Print("Loading session data")
		sessionData, err := session.Load()
		if err != nil {
			return c.JSON(500, err)
		}

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
		if register.Verify(pass.(string)) == false {
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

		if err := session.Delete(); err != nil {
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

		if err := session.Store(sessionData); err != nil {
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		e.Logger.Print("POST /auth/login")
		e.Logger.Print(c.Request().Body)
		e.Logger.Print("Loading session data")
		sessionData, err := session.Load()
		if err != nil {
			return c.JSON(500, err)
		}
		e.Logger.Print("Session data loaded", sessionData)

		credentials, err := webAuthn.FinishLogin(staticUser, *sessionData, c.Request())
		if err != nil {
			e.Logger.Print("Login Failed", err)
			return c.JSON(500, err)
		}
		if err := session.Delete(); err != nil {
			print(err)
		}

		cookie.SetCookie(c, staticUser)

		return c.JSON(200, credentials)
	})

	e.GET("/validate", func(c echo.Context) error {
		err := cookie.ValidateCookie(c, staticUser)
		if err != nil {
			return c.JSON(401, err)
		}
		return c.JSON(200, "OK")
	})

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "frontend/build/client",
		HTML5: true,
	}))

	address := fmt.Sprintf(":%s", config.Server.Port)

	e.Logger.Fatal(e.Start(address))
}
