package main

import (
	"flag"
	"net/http"
)

func main() {
	/* Argument flags */
	dirFlag := flag.String("dir", "torrservedl", "Download directory path")
	portFlag := flag.String("port", ":1010", "HTTP server listening port")
	noupFlag := flag.Bool("noup", false, "Disables BT client upload")
	flag.Parse()

	// Creates the BitTorrent client with user args
	initBTClient(newBtCliConfs(*dirFlag, *noupFlag))

	/* Outputs the download directory and upload status */
	Info.Printf("Download directory is on: %s", btEngine.BTClientConfig.DataDir)
	if btEngine.BTClientConfig.NoUpload {
		Warn.Println("Upload is disabled")
	}

	// Parses torrspec files
	parseSpecFiles()

	/* Initialize HTTP server */
	http.ListenAndServe(*portFlag, nil)
}
