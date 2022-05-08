package main

import (
	"net/http"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/gorilla/mux"
)

// Endpoint handler for torrent adding to client
func apiAddTorrent(w http.ResponseWriter, r *http.Request) {
	var t *torrent.Torrent
	var spec *torrent.TorrentSpec
	/* Decodes the request body */
	body := apiAddTorrentBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Parses the inputs */
	// If magnet link is present
	if body.Magnet != "" {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(body.Magnet)
		if err != nil {
			errorRes(w, "Magnet decoding error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// If manual metainfo is present
	if body.Magnet == "" && body.InfoHash != "" && body.DisplayName != "" {
		spec = makeTorrentSpec(body.InfoHash, body.DisplayName, body.Trackers)
	}

	var terr error
	t, terr = btEngine.addTorrent(spec, false)
	if terr != nil {
		errorRes(w, "Torrent add error: "+terr.Error(), http.StatusInternalServerError)
		return
	}

	/* Creates the response body*/
	res := apiAddTorrentRes{
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

// Endpoint for selecting which file/s to download
func apiTorrentSelectFile(w http.ResponseWriter, r *http.Request) {
	res := apiTorrentSelectFileRes{}

	/* Parse the request body to apiTorrentSelectFileBody */
	body := apiTorrentSelectFileBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Gets torrent handler from client */
	t, err := btEngine.getTorrHandle(body.InfoHash)
	if err != nil {
		errorRes(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/* Create the response body */
	res.InfoHash = t.InfoHash().String()
	res.Name = t.Name()

	/* Initiate download for selected files */

	// If AllFiles is toggled
	if body.AllFiles {
		t.DownloadAll()
		for _, f := range t.Files() {
			saveSpecFile(t.InfoHash().String(), f.DisplayPath())
			res.Files = append(res.Files, apiTorrentSelectFileResFiles{
				FileName: f.DisplayPath(),
				Link:     createFileLink(t.InfoHash().String(), f.DisplayPath()),
			})
		}
	}

	// If specific files are selected
	for _, f := range body.Files {
		tf, tferr := getTorrentFile(t, f)
		if tferr != nil {
			continue
		}
		saveSpecFile(t.InfoHash().String(), tf.DisplayPath())
		res.Files = append(res.Files, apiTorrentSelectFileResFiles{
			FileName: tf.DisplayPath(),
			Link:     createFileLink(t.InfoHash().String(), tf.DisplayPath()),
		})
	}

	encodeRes(w, &res)
	return
}

// Endpoint for streaming a file
func apiStreamTorrentFile(w http.ResponseWriter, r *http.Request) {
	// Get infohash and filename variables
	vars := mux.Vars(r)

	/* Get torrent handle from infohash */
	t, err := btEngine.getTorrHandle(vars["infohash"])
	if err != nil {
		errorRes(w, err.Error(), http.StatusNotFound)
	}

	/* Get torrent file handle from filename */
	f, ferr := getTorrentFile(t, vars["file"])
	if ferr != nil {
		errorRes(w, ferr.Error(), http.StatusNotFound)
	}

	/* Make torrent file reader for streaming */
	reader := f.NewReader()
	defer reader.Close()
	// Set the buffer to 1% of the file size
	reader.SetReadahead(f.Length() / 100)
	// Send the reader as HTTP response
	http.ServeContent(w, r, f.DisplayPath(), time.Now(), reader)
	return
}

// Endpoint for removing a torrent
func apiRemoveTorrent(w http.ResponseWriter, r *http.Request) {
	/* Parses the request body to apiRemoveTorrent */
	body := apiRemoveTorrentBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Getting the torrent handle */
	t, terr := btEngine.getTorrHandle(body.InfoHash)
	if terr != nil {
		errorRes(w, terr.Error(), http.StatusNotFound)
		return
	}

	/* Saving of variables for response body */
	tname := t.Name()
	ih := t.InfoHash().String()

	/* Remover function */
	btEngine.removeTorrentHandle(ih)
	rmerr := removeSpec(ih)
	if rmerr != nil {
		errorRes(w, "Torrent removal error: "+rmerr.Error(), http.StatusInternalServerError)
		return
	}

	/* Creating response body */
	res := apiRemoveTorrentRes{
		Name:     tname,
		InfoHash: ih,
	}
	encodeRes(w, &res)
	return
}

// Torrent stats endpoint
func apiTorrentStats(w http.ResponseWriter, r *http.Request) {
	/* Get infohash variable from the request */
	vars := mux.Vars(r)
	res := apiTorrentStasRes{}

	/* Variables */
	tlist := btEngine.Torrents
	ih := vars["infohash"]

	/* If provided with infohash */
	if ih != "" {
		/* Check if infohash is valid */
		_, terr := btEngine.getTorrHandle(ih)
		if terr != nil {
			errorRes(w, terr.Error(), http.StatusNotFound)
			return
		}

		Info.Printf("hey")
		/* Overwrite tlist with only the selected torrent's handle */
		templist := make(map[string]torrentHandle)
		templist[ih] = btEngine.Torrents[ih]
		tlist = templist
	}

	/* Go through the tlist */
	for _, v := range tlist {
		tstats := apiTorrentStasResTorrents{}

		/* Setting main stats */
		tstats.Name = v.Torrent.Name()
		tstats.InfoHash = v.Torrent.InfoHash().String()
		tstats.TotalPeers = v.Torrent.Stats().TotalPeers
		tstats.ActivePeers = v.Torrent.Stats().ActivePeers
		tstats.PendingPeers = v.Torrent.Stats().PendingPeers
		tstats.HalfOpenPeers = v.Torrent.Stats().HalfOpenPeers
		tstats.DownloadSpeedBytes = int(v.DlSpeedBytes)
		tstats.DownloadSpeedReadable = v.DlSpeedReadable

		/* Setting the files available in the torrent */
		for _, tf := range v.Torrent.Files() {
			tfname := tf.DisplayPath()
			tfsize := int(tf.Length())
			tstats.Files.OnTorrent = append(tstats.Files.OnTorrent, apiTorrentFiles{
				FileName:      tfname,
				FileSizeBytes: tfsize,
			})
			if tf.BytesCompleted() > 0 {
				tstats.Files.OnDisk = append(tstats.Files.OnDisk, apiTorrentStatsTorrentsFilesOnDisk{
					FileName:        tfname,
					FileSizeBytes:   tfsize,
					BytesDownloaded: int(tf.BytesCompleted()),
					Link:            createFileLink(tstats.InfoHash, tfname),
				})
			}
		}

		/* Append it response body */
		res.Torrents = append(res.Torrents, tstats)
	}

	/* Send response */
	encodeRes(w, &res)
	return
}
