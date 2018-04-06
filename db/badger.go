package db

import (
	"github.com/dgraph-io/badger"
)

var opts badger.Options

func init() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts = badger.DefaultOptions
	opts.Dir = "./badger"
	opts.ValueDir = "./badger"
}

func SaveByKey(key string, value []byte) error {
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})

	return err
}

func LoadByKey(key string) ([]byte, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return []byte{}, err
	}
	defer db.Close()

	var result []byte
	err = db.View(func(txn *badger.Txn) error {
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
		return []byte{}, err
	}

	return result, nil
}
