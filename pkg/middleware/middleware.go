package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/thingful/big-iot-gateway/pkg/log"
)

// authMiddleware is a middleware instance that exposes functionality to
// validate incoming requests for the presence of a valid JWT provided by the
// marketplace.
type auth struct {
	secret []byte
}

// NewAuthMiddleware initializes our authMiddleware instance,
// converting the string secret into a base64 byte slice. If this decoding fails
// it returns an error, else the initialized middleware instance.
func NewAuth(s string) (*auth, error) {
	/*
		secret := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
		base64.StdEncoding.Encode(secret, []byte(s))
	*/
	return &auth{
		secret: []byte(s),
	}, nil
}

func (a *auth) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := getToken(r)
		if err != nil {
			http.Error(w, "Unable to read token", http.StatusBadRequest)
			log.Log("Unable to read token")
			return
		}
		if token != string(a.secret) {
			http.Error(w, "", http.StatusUnauthorized)
			log.Log("token", token, "secret", string(a.secret))
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// getToken extracts the token string from the request or returns an error
func getToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	reqToken = splitToken[1]
	if reqToken == "" {
		return "", errors.New("no auth token!")
	}

	return strings.Replace(reqToken, " ", "", -1), nil
}
