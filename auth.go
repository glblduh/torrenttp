package main

import (
	"net/http"
	"net/url"
	"os"
)

// Check if authentication is enabled
func checkAuthEnabled(isEnabled bool) {
	authEnabled = isEnabled
	// If authentication is disabled
	if !isEnabled {
		Warn.Println("Authentication is disabled")
		return
	}

	// Gets the key from env variable
	key, isValid := os.LookupEnv("TORRENTTPKEY")

	// Check if key is empty or unset
	if key == "" || !isValid {
		Error.Fatalln("Auth flag is enabled but TORRENTTPKEY env variable is empty or unset")
	}

	// Set the API key to the value of TORRENTTPKEY
	apiKey = key

	Info.Println("Authentication is enabled")
}

// Check for API key on the HTTP query
func checkAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from HTTP query
		key := r.URL.Query().Get("key")

		if authEnabled {
			// Unescape the API key
			unescapedKey, unescapeErr := url.QueryUnescape(key)
			if unescapeErr != nil {
				errorRes(w, "Error unescaping the API key", http.StatusInternalServerError)
				return
			}

			// Check if API key is valid
			if unescapedKey != apiKey {
				errorRes(w, "Key is not valid", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}
