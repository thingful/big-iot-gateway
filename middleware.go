package main

import (
	"encoding/base64"
	"net/http"
)

// authMiddleware is a middleware instance that exposes functionality to
// validate incoming requests for the presence of a valid JWT provided by the
// marketplace.
type authMiddleware struct {
	secret []byte
}

// NewAuthMiddleware initializes our authMiddleware instance,
// converting the string secret into a base64 byte slice. If this decoding fails
// it returns an error, else the initialized middleware instance.
func NewAuthMiddleware(s string) (*authMiddleware, error) {
	secret, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return &authMiddleware{
		secret: secret,
	}, nil
}

func (a *authMiddleware) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err = getToken(r)
		if err != nil {
			http.Error(w, "Unable to read token", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// getToken extracts the token string from the request or returns an error
func getToken(r *http.Request) (string, error) {
	return "", nil
}
