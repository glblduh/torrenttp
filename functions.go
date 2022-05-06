/* Contains all of the functions of the program */

package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// Function for sending error message as JSON response
func errorRes(w http.ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(jsonErrorRes{
		Error: error,
	})
	if err != nil {
		w.Write([]byte(error))
	}
}

// Compiles infohash, display name, and trackers to *torrent.TorrentSpec
func makeTorrentSpec(infohash string, displayname string, trackers []string) *torrent.TorrentSpec {
	spec := torrent.TorrentSpec{}
	spec.InfoHash = metainfo.NewHashFromHex(infohash)
	spec.DisplayName = displayname
	for _, tracker := range trackers {
		spec.Trackers = append(spec.Trackers, []string{tracker})
	}
	return &spec
}

// Decodes r.Body as JSON and automatically creates an error response if error
func decodeBody(w http.ResponseWriter, body io.Reader, v any) error {
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		errorRes(w, "JSON Decoder error", http.StatusInternalServerError)
	}
	return err
}

// Encodes a struct as JSON response and automatically creates an error response if error
func encodeRes(w http.ResponseWriter, v any) error {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		errorRes(w, "JSON Encoder error", http.StatusInternalServerError)
	}
	return err
}
