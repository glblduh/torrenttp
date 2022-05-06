/* Contains functions for manipulating the BitTorrent client */

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
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

// Add torrent to client
func addTorrent(spec *torrent.TorrentSpec, noSave bool) (*torrent.Torrent, error) {
	t, _, err := btEngine.BTClient.AddTorrentSpec(spec)
	if err != nil {
		Warn.Printf("Cannot add torrent spec: %s\n", err)
		return nil, err
	}
	if !noSave {
		saveSpec(spec)
	}
	return t, nil
}

// Saves torrent spec for persistence
func saveSpec(spec *torrent.TorrentSpec) error {
	json, err := json.Marshal(persistentSpec{
		Trackers:                 spec.Trackers,
		InfoHash:                 spec.InfoHash.String(),
		DisplayName:              spec.DisplayName,
		Webseeds:                 spec.Webseeds,
		DhtNodes:                 spec.DhtNodes,
		PeerAddrs:                spec.PeerAddrs,
		Sources:                  spec.Sources,
		DisableInitialPieceCheck: spec.DisableInitialPieceCheck,
		DisallowDataUpload:       spec.DisallowDataUpload,
		DisallowDataDownload:     spec.DisallowDataDownload,
	})
	if err != nil {
		Warn.Printf("Cannot marshal torrent spec: %s\n", err)
		return err
	}
	wferr := ioutil.WriteFile(
		filepath.Join(btEngine.BTClientConfig.DataDir, "."+spec.InfoHash.String()+".torrspec"),
		json,
		0644)
	if wferr != nil {
		Warn.Printf("Cannot write torrent spec to disk: %s\n", wferr)
		return wferr
	}
	return nil
}

// Loads the torrent spec from files
func parseSpecFiles() {
	files, err := ioutil.ReadDir(btEngine.BTClientConfig.DataDir)
	if err != nil {
		Warn.Printf("Cannot read directory: %s\n", err)
	}
	for _, file := range files {
		splitted := strings.Split(file.Name(), ".")
		if splitted[len(splitted)-1] != "torrspec" {
			continue
		}
		spec, dsferr := decodeSpecFile(file.Name())
		if dsferr == nil {
			addTorrent(spec, true)
		}
	}
}

// Decodes JSON spec to *torrent.TorrentSpec
func decodeSpecFile(fn string) (*torrent.TorrentSpec, error) {
	sjson, err := ioutil.ReadFile(filepath.Join(btEngine.BTClientConfig.DataDir, fn))
	if err != nil {
		Warn.Printf("Cannot read file: %s\n", err)
		return nil, err
	}
	spec := persistentSpec{}
	unerr := json.Unmarshal(sjson, &spec)
	if unerr != nil {
		Warn.Printf("Cannot unmarshal file: %s\n", unerr)
		return nil, unerr
	}

	return &torrent.TorrentSpec{
		Trackers:                 spec.Trackers,
		InfoHash:                 metainfo.NewHashFromHex(spec.InfoHash),
		DisplayName:              spec.DisplayName,
		Webseeds:                 spec.Webseeds,
		DhtNodes:                 spec.DhtNodes,
		PeerAddrs:                spec.PeerAddrs,
		Sources:                  spec.Sources,
		DisableInitialPieceCheck: spec.DisableInitialPieceCheck,
		DisallowDataUpload:       spec.DisallowDataUpload,
		DisallowDataDownload:     spec.DisallowDataDownload,
	}, nil
}

// Get *torrent.Torrent from infohash
func getTorrHandle(infohash string) (*torrent.Torrent, error) {
	if len(infohash) != 40 {
		Warn.Println("Invalid infohash")
		return nil, errors.New("Invalid infohash")
	}
	t, ok := btEngine.BTClient.Torrent(metainfo.NewHashFromHex(infohash))
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
	rmerr := os.Remove(filepath.Join(
		btEngine.BTClientConfig.DataDir,
		"."+t.InfoHash().String()+".torrspec"))
	if rmerr != nil {
		Warn.Printf("Cannot remove torrspec file: %s\n", rmerr)
		return rmerr
	}
	return nil
}
