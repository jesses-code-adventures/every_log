package endpoints

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

type EndpointHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeJson(w http.ResponseWriter, r *http.Request)
}

func BasicValidateRequest(w http.ResponseWriter, r *http.Request) error {
	accept := r.Header.Get("Accept")
	if accept != "application/json" {
		http.Error(w, "Invalid Accept Header", http.StatusBadRequest)
		return errors.New("Accept")
	}
	apiKey := r.Header.Get("x-api-key")
	authorization := r.Header.Get("Authorization")
	if !(r.Method == http.MethodPost) && !(r.RequestURI == "/user") && !(r.RequestURI == "/authenticate") && apiKey == "" && authorization == "" {
		http.Error(w, "Missing authorization header and api key", http.StatusUnauthorized)
		return errors.New("Auth")
	}
	return nil
}

// Taken from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
