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

/* Structs for non-HTTP handlers */
type (
	// BitTorrent client struct
	btEng struct {
		Client       *torrent.Client
		ClientConfig *torrent.ClientConfig
		Torrents     map[string]*torrentHandle
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
		Files                    []string
	}

	// Holds non-native stats for *torrent.Torrent
	torrentHandle struct {
		/* Main handles */
		Torrent *torrent.Torrent
		Spec    *torrent.TorrentSpec

		/* Stats */
		DlSpeedBytes    int64
		DlSpeedReadable string
		UlSpeedBytes    int64
		UlSpeedReadable string

		/* Temporary */
		LastDlBytes int64
		LastUlBytes int64
	}
)

/* Structs of HTTP handlers */
type (
	// Response of JSON error
	jsonErrorRes struct {
		Error string `json:"error"`
	}

	// Expected request body to addTorrent
	apiAddTorrentBody struct {
		Magnet string `json:"magnet"`

		/*
			These are optional for manual adding of torrent spec
			The first priority is the magnet link rather than these
		*/
		InfoHash    string   `json:"infohash"`
		DisplayName string   `json:"displayname"`
		Trackers    []string `json:"trackers"`
	}

	// Expected response from addTorrent
	apiAddTorrentRes struct {
		Name          string            `json:"name"`
		InfoHash      string            `json:"infohash"`
		TotalPeers    int               `json:"totalpeers"`
		ActivePeers   int               `json:"activepeers"`
		PendingPeers  int               `json:"pendingpeers"`
		HalfOpenPeers int               `json:"halfopenpeers"`
		Files         []apiTorrentFiles `json:"files"`
	}

	// Struct for files in torrent
	apiTorrentFiles struct {
		FileName      string `json:"filename"`
		FileSizeBytes int    `json:"filesizebytes"`
	}

	// Expected request body to selectFile
	apiTorrentSelectFileBody struct {
		InfoHash string   `json:"infohash"`
		AllFiles bool     `json:"allfiles"`
		Files    []string `json:"files"`
	}

	// Expected response body from selectFile
	apiTorrentSelectFileRes struct {
		Name     string                         `json:"name"`
		InfoHash string                         `json:"infohash"`
		Files    []apiTorrentSelectFileResFiles `json:"files"`
	}

	// Struct for selectFile Files
	apiTorrentSelectFileResFiles struct {
		FileName string `json:"filename"`
		Stream   string `json:"stream"`
		Download string `json:"download"`
	}

	// Expected request body to removeTorrent
	apiRemoveTorrentBody struct {
		InfoHash    string `json:"infohash"`
		RemoveFiles bool   `json:"removefiles"`
	}

	// Expected response body from removeTorrent
	apiRemoveTorrentRes struct {
		Name     string `json:"name"`
		InfoHash string `json:"infohash"`
	}

	// Expected response body from torrentStats
	apiTorrentStasRes struct {
		Torrents []apiTorrentStasResTorrents `json:"torrents"`
	}

	apiTorrentStasResTorrents struct {
		Name          string                         `json:"name"`
		InfoHash      string                         `json:"infohash"`
		TotalPeers    int                            `json:"totalpeers"`
		ActivePeers   int                            `json:"activepeers"`
		PendingPeers  int                            `json:"pendingpeers"`
		HalfOpenPeers int                            `json:"halfopenpeers"`
		DownloadSpeed string                         `json:"downloadspeed"`
		UploadSpeed   string                         `json:"uploadspeed"`
		Progress      string                         `json:"progress"`
		Files         []apiTorrentStatsTorrentsFiles `json:"files"`
	}

	apiTorrentStatsTorrentsFiles struct {
		FileName        string `json:"filename"`
		FileSizeBytes   int    `json:"filesizebytes"`
		BytesDownloaded int    `json:"bytesdownloaded"`
		Stream          string `json:"stream,omitempty"`
		Download        string `json:"download,omitempty"`
	}
)
