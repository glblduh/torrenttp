package main

import (
	"net/http"
)

func apiAddTorrent(w http.ResponseWriter, r *http.Request) {
	body := apiAddMagnetBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

}
