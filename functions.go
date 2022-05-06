/* Contains all of the functions of the program */

package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func errorRes(w http.ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(jsonErrorRes{
		Error: error,
	})
	if err != nil {
		w.Write([]byte(error))
	}
}

func makeTorrentSpec(infohash string, displayname string, trackers []string) *torrent.TorrentSpec {
	spec := torrent.TorrentSpec{}
	spec.InfoHash = metainfo.NewHashFromHex(infohash)
	spec.DisplayName = displayname
	for _, tracker := range trackers {
		spec.Trackers = append(spec.Trackers, []string{tracker})
	}
	return &spec
}

func decodeBody(w http.ResponseWriter, body io.Reader, v any) error {
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		errorRes(w, "JSON Decoder error", http.StatusInternalServerError)
	}
	return err
}
