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
	authFlag := flag.Bool("auth", false, "Enable API key authentication from the env varible TORRENTTPKEY")
	flag.Parse()

	// Check if authentication is enabled
	checkAuthEnabled(*authFlag)

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
	r.HandleFunc("/api/addtorrent", checkAuth(apiAddTorrent)).Methods("POST")
	r.HandleFunc("/api/selectfile", checkAuth(apiTorrentSelectFile)).Methods("POST")
	r.HandleFunc("/api/setpriority", checkAuth(apiTorrentPriorityFile)).Methods("POST")
	r.HandleFunc("/api/addtorrentfile", checkAuth(apiAddTorrentFile)).Methods("POST")

	/* DELETE */
	r.HandleFunc("/api/removetorrent", checkAuth(apiRemoveTorrent)).Methods("DELETE")

	/* GET */
	r.HandleFunc("/api/stream/{infohash}/{file:.*}", checkAuth(apiStreamTorrentFile)).Methods("GET")
	r.HandleFunc("/api/file/{infohash}/{file:.*}", checkAuth(apiDownloadFile)).Methods("GET")
	r.HandleFunc("/api/torrents", checkAuth(apiTorrentStats)).Methods("GET")
	r.HandleFunc("/api/torrents/{infohash}", checkAuth(apiTorrentStats)).Methods("GET")

	/* CORS middleware */
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	}).Handler(r)

	Info.Printf("Starting HTTP server on port: %s", *portFlag)
	Error.Fatalln(http.ListenAndServe(*portFlag, c))
}
