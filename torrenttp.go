package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	/* Argument flags */
	dirFlag := flag.String("dir", "torrenttpdl", "Download directory path")
	portFlag := flag.String("port", ":1010", "HTTP server listening port")
	noupFlag := flag.Bool("noup", false, "Disables BT client upload")
	flag.Parse()

	// Creates the BitTorrent client with user args
	btEngine.initialize(newBtCliConfs(*dirFlag, *noupFlag))

	/* Outputs the download directory and upload status */
	Info.Printf("Download directory is on: %s\n", btEngine.ClientConfig.DataDir)
	if btEngine.ClientConfig.NoUpload {
		Warn.Println("Upload is disabled")
	}

	/* Initilize DB */
	dberr := specDB.initDB()
	if dberr != nil {
		Error.Fatalf("Cannot initialize DB: %s\n", dberr)
	}
	specDB.createSpecBucket()
	// Parses torrent specs in DB
	lperr := specDB.loadPersist()
	if lperr != nil {
		Warn.Printf("Cannot load torrent specs: %s\n", lperr)
	}

	/* Initialize endpoints and HTTP server */
	r := mux.NewRouter().StrictSlash(true)

	/* Handlers for endpoints */
	r.HandleFunc("/api/addtorrent", apiAddTorrent).Methods("POST")
	r.HandleFunc("/api/selectfile", apiTorrentSelectFile).Methods("POST")
	r.HandleFunc("/api/stream/{infohash}/{file:.*}", apiStreamTorrentFile).Methods("GET")

	Error.Fatalln(http.ListenAndServe(*portFlag, r))
}
