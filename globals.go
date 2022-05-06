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
	/* BitTorrent client struct */
	btEng struct {
		BTClient       *torrent.Client
		BTClientConfig *torrent.ClientConfig
	}
)

/* Functions */
