package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/thingful/big-iot-gateway/pkg/log"
	"github.com/thingful/bigiot"
	"goji.io/pat"
)

// authMiddleware is a middleware instance that exposes functionality to
// validate incoming requests for the presence of a valid JWT provided by the
// marketplace.
type auth struct {
	provider *bigiot.Provider
}

// NewAuth initializes our authMiddleware instance,
// converting the string secret into a base64 byte slice. If this decoding fails
// it returns an error, else the initialized middleware instance.
func NewAuth(p *bigiot.Provider) (*auth, error) {
	return &auth{
		provider: p,
	}, nil
}

func (a *auth) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := getToken(r)
		if err != nil {
			http.Error(w, "Missing Token", http.StatusBadRequest)
			log.Log("error", "Unable to read token")
			return
		}

		id, err := a.provider.ValidateToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Log("error", "non valid token")
			return
		}

		// we need to convert the full id into how we describe locally which is just the last part
		idParts := strings.Split(id, "-")
		if len(idParts) != 3 {
			log.Log("error", "id does not have enough parts")
		}

		offeringID := pat.Param(r, "offeringID")
		log.Log("debug-offeringID", offeringID)

		if idParts[2] != offeringID {
			log.Log("tokenID", idParts[2], "requestedID", offeringID, "error", "token id does not match requested")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// getToken extracts the token string from the request or returns an error
func getToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		return "", errors.New("no auth token")
	}
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) < 2 {
		return "", errors.New("no valid token")
	}
	reqToken = splitToken[1]
	if reqToken == "" {
		return "", errors.New("no auth token")
	}
	return strings.Replace(reqToken, " ", "", -1), nil
}
