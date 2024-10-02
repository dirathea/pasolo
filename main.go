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
		e.Logger.Error("Failed to create WebAuthn", err)
	}

	var staticUser *user.User

	// Static user
	e.Logger.Debugf("Key: %v", config.EncyptionKey)
	if staticUser, err = user.LoadUser(); err != nil {
		fmt.Println("User does not exist, creating a new one")
		newUser := user.GetUser()
		staticUser = newUser.(*user.User)
		staticUser.Persist()
	}

	e.GET("/auth/register", func(c echo.Context) error {
		e.Logger.Debug("GET /auth/register")
		e.Logger.Debug("Generating registration session")
		options, sessionData, err := webAuthn.BeginRegistration(staticUser)
		if err != nil {
			e.Logger.Error("Failed to generate registration session", err)
			return c.JSON(500, err)
		}

		e.Logger.Debug("Storing registration session")
		if err := session.Store(sessionData); err != nil {
			e.Logger.Error("Failed to store registration session", err)
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/register", func(c echo.Context) error {
		e.Logger.Debug("POST /auth/register")
		e.Logger.Debug("Loading session data")
		sessionData, err := session.Load()
		if err != nil {
			e.Logger.Error("Failed to load session data", err)
			return c.JSON(500, err)
		}

		e.Logger.Debug("Parsing request body")
		// clone request body to different request variable
		authRequest := c.Request().Clone(c.Request().Context())
		var body map[string]interface{}
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			e.Logger.Debug("Failed to decode request body", err)
			return c.JSON(400, err)
		}

		// check if password is provided
		pass, ok := body["password"]
		if !ok {
			e.Logger.Error("missing password attribute")
			return c.JSON(400, "missing password attribute")
		}

		// check if password is correct
		if register.Verify(pass.(string)) == false {
			e.Logger.Error("incorrect password")
			return c.JSON(400, "incorrect password")
		}

		authCred, ok := body["credential"]
		if !ok {
			e.Logger.Error("missing credential attribute")
			return c.JSON(400, "missing credential attribute")
		}

		e.Logger.Debug("Parsing credential")
		credentialBytes, err := json.Marshal(authCred)
		if err != nil {
			e.Logger.Error("Failed to marshal credential", err)
			return c.JSON(400, err)
		}

		e.Logger.Debug("Finishing registration")
		authRequest.Body = io.NopCloser(bytes.NewReader(credentialBytes))
		credential, err := webAuthn.FinishRegistration(staticUser, *sessionData, authRequest)
		if err != nil {
			e.Logger.Error("Registration Failed", err)
			return c.JSON(500, err)
		}

		if err := session.Delete(); err != nil {
			e.Logger.Error("Failed to remove registration session", err)
		}

		e.Logger.Debug("Adding credential to user")
		staticUser.AddCredential(*credential)
		if err := staticUser.Persist(); err != nil {
			e.Logger.Error("Failed to persist user", err)
			return c.JSON(500, err)
		}

		return c.JSON(200, credential)
	})

	e.GET("/auth/login", func(c echo.Context) error {
		e.Logger.Debug("GET /auth/login")
		e.Logger.Debug("Generating login session")
		options, sessionData, err := webAuthn.BeginLogin(staticUser)
		if err != nil {
			e.Logger.Error("Failed to generate login session", err)
			return c.JSON(500, err)
		}
		e.Logger.Debug("Storing Login session")
		if err := session.Store(sessionData); err != nil {
			e.Logger.Error("Failed to store login session", err)
			return c.JSON(500, err)
		}

		return c.JSON(200, options.Response)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		e.Logger.Debug("POST /auth/login")
		e.Logger.Debug("Loading session data")
		sessionData, err := session.Load()
		if err != nil {
			return c.JSON(500, err)
		}
		e.Logger.Debug("Session data loaded")

		credentials, err := webAuthn.FinishLogin(staticUser, *sessionData, c.Request())
		if err != nil {
			e.Logger.Error("Login Failed", err)
			return c.JSON(500, err)
		}
		if err := session.Delete(); err != nil {
			e.Logger.Error("Failed to remove login session", err)
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
