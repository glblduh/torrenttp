/* Contains functions for manipulating the BoltDB file */

package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"go.etcd.io/bbolt"
)

func openDB() (*bbolt.DB, error) {
	return bbolt.Open(
		filepath.Join(btEngine.ClientConfig.DataDir, ".torrserve.db"),
		0660,
		&bbolt.Options{
			Timeout: time.Second,
		})
}

func createSpecBucket() error {
	/* Opens DB file */
	db, dberr := openDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()

	/* Create TorrSpec bucket */
	return db.Update(func(tx *bbolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("TorrSpecs"))
		return nil
	})
}

// Saves torrent spec to database file
func saveSpec(spec *torrent.TorrentSpec) error {
	/* Marshal torrent spec to JSON persistentSpec */
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
	/* Opens DB file */
	db, dberr := openDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()

	/* Adds marshal'd spec to DB file */
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		return b.Put([]byte(strings.ToLower(infohash)), json)
	})
}

// Loads all persistentSpec to BitTorrent client
func loadPersist() {
	/* Get all specs from DB */
	specs, err := getSpecs()
	if err != nil {
		Warn.Printf("Cannot get persistent specs: %s\n", err)
		return
	}

	/* Iterates over all specs */
	for _, spec := range specs {
		/* Add spec to BitTorrent client */
		t, terr := btEngine.addTorrent(persistSpecToTorrentSpec(spec), true)
		if terr != nil {
			Warn.Printf("Cannot load spec \"%s\": %s\n", spec.InfoHash, terr)
			rmerr := removeSpec(spec.InfoHash)
			if rmerr != nil {
				Warn.Printf("Cannot remove spec \"%s\": %s\n", spec.InfoHash, rmerr)
			}
			continue
		}

		/* Start download of files in persistent spec */
		for _, f := range spec.Files {
			tf, tferr := getTorrentFile(t, f)
			if tferr != nil {
				Warn.Printf("Cannot load file \"%s\": %s\n", f, tferr)
				continue
			}
			tf.Download()
		}
	}
}

// Returns all persistentSpec in DB
func getSpecs() ([]persistentSpec, error) {
	/* Opens DB file */
	db, dberr := openDB()
	if dberr != nil {
		return []persistentSpec{}, dberr
	}
	defer db.Close()

	/* Iterates over all specs in DB to make array of specs */
	specs := []persistentSpec{}
	verr := db.View(func(tx *bbolt.Tx) error {
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
	/* Get all specs from DB */
	specs, err := getSpecs()
	if err != nil {
		return persistentSpec{}, err
	}

	/* Returns specified spec */
	for _, spec := range specs {
		if spec.InfoHash == infohash {
			return spec, nil
		}
	}
	return persistentSpec{}, errors.New("torrent spec not found")
}

func removeSpec(infohash string) error {
	/* Opens DB file */
	db, dberr := openDB()
	if dberr != nil {
		return dberr
	}
	defer db.Close()

	/* Deletes spec */
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("TorrSpecs"))
		return b.Delete([]byte(strings.ToLower(infohash)))
	})
}

// Adds file of torrent to DB for persistence
func saveSpecFile(infohash string, filename string) error {
	/* Get persistence spec from infohash */
	spec, err := getSpec(infohash)
	if err != nil {
		return err
	}

	/* Check for duplicates */
	for _, f := range spec.Files {
		if f == filename {
			return nil
		}
	}

	/* Remove unmodified spec */
	rmerr := removeSpec(infohash)
	if rmerr != nil {
		return rmerr
	}

	/* Create new spec with file */
	spec.Files = append(spec.Files, filename)
	json, jerr := json.Marshal(&spec)
	if jerr != nil {
		return jerr
	}
	return specToDB(infohash, json)
}
