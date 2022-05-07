/* Contains functions for manipulating the BitTorrent client */

package main

import (
	"errors"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
)

// Creates the BitTorrent client
func (Engine *btEng) initialize(opts *torrent.ClientConfig) {
	Engine.ClientConfig = opts
	var err error
	Engine.Client, err = torrent.NewClient(Engine.ClientConfig)
	if err != nil {
		Error.Fatalf("Cannot initialize BitTorrent client: %s", err)
	}
	Engine.Torrents = make(map[string]torrentHandle)
}

// Add torrent to client
func (Engine *btEng) addTorrent(spec *torrent.TorrentSpec, noSave bool) (*torrent.Torrent, error) {
	t, new, err := Engine.Client.AddTorrentSpec(spec)
	if err != nil {
		Warn.Printf("Cannot add torrent spec: %s\n", err)
		return nil, err
	}
	if new && !noSave {
		sserr := specDB.saveSpec(spec)
		if sserr != nil {
			Warn.Printf("Cannot save torrent spec: %s\n", sserr)
		}
	}

	Engine.addTorrentHandle(t, spec)

	<-t.GotInfo()
	return t, nil
}

// Get *torrent.Torrent from infohash
func (Engine *btEng) getTorrHandle(infohash string) (*torrent.Torrent, error) {
	if len(infohash) != 40 {
		Warn.Println("Invalid infohash")
		return nil, errors.New("Invalid infohash")
	}
	t, ok := Engine.Client.Torrent(metainfo.NewHashFromHex(infohash))
	if !ok {
		Warn.Println("Torrent not found")
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
	rmerr := specDB.removeSpec(t.InfoHash().String())
	if rmerr != nil {
		Warn.Printf("Cannot remove spec from DB: %s\n", rmerr)
	}
	return rmerr
}

// Adds torrent handle to custom torrent handler
func (Engine *btEng) addTorrentHandle(t *torrent.Torrent, spec *torrent.TorrentSpec) {
	Engine.Torrents[t.InfoHash().String()] = torrentHandle{
		Torrent:        t,
		Spec:           spec,
		Name:           t.Name(),
		InfoHash:       t.InfoHash(),
		InfoHashString: t.InfoHash().String(),
	}
}

// Remove torrent handle from custom torrent handle
func (Engine *btEng) removeTorrentHandle(infohash string) {
	delete(Engine.Torrents, infohash)
}

func (Engine *btEng) calculateSpeeds(infohash string) {
	handle := Engine.Torrents[infohash]

	/* Download speed */
	curprog := handle.Torrent.BytesCompleted()
	handle.DlSpeedBytes = curprog - handle.DlLastProgress
	handle.DlSpeedReadable = humanize.Bytes(uint64(handle.DlSpeedBytes)) + "/s"
	handle.DlLastProgress = curprog
}
