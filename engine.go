/* Contains functions for manipulating the BitTorrent client */

package main

import (
	"errors"
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

	/* Initialize custom torrent map and speed calculator */
	Engine.Torrents = make(map[string]*torrentHandle)
	go btEngine.calculateSpeeds()
}

// Add torrent to client
func (Engine *btEng) addTorrent(spec *torrent.TorrentSpec, noSave bool) (*torrent.Torrent, error) {
	t, new, err := Engine.Client.AddTorrentSpec(spec)
	if err != nil {
		return nil, err
	}
	if new && !noSave {
		sserr := saveSpec(spec)
		if sserr != nil {
			return nil, sserr
		}
	}

	Engine.addTorrentHandle(t, spec)

	<-t.GotInfo()
	return t, nil
}

// Get *torrent.Torrent from infohash
func (Engine *btEng) getTorrHandle(infohash string) (*torrent.Torrent, error) {
	if len(infohash) != 40 {
		return nil, errors.New("Invalid infohash")
	}
	t, ok := Engine.Client.Torrent(metainfo.NewHashFromHex(infohash))
	if !ok {
		return nil, errors.New("Torrent not found")
	}
	return t, nil
}

// Removes torrent from BitTorrent client and removes it's persistence spec
func (Engine *btEng) dropTorrent(infohash string) error {
	t, err := Engine.getTorrHandle(infohash)
	if err != nil {
		return err
	}
	t.Drop()
	Engine.removeTorrentHandle(infohash)
	rmerr := removeSpec(t.InfoHash().String())
	return rmerr
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
			/* Download speed */
			curprog := torrents[k].Torrent.BytesCompleted()
			torrents[k].DlSpeedBytes = (int64(time.Second) * (curprog - torrents[k].LastDlBytes)) / (int64(1 * time.Second))
			torrents[k].LastDlBytes = curprog
			torrents[k].DlSpeedReadable = humanize.Bytes(uint64(torrents[k].DlSpeedBytes)) + "/s"
		}
		time.Sleep(1 * time.Second)
	}
}
