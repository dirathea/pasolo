package cookie

import (
	"net/http"
	"time"

	"github.com/dirathea/passkey-backend/pkg/config"
	"github.com/dirathea/passkey-backend/pkg/user"
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
		Name:    config.Cookie.Name,
		Value:   jwt,
		Expires: time.Now().Add(24 * time.Hour),
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
