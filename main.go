package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/gommon/log"

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
	if staticUser, err = user.LoadUser(); err != nil {
		fmt.Println("User does not exist, creating a new one")
		newUser := user.GetUser()
		staticUser = newUser.(*user.User)
		if err := staticUser.Persist(); err != nil {
			log.Fatal("Failed to persist user. Exiting.", err)
		}
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

		e.Logger.Debug("Parsing request body")
		// clone request body to different request variable

		bodybytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			e.Logger.Debug("Failed to read request body", err)
			return c.JSON(400, err)
		}

		var body register.RegisterRequest
		if err := json.Unmarshal(bodybytes, &body); err != nil {
			e.Logger.Debug("Failed to decode request body", err)
			return c.JSON(400, err)
		}

		e.Logger.Debug("Verifying password")
		// check if password is correct
		if register.Verify(body.Password) == false {
			e.Logger.Error("incorrect password")
			return c.JSON(400, "incorrect password")
		}

		e.Logger.Debug("Load Session")
		challenge, err := body.GetSessionChallenge()
		if err != nil {
			e.Logger.Error("Failed to get session challenge", err)
			return c.JSON(500, err)
		}
		sessionData, err := session.Load(challenge)
		if err != nil {
			e.Logger.Error("Failed to load session data", err)
			return c.JSON(500, err)
		}

		e.Logger.Debug("Finishing registration")
		credentialBytes, err := json.Marshal(body.Credential)
		if err != nil {
			e.Logger.Error("Failed to marshal request body", err)
			return c.JSON(500, err)
		}

		authRequest := c.Request().Clone(c.Request().Context())
		authRequest.Body = io.NopCloser(bytes.NewReader(credentialBytes))
		credential, err := webAuthn.FinishRegistration(staticUser, *sessionData, authRequest)
		if err != nil {
			e.Logger.Error("Registration Failed", err)
			return c.JSON(500, err)
		}

		if err := session.Delete(challenge); err != nil {
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

		e.Logger.Debug("Parsing request body")
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			e.Logger.Debug("Failed to read request body", err)
			return c.JSON(400, err)
		}

		var body session.ResponseSession
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			e.Logger.Debug("Failed to unmarshal request body", err)
			return c.JSON(400, err)
		}

		e.Logger.Debug("Parsing session data")
		challenge, err := body.GetSessionChallenge()
		if err != nil {
			e.Logger.Error("Failed to get session challenge", err)
			return c.JSON(500, err)
		}

		e.Logger.Debug("Loading session data")
		sessionData, err := session.Load(challenge)
		if err != nil {
			e.Logger.Error("Failed to load session data", err)
			return c.JSON(500, err)
		}

		e.Logger.Debug("Session data loaded")
		authRequest := c.Request().Clone(c.Request().Context())
		authRequest.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		credentials, err := webAuthn.FinishLogin(staticUser, *sessionData, authRequest)
		if err != nil {
			e.Logger.Error("Login Failed", err)
			return c.JSON(500, err)
		}
		if err := session.Delete(challenge); err != nil {
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
