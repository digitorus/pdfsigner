package db

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

var opts badger.Options

func init() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts = badger.DefaultOptions
	opts.Dir = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/badger"
	opts.ValueDir = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/badger"
}

type DB struct {
	db *badger.DB
}

func NewDBConnection() (*badger.DB, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return db, err
	}

	return db, nil
}

func (d *DB) Close() {

}

func SaveByKey(key string, value []byte) error {
	db, err := badger.Open(opts)
	if err != nil {
		return errors.Wrap(err, "open connection")
	}
	defer db.Close()

	err = db.Update(func(txn *badger.Txn) error {
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
	db, err := badger.Open(opts)
	if err != nil {
		return result, errors.Wrap(err, "open connection")
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
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
