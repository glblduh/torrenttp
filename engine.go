/* Contains global variables and structs of the program */

package main

import (
	"github.com/anacrolix/torrent"
)

// Creates the BitTorrent client
func initBTClient(opts *torrent.ClientConfig) {
	btEngine.BTClientConfig = opts
	var err error
	btEngine.BTClient, err = torrent.NewClient(btEngine.BTClientConfig)
	if err != nil {
		Error.Fatalf("Cannot initialize BitTorrent client: %s", err)
	}
}

// Create config for BitTorrent client with confs from args
func newBtCliConfs(dir string, noup bool) *torrent.ClientConfig {
	opts := torrent.NewDefaultClientConfig()
	opts.DataDir = dir
	opts.NoUpload = noup
	return opts
}
