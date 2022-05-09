/* Contains functions for manipulating the BitTorrent client */

package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
)

// Creates the BitTorrent client
func (Engine *btEng) initialize(opts *torrent.ClientConfig) {
	/* Make client with confs */
	Engine.ClientConfig = opts
	var err error
	Engine.Client, err = torrent.NewClient(Engine.ClientConfig)
	if err != nil {
		Error.Fatalf("Cannot initialize BitTorrent client: %s", err)
	}

	/* Outputs the download directory and upload status */
	Info.Printf("Download directory is on: %s\n", Engine.ClientConfig.DataDir)
	if Engine.ClientConfig.NoUpload {
		Warn.Println("Upload is disabled")
	}

	/* Initialize custom torrent map, persistent loader, and speed calculator */
	Engine.Torrents = make(map[string]*torrentHandle)
	go loadPersist()
	go btEngine.calculateSpeeds()
}

// Add torrent to client
func (Engine *btEng) addTorrent(spec *torrent.TorrentSpec, noSave bool) (*torrent.Torrent, error) {
	/* Adds spec to BitTorrent client */
	t, new, err := Engine.Client.AddTorrentSpec(spec)
	if err != nil {
		return nil, err
	}

	/* Check if torrent is new then save its spec for persistence */
	if new && !noSave {
		sserr := saveSpec(spec)
		if sserr != nil {
			return nil, sserr
		}
	}

	// Wait for torrent info
	<-t.GotInfo()

	// Adds spec to custom torrent handler
	Engine.addTorrentHandle(t, spec)

	return t, nil
}

// Get *torrent.Torrent from infohash
func (Engine *btEng) getTorrHandle(infohash string) (*torrent.Torrent, error) {
	/* Checks if infohash is 40 characters */
	if len(infohash) != 40 {
		return nil, errors.New("Invalid infohash")
	}

	/* Get torrent handle */
	t, ok := Engine.Client.Torrent(metainfo.NewHashFromHex(infohash))
	if !ok {
		return nil, errors.New("Torrent not found")
	}
	return t, nil
}

// Removes torrent from BitTorrent client and removes it's persistence spec
func (Engine *btEng) dropTorrent(infohash string, rmfiles bool) error {
	/* Get torrent handle */
	t, err := Engine.getTorrHandle(infohash)
	if err != nil {
		return err
	}

	/* Remove torrent handles */
	Engine.removeTorrentHandle(infohash)
	t.Drop()

	/* Removes torrent persistence spec */
	rmerr := removeSpec(t.InfoHash().String())
	if rmerr != nil {
		return rmerr
	}

	/* Removes torrent files */
	if rmfiles {
		return os.RemoveAll(filepath.Join(Engine.ClientConfig.DataDir, t.Name()))
	}
	return nil
}

// Adds torrent handle to custom torrent handler
func (Engine *btEng) addTorrentHandle(t *torrent.Torrent, spec *torrent.TorrentSpec) {
	Engine.Torrents[t.InfoHash().String()] = &torrentHandle{
		Torrent: t,
		Spec:    spec,
	}
}

// Remove torrent handle from custom torrent handle
func (Engine *btEng) removeTorrentHandle(infohash string) {
	delete(Engine.Torrents, infohash)
}

func (Engine *btEng) calculateSpeeds() {
	torrents := Engine.Torrents

	for {
		for k := range torrents {
			/*
				Work-around for the oddity cause by atomics
				See: https://github.com/anacrolix/torrent/issues/745
			*/
			curstats := torrents[k].Torrent.Stats()

			/* Download speed */
			dlcurprog := curstats.BytesReadUsefulData.Int64()
			torrents[k].DlSpeedBytes = (int64(time.Second) * (dlcurprog - torrents[k].LastDlBytes)) / int64(1*time.Second)
			torrents[k].LastDlBytes = dlcurprog
			torrents[k].DlSpeedReadable = humanize.Bytes(uint64(torrents[k].DlSpeedBytes)) + "/s"

			/* Upload speed */
			ulcurprog := curstats.BytesWrittenData.Int64()
			torrents[k].UlSpeedBytes = (int64(time.Second) * (ulcurprog - torrents[k].LastUlBytes)) / int64(1*time.Second)
			torrents[k].LastUlBytes = ulcurprog
			torrents[k].UlSpeedReadable = humanize.Bytes(uint64(torrents[k].UlSpeedBytes)) + "/s"
		}
		time.Sleep(1 * time.Second)
	}
}
