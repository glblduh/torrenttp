/* Contains functions for manipulating the BoltDB file */

package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/boltdb/bolt"
)

func openDB() (*bolt.DB, error) {
	return bolt.Open(
		filepath.Join(btEngine.ClientConfig.DataDir, ".torrserve.db"),
		0600,
		nil)
}

func createSpecBucket() error {
	db, dberr := openDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("TorrSpecs"))
		return nil
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
	db, dberr := openDB()
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
		t, terr := btEngine.addTorrent(persistSpecToTorrentSpec(spec), true)
		if terr != nil {
			Warn.Printf("Cannot load spec \"%s\": %s\n", spec.InfoHash, terr)
			rmerr := removeSpec(spec.InfoHash)
			if rmerr != nil {
				Warn.Printf("Cannot remove spec \"%s\": %s\n", spec.InfoHash, rmerr)
			}
			continue
		}
		Info.Printf("Loaded torrent \"%s\"\n", t.Name())
		for _, f := range spec.Files {
			tf, tferr := getTorrentFile(t, f)
			if tferr != nil {
				Warn.Printf("Cannot load file \"%s\": %s\n", f, tferr)
				continue
			}
			tf.Download()
			Info.Printf("Started download of file \"%s\"", tf.DisplayPath())
		}
	}
	return nil
}

// Returns all persistentSpec in DB
func getSpecs() ([]persistentSpec, error) {
	db, dberr := openDB()
	if dberr != nil {
		return []persistentSpec{}, dberr
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

func removeSpec(infohash string) error {
	db, dberr := openDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		return b.Delete([]byte(strings.ToLower(infohash)))
	})
}

// Adds file of torrent to DB for persistence
func saveSpecFile(infohash string, filename string) error {
	spec, err := getSpec(infohash)
	if err != nil {
		return err
	}
	rmerr := removeSpec(infohash)
	if rmerr != nil {
		return rmerr
	}
	spec.Files = append(spec.Files, filename)
	json, jerr := json.Marshal(&spec)
	if jerr != nil {
		return jerr
	}
	return specToDB(infohash, json)
}
