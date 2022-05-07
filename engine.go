/* Contains functions for manipulating the BitTorrent client */

package main

import (
	"errors"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// Creates the BitTorrent client
func initBTClient(opts *torrent.ClientConfig) {
	btEngine.ClientConfig = opts
	var err error
	btEngine.Client, err = torrent.NewClient(btEngine.ClientConfig)
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

// Add torrent to client
func addTorrent(spec *torrent.TorrentSpec, noSave bool) (*torrent.Torrent, error) {
	t, new, err := btEngine.Client.AddTorrentSpec(spec)
	if err != nil {
		Warn.Printf("Cannot add torrent spec: %s\n", err)
		return nil, err
	}
	if new && !noSave {
		sserr := saveSpec(spec)
		if sserr != nil {
			Warn.Printf("Cannot save torrent spec: %s\n", sserr)
		}
	}
	<-t.GotInfo()
	return t, nil
}

// Get *torrent.Torrent from infohash
func getTorrHandle(infohash string) (*torrent.Torrent, error) {
	if len(infohash) != 40 {
		Warn.Println("Invalid infohash")
		return nil, errors.New("Invalid infohash")
	}
	t, ok := btEngine.Client.Torrent(metainfo.NewHashFromHex(infohash))
	if !ok {
		Warn.Println("Torrent not found")
		return nil, errors.New("Torrent not found")
	}
	return t, nil
}

// Removes torrent from BitTorrent client and removes it's persistence spec
func dropTorrent(infohash string) error {
	t, err := getTorrHandle(infohash)
	if err != nil {
		return err
	}
	t.Drop()
	rmerr := removeSpec(t.InfoHash().String())
	if rmerr != nil {
		Warn.Printf("Cannot remove spec from DB: %s\n", rmerr)
	}
	return rmerr
}

// Get the file handle inside the torrent
func getTorrentFile(t *torrent.Torrent, filename string) (*torrent.File, error) {
	for _, f := range t.Files() {
		if f.DisplayPath() == filename {
			return f, nil
		}
	}
	return nil, errors.New("File not found")
}
