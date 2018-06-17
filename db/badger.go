package db

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bitbucket.org/digitorus/pdfsigner/utils"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

// badgerFolder represents the last relative part of the path to the badger
const badgerFolder = "badger"

var (
	opts badger.Options
	DB   *badger.DB
)

func init() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts = badger.DefaultOptions

	// set path for badger folder
	var err error
	var currentFolder string

	// get path to executable file
	runFileFolder, err := utils.GetRunFileFolder()
	if err != nil {
		log.Fatal(err)
	}

	// determine path to use for the db
	if utils.IsTestEnvironment() || strings.Contains(runFileFolder, os.TempDir()) {
		currentFolder = path.Join(utils.GetGoPath(), "/src/bitbucket.org/digitorus/pdfsigner")
	} else {
		currentFolder = runFileFolder
	}

	opts.Dir = filepath.Join(currentFolder, badgerFolder)
	opts.ValueDir = filepath.Join(currentFolder, badgerFolder)

	// initialize badger
	DB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
}

// SaveByKey saves value into badger by key
func SaveByKey(key string, value []byte) error {
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})
	if err != nil {
		return errors.Wrap(err, "update by key")
	}

	return nil
}

// LoadByKey loads from badger by key
func LoadByKey(key string) ([]byte, error) {
	var result []byte
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		result, err = item.Value()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, errors.Wrap(err, "view by key")
	}

	return result, nil
}

// DeleteByKey deletes value by key
func DeleteByKey(key string) error {
	err := DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	return err
}

// BatchUpsert inserts or updates multiple values in single transaction
func BatchUpsert(objectsByID map[string][]byte) error {
	err := DB.Update(func(txn *badger.Txn) error {
		for id, object := range objectsByID {
			err := txn.Set([]byte(id), object)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// BatchDelete deletes multiple values with single transaction
func BatchDelete(objectIDs []string) error {
	err := DB.Update(func(txn *badger.Txn) error {
		for _, id := range objectIDs {
			err := txn.Delete([]byte(id))
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// BatchLoad loads multiple values and returns map
func BatchLoad(prefix string) (map[string][]byte, error) {
	var result = make(map[string][]byte)

	err := DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			if err != nil {
				return err
			}

			var key = string(k[:])
			if strings.Contains(key, prefix) {
				result[string(k[:])] = v
			}
		}
		return nil
	})

	if err != nil {
		return result, err
	}

	return result, nil
}
