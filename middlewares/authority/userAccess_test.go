package authority

import (
	"bytes"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"
	"time"
)

var (
	jwtContextKey = "TestContext"
	jwtCookieKey  = "TestCookie"
	TestJwt       = ""
	rw            http.ResponseWriter
)

func UserAccessCreate_Test(t *testing.T) {
	req, _ := http.NewRequest("get", "http://test.com", bytes.NewBuffer([]byte("")))

	c := echo.New().NewContext(req, rw)
	data := JWTClaimsCustom{
		ID:         1,
		Type:       RoleManager,
		Name:       "test_manager",
		Permission: 0,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	wToken, err := token.SignedString([]byte("JWTSecret"))
	if err != nil {
		logrus.WithError(err).WithField("data", data).Errorln("jwt token sign err")
	}
	c.SetCookie(getJWTCookie(wToken, time.Now().Add(time.Second*10)))
}

func UserAccessAuth_Test(t *testing.T) {
	fmt.Println(t.Name())
	JWTWithConfig(JWTConfig{
		SigningMethod: AlgorithmHS256,
		ContextKey:    jwtContextKey,
		TokenLookup:   "cookie:" + jwtCookieKey,
		Claims:        JWTClaimsCustom{},
		SigningKey:    []byte("JWTSecret"),
		SuccessHandler: func(c echo.Context, claims *JWTClaimsCustom) bool {
			// TODO: AUTH
			jwtCookie, err := c.Cookie(jwtCookieKey)
			if err == nil && jwtCookie.Value != "" {

				c.SetCookie(getJWTCookie(jwtCookie.Value, time.Now().Add(time.Second*10)))
			}
			return true
		},
		ErrorHandler: func(err error) error {
			logrus.WithError(err).Warningln("invalid token")
			return &echo.HTTPError{
				Code: http.StatusUnauthorized,
			}
		},
	})
}
