/* Contains global variables and structs of the program */

package main

import (
	"log"
	"os"
	"time"

	"github.com/anacrolix/torrent"
)

/* Variables */
var (
	/* BitTorrent client */
	btEngine btEng

	/* Loggers */
	// For information
	Info = log.New(os.Stderr, "["+time.Now().Format("2006/01/02 15:04:05")+"] [INFO] ", log.Lmsgprefix)
	// For non-critical errors
	Warn = log.New(os.Stderr, "["+time.Now().Format("2006/01/02 15:04:05")+"] [WARN] ", log.Lmsgprefix)
	// For critical errors
	Error = log.New(os.Stderr, "["+time.Now().Format("2006/01/02 15:04:05")+"] [ERROR] ", log.Lmsgprefix)
)

/* Structs */
type (
	// BitTorrent client struct
	btEng struct {
		Client       *torrent.Client
		ClientConfig *torrent.ClientConfig
	}

	// Struct for persistent spec
	persistentSpec struct {
		Trackers                 [][]string
		InfoHash                 string
		DisplayName              string
		Webseeds                 []string
		DhtNodes                 []string
		PeerAddrs                []string
		Sources                  []string
		DisableInitialPieceCheck bool
		DisallowDataUpload       bool
		DisallowDataDownload     bool
		AllFiles                 bool
		Files                    []string
	}

	jsonErrorRes struct {
		Error string `json:"error"`
	}

	apiAddMagnetBody struct {
		Magnet string `json:"magnet"`

		/*
			These are optional for manual adding of torrent spec
			The first priority is the magnet link rather than these
		*/
		InfoHash    string   `json:"infohash"`
		DisplayName string   `json:"displayname"`
		Trackers    []string `json:"trackers"`
	}

	apiAddMagnetRes struct {
		Name          string            `json:"name"`
		InfoHash      string            `json:"infohash"`
		TotalPeers    int               `json:"totalpeers"`
		ActivePeers   int               `json:"activepeers"`
		PendingPeers  int               `json:"pendingpeers"`
		HalfOpenPeers int               `json:"halfopenpeers"`
		Files         []apiTorrentFiles `json:"files"`
	}

	apiTorrentFiles struct {
		FileName      string `json:"filename"`
		FileSizeBytes int    `json:"filesizebytes"`
	}
)
