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

	/* Initilize DB and load persistent specs */
	dberr := createSpecBucket()
	if dberr != nil {
		Error.Fatalf("Cannot initialize DB: %s\n", dberr)
	}
	go loadPersist()

	/* Initialize endpoints and HTTP server */
	r := mux.NewRouter().StrictSlash(true)

	/* Handlers for endpoints */

	/* POST */
	r.HandleFunc("/api/addtorrent", apiAddTorrent).Methods("POST")
	r.HandleFunc("/api/selectfile", apiTorrentSelectFile).Methods("POST")

	/* DELETE */
	r.HandleFunc("/api/removetorrent", apiRemoveTorrent).Methods("DELETE")

	/* GET */
	r.HandleFunc("/api/stream/{infohash}/{file:.*}", apiStreamTorrentFile).Methods("GET")
	r.HandleFunc("/api/file/{infohash}/{file:.*}", apiDownloadFile).Methods("GET")
	r.HandleFunc("/api/torrents", apiTorrentStats).Methods("GET")
	r.HandleFunc("/api/torrents/{infohash}", apiTorrentStats).Methods("GET")

	/* CORS middleware */
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	}).Handler(r)

	Info.Printf("Starting HTTP server on port: http://%s", *portFlag)
	Error.Fatalln(http.ListenAndServe(*portFlag, c))
}
