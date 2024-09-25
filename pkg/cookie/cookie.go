package cookie

import (
	"net/http"
	"time"

	"github.com/dirathea/pasolo/pkg/config"
	"github.com/dirathea/pasolo/pkg/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func SetCookie(c echo.Context, user *user.User) {
	config := config.LoadConfig()
	jwt, err := generateJWTFromUser(user, []byte(config.Cookie.Secret), config.Server.Domain)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to generate JWT")
		return
	}
	cookie := &http.Cookie{
		Name:     config.Cookie.Name,
		Value:    jwt,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Domain:   config.Server.Domain,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	c.SetCookie(cookie)
}

func generateJWTFromUser(user *user.User, key []byte, domain string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.ID,
		Issuer:    domain,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateCookie(c echo.Context, user *user.User) error {
	config := config.LoadConfig()
	cookie, err := c.Cookie(config.Cookie.Name)
	if err != nil {
		return err
	}
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cookie.Secret), nil
	})
	if err != nil {
		return err
	}

	// Validate the token is not expired
	if !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims := token.Claims

	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if expiration.Before(time.Now()) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	userID, err := claims.GetSubject()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}
	if userID != user.ID {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid subject")
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if issuer != config.Server.Domain {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid issuer")
	}

	return nil
}
