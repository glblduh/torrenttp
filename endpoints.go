package main

import (
	"net/http"

	"github.com/anacrolix/torrent"
)

// Endpoint handler for torrent adding to client
func apiAddTorrent(w http.ResponseWriter, r *http.Request) {
	var t *torrent.Torrent
	var spec *torrent.TorrentSpec
	/* Decodes the request body */
	body := apiAddMagnetBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Parses the inputs */
	// If magnet link is present
	if body.Magnet != "" {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(body.Magnet)
		if err != nil {
			errorRes(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// If manual metainfo is present
	if body.Magnet == "" && body.InfoHash != "" && body.DisplayName != "" {
		spec = makeTorrentSpec(body.InfoHash, body.DisplayName, body.Trackers)
		addTorrent(spec, false)
	}

	var terr error
	t, terr = addTorrent(spec, false)
	if terr != nil {
		errorRes(w, terr.Error(), http.StatusInternalServerError)
		return
	}

	/* Creates the response body*/
	res := apiAddMagnetRes{
		Name:          t.Name(),
		InfoHash:      t.InfoHash().String(),
		TotalPeers:    t.Stats().TotalPeers,
		ActivePeers:   t.Stats().ActivePeers,
		PendingPeers:  t.Stats().PendingPeers,
		HalfOpenPeers: t.Stats().HalfOpenPeers,
	}
	for _, f := range t.Files() {
		res.Files = append(res.Files, apiTorrentFiles{
			FileName:      f.DisplayPath(),
			FileSizeBytes: int(f.Length()),
		})
	}
	encodeRes(w, &res)
	return
}
