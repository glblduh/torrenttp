package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	dberr := createSpecBucket()
	if dberr != nil {
		Error.Fatalf("Cannot initialize DB: %s\n", dberr)
	}
	// Parses torrent specs in DB
	lperr := loadPersist()
	if lperr != nil {
		Warn.Printf("Cannot load torrent specs: %s\n", lperr)
	}

	/* Initialize endpoints and HTTP server */
	r := mux.NewRouter().StrictSlash(true)

	/* Handlers for endpoints */
	r.HandleFunc("/api/addtorrent", apiAddTorrent).Methods("POST")
	r.HandleFunc("/api/selectfile", apiTorrentSelectFile).Methods("POST")
	r.HandleFunc("/api/stream/{infohash}/{file:.*}", apiStreamTorrentFile).Methods("GET")
	r.HandleFunc("/api/removetorrent", apiRemoveTorrent).Methods("DELETE")
	r.HandleFunc("/api/torrents", apiTorrentStats).Methods("GET")
	r.HandleFunc("/api/torrents/{infohash}", apiTorrentStats).Methods("GET")

	/* CORS middleware */
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	}).Handler(r)

	Error.Fatalln(http.ListenAndServe(*portFlag, c))
}
