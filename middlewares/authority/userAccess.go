package authority

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(echo.Context) bool

	// BeforeFunc defines a function which is executed just before the middleware.
	BeforeFunc func(echo.Context)

	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// BeforeFunc defines a function which is executed just before the middleware.
		BeforeFunc BeforeFunc

		// SuccessHandler defines a function which is executed for a valid token.
		SuccessHandler JWTSuccessHandler

		// ErrorHandler defines a function which is executed for an invalid token.
		// It may be used to define a custom JWT err.
		ErrorHandler JWTErrorHandler

		// Signing key to validate token.
		// Required.
		SigningKey interface{}

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// VersionHandler func(int, int) (bool, err)

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		keyFunc jwt.Keyfunc
	}

	// JWTSuccessHandler defines a function which is executed for a valid token.
	JWTSuccessHandler func(echo.Context, *JWTClaimsCustom) bool

	// JWTErrorHandler defines a function which is executed for an invalid token.
	JWTErrorHandler func(error) error

	jwtExtractor func(echo.Context) (string, error)
)

// JWTClaimsCustom 自定义jwt.Claims
type JWTClaimsCustom struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Permission int    `json:"permission"`
}

var (
	RoleVisitor    = "visitor"
	RoleAdmin      = "admin"
	RoleManager    = "manager"
	RoleSubManager = "sub_manager"
)

// Algorithms
const (
	AlgorithmHS256 = "HS256"
)

// Errors
var (
	ErrJWTMissing = echo.NewHTTPError(http.StatusBadRequest, "missing or malformed jwt")

	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		Skipper:       DefaultSkipper,
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup:   "header:" + echo.HeaderAuthorization,
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
	}
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(echo.Context) bool {
	return false
}

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" err.
// For missing token, it returns "400 - Bad Request" err.
//
// See: https://jwt.io/introduction
// See `JWTConfig.TokenLookup`
func JWT(key interface{}) echo.MiddlewareFunc {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

// JWTWithConfig returns a JWT auth middleware with config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.SigningKey == nil {
		panic("echo: jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return config.SigningKey, nil
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = jwtFromQuery(parts[1])
	case "cookie":
		extractor = jwtFromCookie(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			auth, err := extractor(c)
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(err)
				}
				return err
			}
			token := new(jwt.Token)

			claims := &JWTClaimsCustom{}
			token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)

			if err == nil && token.Valid {
				// Store user information from token into context.
				if config.SuccessHandler == nil || config.SuccessHandler(c, claims) {
					c.Set(config.ContextKey, claims)
					return next(c)
				}
				err = fmt.Errorf("invalid jwt")
			}
			if config.ErrorHandler != nil {
				return config.ErrorHandler(err)
			}
			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  "invalid or expired jwt",
				Internal: err,
			}
		}
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrJWTMissing
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", ErrJWTMissing
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", ErrJWTMissing
		}
		return cookie.Value, nil
	}
}

// 设置JWT
func setJWTCookie(c echo.Context, data JWTClaimsCustom) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	// Generate encoded token and send it as response.
	wToken, err := token.SignedString([]byte("JWTSecret"))
	if err != nil {
		logrus.WithError(err).WithField("data", data).Errorln("jwt token sign err")
		return errors.New("err")
	}
	c.SetCookie(getJWTCookie(wToken, time.Now().Add(time.Second*10)))

	return nil
}

// 获取JWT（统一格式）
func getJWTCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "JwtToken",
		Value:    value,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
	}
}

// Valid 实现jwt.Claims接口，验证数据有效性
func (j JWTClaimsCustom) Valid() error {
	if j.ID <= 0 {
		return fmt.Errorf("jwt param err")
	}
	return nil
}
