package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
)

// Endpoint handler for torrent adding to client
func apiAddTorrent(w http.ResponseWriter, r *http.Request) {
	var t *torrent.Torrent
	var spec *torrent.TorrentSpec = nil
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

	if spec == nil {
		errorRes(w, "No torrent provided", http.StatusNotFound)
		return
	}

	var terr error
	t, terr = btEngine.addTorrent(spec, false)
	if terr != nil {
		errorRes(w, "Torrent add error: "+terr.Error(), http.StatusInternalServerError)
		return
	}

	/* Creates the response body*/
	res := createAddTorrentRes(t)
	encodeRes(w, &res)
}

// Endpoint for selecting which file/s to download
func apiTorrentSelectFile(w http.ResponseWriter, r *http.Request) {
	res := apiTorrentSelectFileRes{}

	/* Parse the request body to apiTorrentSelectFileBody */
	body := apiTorrentSelectFileBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Check if no provided files */
	if !body.AllFiles && len(body.Files) < 1 {
		errorRes(w, "No files provided", http.StatusNotFound)
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
		// Empties the Files slice to prevent the execution of the code below when AllFiles if toggled
		body.Files = nil

		// Starts download for all files in the torrent
		/* Go through the selected files to append its info to the response */
		for _, f := range t.Files() {
			f.SetPriority(torrent.PiecePriorityNormal)
			saveSpecFile(t.InfoHash().String(), f.DisplayPath(), f.Priority())
			res.Files = append(res.Files, apiTorrentSelectFileResFiles{
				FileName: f.DisplayPath(),
				Stream:   createFileLink(t.InfoHash().String(), f.DisplayPath(), false),
				Download: createFileLink(t.InfoHash().String(), f.DisplayPath(), true),
			})
		}
	}

	// If specific files are selected
	for _, f := range body.Files {
		/* Get the handle of the torrent file from its DisplayPath */
		tf, tferr := getTorrentFile(t, f)
		if tferr != nil {
			continue
		}

		// Starts download of said torrent file
		tf.SetPriority(torrent.PiecePriorityNormal)

		// Save the filename to the DB for persistence
		saveSpecFile(t.InfoHash().String(), tf.DisplayPath(), tf.Priority())

		/* Go through the selected files to append its info to the response */
		res.Files = append(res.Files, apiTorrentSelectFileResFiles{
			FileName: tf.DisplayPath(),
			Stream:   createFileLink(t.InfoHash().String(), tf.DisplayPath(), false),
			Download: createFileLink(t.InfoHash().String(), tf.DisplayPath(), true),
		})
	}

	encodeRes(w, &res)
}

// Endpoint for setting the file/s priority
func apiTorrentPriorityFile(w http.ResponseWriter, r *http.Request) {
	res := apiTorrentPriorityFileRes{}

	/* Parse the request body to apiTorrentSelectFileBody */
	body := apiTorrentPriorityFileBody{}
	if decodeBody(w, r.Body, &body) != nil {
		return
	}

	/* Check if no provided files */
	if !body.AllFiles && len(body.Files) < 1 {
		errorRes(w, "No files provided", http.StatusNotFound)
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
	res.Priority = body.Priority

	// Parse the priority from the request body
	var selectedPriority torrent.PiecePriority
	switch strings.ToLower(body.Priority) {
	case "none":
		selectedPriority = torrent.PiecePriorityNone
	case "normal":
		selectedPriority = torrent.PiecePriorityNormal
	case "high":
		selectedPriority = torrent.PiecePriorityHigh
	case "readahead":
		selectedPriority = torrent.PiecePriorityReadahead
	}

	/* Set file priority for selected files */

	// If AllFiles is toggled
	if body.AllFiles {
		// Empties the Files slice to prevent the execution of the code below when AllFiles if toggled
		body.Files = nil

		/* Go through the selected files to append its info to the response */
		for _, f := range t.Files() {

			// Set priority of said torrent file to none essentially disabling its download
			f.SetPriority(selectedPriority)

			// Save the filename to the DB for persistence
			saveSpecFile(t.InfoHash().String(), f.DisplayPath(), f.Priority())

			/* Go through the all files to append its info to the response */
			res.Files = append(res.Files, apiTorrentPriorityFileResFiles{
				FileName: f.DisplayPath(),
			})
		}
	}

	// If specific files are selected
	for _, f := range body.Files {
		/* Get the handle of the torrent file from its DisplayPath */
		tf, tferr := getTorrentFile(t, f)
		if tferr != nil {
			continue
		}

		// Set priority of said torrent file to none essentially disabling its download
		tf.SetPriority(selectedPriority)

		// Save the filename to the DB for persistence
		saveSpecFile(t.InfoHash().String(), tf.DisplayPath(), tf.Priority())

		/* Go through the selected files to append its info to the response */
		res.Files = append(res.Files, apiTorrentPriorityFileResFiles{
			FileName: tf.DisplayPath(),
		})
	}

	encodeRes(w, &res)
}

// Endpoint for streaming a file
func apiStreamTorrentFile(w http.ResponseWriter, r *http.Request) {
	// Get infohash and filename variables
	vars := mux.Vars(r)

	/* Get torrent handle from infohash */
	t, err := btEngine.getTorrHandle(vars["infohash"])
	if err != nil {
		errorRes(w, err.Error(), http.StatusNotFound)
		return
	}

	/* Unescape given filename */
	fn, fnerr := url.QueryUnescape(vars["file"])
	if fnerr != nil {
		errorRes(w, "Filename unescaping error: "+fnerr.Error(), http.StatusInternalServerError)
		return
	}

	/* Get torrent file handle from filename */
	f, ferr := getTorrentFile(t, fn)
	if ferr != nil {
		errorRes(w, ferr.Error(), http.StatusNotFound)
		return
	}

	/* Make torrent file reader for streaming */
	reader := f.NewReader()
	defer reader.Close()
	// Set the buffer to 1% of the file size
	reader.SetReadahead(f.Length() / 100)
	// Send the reader as HTTP response
	http.ServeContent(w, r, f.DisplayPath(), time.Now(), reader)
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
	rmerr := btEngine.dropTorrent(ih, body.RemoveFiles)
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

		/* Overwrite tlist with only the selected torrent's handle */
		templist := make(map[string]*torrentHandle)
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
		tstats.DownloadSpeed = v.DlSpeedReadable
		tstats.UploadSpeed = v.UlSpeedReadable
		tstats.Progress = calcTorrentProgress(v.Torrent)

		/* Setting the peers info */
		for _, peer := range v.Torrent.PeerConns() {
			paddr := peer.Peer.RemoteAddr.String()
			pcli, ok := peer.PeerClientName.Load().(string)
			if !ok {
				pcli = "NOTPROVIDED"
			}

			tstats.Peers = append(tstats.Peers, apiTorrentStatsPeersInfo{
				PeerAddr:   paddr,
				PeerClient: pcli,
			})
		}

		/* Setting the files available in the torrent */
		for _, tf := range v.Torrent.Files() {
			tfname := tf.DisplayPath()
			tfbc := tf.BytesCompleted()
			tflen := tf.Length()
			curf := apiTorrentStatsTorrentsFiles{
				FileName:           tfname,
				FileSizeBytes:      int(tflen),
				FileSizeReadable:   humanize.Bytes(uint64(tflen)),
				DownloadedBytes:    int(tfbc),
				DownloadedReadable: humanize.Bytes(uint64(tfbc)),
				Priority:           torrentPriorityToString(tf.Priority()),
			}
			if tf.BytesCompleted() > 0 {
				curf.Stream = createFileLink(tstats.InfoHash, tfname, false)
				curf.Download = createFileLink(tstats.InfoHash, tfname, true)
			}
			tstats.Files = append(tstats.Files, curf)
		}

		/* Append it response body */
		res.Torrents = append(res.Torrents, tstats)
	}

	/* Send response */
	encodeRes(w, &res)
}

func apiDownloadFile(w http.ResponseWriter, r *http.Request) {
	/* Get infohash and filename vars*/
	vars := mux.Vars(r)

	/* Get torrent handle from infohash */
	t, err := btEngine.getTorrHandle(vars["infohash"])
	if err != nil {
		errorRes(w, err.Error(), http.StatusNotFound)
		return
	}

	/* Unescape given filename */
	fn, fnerr := url.QueryUnescape(vars["file"])
	if fnerr != nil {
		errorRes(w, "Filename unescaping error: "+fnerr.Error(), http.StatusInternalServerError)
		return
	}

	/* Get torrent file handle from filename */
	f, ferr := getTorrentFile(t, fn)
	if ferr != nil {
		errorRes(w, ferr.Error(), http.StatusNotFound)
		return
	}

	/* Check if file is finished downloading */
	if f.BytesCompleted() != f.Length() {
		errorRes(w, "File is not completed", http.StatusAccepted)
		return
	}

	/* Set Content-Disposition as f.DisplayPath() */
	w.Header().Add("Content-Disposition", "attachment; filename=\""+safenDisplayPath(f.DisplayPath())+"\"")

	/* Send file as response */
	reader := f.NewReader()
	defer reader.Close()
	http.ServeContent(w, r, f.DisplayPath(), time.Now(), reader)
}

func apiAddTorrentFile(w http.ResponseWriter, r *http.Request) {
	/* Gets file from form */
	torrfile, _, err := r.FormFile("torrent")
	if err != nil {
		errorRes(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer torrfile.Close()

	/* Loads torrent file to the BitTorrent client */
	/* Loads the torrent file as a MetaInfo */
	mi, mierr := metainfo.Load(torrfile)
	if mierr != nil {
		errorRes(w, mierr.Error(), http.StatusInternalServerError)
		return
	}
	/* Makes torrent spec from given MetaInfo */
	spec, specerr := torrent.TorrentSpecFromMetaInfoErr(mi)
	if specerr != nil {
		errorRes(w, specerr.Error(), http.StatusInternalServerError)
		return
	}
	/* Adds torrent spec to the BitTorrent client */
	t, terr := btEngine.addTorrent(spec, false)
	if terr != nil {
		errorRes(w, terr.Error(), http.StatusInternalServerError)
		return
	}

	/* Create response */
	res := createAddTorrentRes(t)
	encodeRes(w, &res)
}

// Make playlist for direct streaming of magnet link or torrent files
func apiDirectPlay(w http.ResponseWriter, r *http.Request) {
	/* Parse magnet link from query */
	magnet := r.URL.Query().Get("magnet")
	displayName := r.URL.Query().Get("dn")
	trackers, trackersOk := r.URL.Query()["tr"]
	files, filesOk := r.URL.Query()["file"]

	// Check if magnet link is present
	if magnet == "" {
		errorRes(w, "Invalid magnet link", http.StatusBadRequest)
		return
	}

	// Parses magnet link to add display name and trackers from query
	parsedMagnet, parseMagnetErr := metainfo.ParseMagnetUri(magnet)
	if parseMagnetErr != nil {
		errorRes(w, "Parsing magnet link error", http.StatusInternalServerError)
		return
	}
	if displayName != "" {
		parsedMagnet.DisplayName = displayName
	}
	if trackersOk {
		parsedMagnet.Trackers = trackers
	}

	// Parse magnet link to torrent spec
	spec, specErr := torrent.TorrentSpecFromMagnetUri(parsedMagnet.String())
	if specErr != nil {
		errorRes(w, "Creating torrent spec error", http.StatusInternalServerError)
		return
	}

	// Add torrent spec to BT engine
	t, addTorrentErr := btEngine.addTorrent(spec, false)
	if addTorrentErr != nil {
		errorRes(w, "Adding torrent error", http.StatusInternalServerError)
		return
	}

	/* Create the playlist file with the selected files */
	w.Header().Set("Content-Disposition", "attachment; filename=\""+t.InfoHash().String()+".m3u\"")
	playList := "#EXTM3U\n"

	httpScheme := "http"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		httpScheme = r.Header.Get("X-Forwarded-Proto")
	}

	if !filesOk {
		for _, file := range t.Files() {
			file.SetPriority(torrent.PiecePriorityNormal)
			saveSpecFile(t.InfoHash().String(), file.DisplayPath(), file.Priority())
			playList += appendFilePlaylist(httpScheme, r.Host, t.InfoHash().String(), file.DisplayPath())
		}
	}

	for _, file := range files {
		torrentFile, getTorretnFileErr := getTorrentFile(t, file)
		if getTorretnFileErr != nil {
			continue
		}
		torrentFile.SetPriority(torrent.PiecePriorityNormal)
		saveSpecFile(t.InfoHash().String(), torrentFile.DisplayPath(), torrentFile.Priority())
		playList += appendFilePlaylist(httpScheme, r.Host, t.InfoHash().String(), torrentFile.DisplayPath())
	}

	w.Write([]byte(playList))
}
