package db

import (
	"log"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

var opts badger.Options

var DB *badger.DB

func init() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts = badger.DefaultOptions
	opts.Dir = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/badger"
	opts.ValueDir = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/badger"

	var err error
	DB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
}

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

func LoadByKey(key string) ([]byte, error) {
	var result []byte
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		result = append(result, val...)
		return nil
	})
	if err != nil {
		return result, errors.Wrap(err, "view by key")
	}

	return result, nil
}

func DeleteByKey(key string) error {
	err := DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	return err
}

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
