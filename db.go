/* Contains functions for manipulating the BoltDB file */

package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/boltdb/bolt"
)

func getDB() (*bolt.DB, error) {
	db, err := bolt.Open(
		filepath.Join(btEngine.ClientConfig.DataDir, ".torrserve.db"),
		0600,
		nil)
	return db, err
}

func createSpecBucket() error {
	db, dberr := getDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		_, berr := tx.CreateBucketIfNotExists([]byte("TorrSpecs"))
		return berr
	})
}

// Saves torrent spec to database file
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
		return err
	}
	return specToDB(spec.InfoHash.String(), json)
}

// Commit a persistentSpec to DB
func specToDB(infohash string, json []byte) error {
	db, dberr := getDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		return b.Put([]byte(strings.ToLower(infohash)), json)
	})
}

// Loads all persistentSpec to BitTorrent client
func loadPersist() error {
	specs, err := getSpecs()
	if err != nil {
		return err
	}
	for _, spec := range specs {
		t, terr := addTorrent(persistSpecToTorrentSpec(spec), true)
		if terr != nil {
			Warn.Printf("Cannot load spec \"%s\": %s\n", spec.InfoHash, terr)
			rmerr := removeSpec(spec.InfoHash)
			if rmerr != nil {
				Warn.Printf("Cannot remove spec \"%s\": %s\n", spec.InfoHash, rmerr)
			}
			continue
		}
		if spec.AllFiles {
			t.DownloadAll()
		}
		for _, f := range spec.Files {
			tf, tferr := getTorrentFile(t, f)
			if tferr != nil {
				Warn.Printf("Cannot load file %s: %s\n", f, tferr)
				continue
			}
			tf.Download()
		}
		Info.Printf("Loaded torrent \"%s\"\n", t.Name())
	}
	return nil
}

// Returns all persistentSpec in DB
func getSpecs() ([]persistentSpec, error) {
	db, dberr := getDB()
	if dberr != nil {
		return nil, dberr
	}
	defer db.Close()
	specs := []persistentSpec{}
	verr := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			spec := persistentSpec{}
			derr := json.Unmarshal(v, &spec)
			if derr != nil {
				return derr
			}
			specs = append(specs, spec)
		}
		return nil
	})
	return specs, verr
}

// Get specific persistentSpec from infohash
func getSpec(infohash string) (persistentSpec, error) {
	specs, err := getSpecs()
	if err != nil {
		return persistentSpec{}, err
	}
	for _, spec := range specs {
		if spec.InfoHash == infohash {
			return spec, nil
		}
	}
	return persistentSpec{}, errors.New("Torrent spec not found")
}

// Turns persistentSpec to *torrent.TorrentSpec
func persistSpecToTorrentSpec(spec persistentSpec) *torrent.TorrentSpec {
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
	}
}

func removeSpec(infohash string) error {
	db, dberr := getDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		return b.Delete([]byte(strings.ToLower(infohash)))
	})
}

// Adds selected files of torrent to DB for persistence
func saveSpecFiles(infohash string, allfiles bool, files []string) error {
	spec, err := getSpec(infohash)
	if err != nil {
		return err
	}
	rmerr := removeSpec(infohash)
	if rmerr != nil {
		return rmerr
	}
	spec.AllFiles = allfiles
	spec.Files = files
	json, jerr := json.Marshal(&spec)
	if jerr != nil {
		return jerr
	}
	return specToDB(infohash, json)
}
