package db

import (
	"log"
	"os"
	"path"
	"strings"

	"bitbucket.org/digitorus/pdfsigner/utils"
	"github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

var (
	DB *bolt.DB
)

func init() {
	// set path for bolt folder
	var err error
	var currentFolder string

	// get path to executable file
	runFileFolder, err := utils.GetRunFileFolder()
	if err != nil {
		log.Fatal(err)
	}

	// determine path to use for the db
	if utils.IsTestEnvironment() || strings.Contains(runFileFolder+"/", os.TempDir()) {
		currentFolder = path.Join(utils.GetGoPath(), "/src/bitbucket.org/digitorus/pdfsigner")
	} else {
		currentFolder = runFileFolder
	}

	DB, err = bolt.Open(path.Join(currentFolder, "pdfsigner.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getBucketName(key string) string {
	if strings.Contains(key, "_") {
		return strings.Split(key, "_")[0]
	}

	return key
}

// SaveByKey saves value into bolt by key
func SaveByKey(key string, value []byte) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		//spew.Dump(key, string(value))
		bucketName := getBucketName(key)
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), value)
		return err
	})

	if err != nil {
		return errors.Wrap(err, "update by key")
	}

	return nil
}

// LoadByKey loads from bolt by key
func LoadByKey(key string) ([]byte, error) {
	var result []byte
	err := DB.View(func(tx *bolt.Tx) error {
		bucketName := getBucketName(key)
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}

		result = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return result, errors.Wrap(err, "view by key")
	}

	return result, nil
}

// DeleteByKey deletes value by key
func DeleteByKey(key string) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		bucketName := getBucketName(key)
		return tx.Bucket([]byte(bucketName)).Delete([]byte(key))
	})

	return err
}

// BatchLoad loads multiple values and returns map
func BatchLoad(prefix string) (map[string][]byte, error) {
	var result = make(map[string][]byte)
	prefix = getBucketName(prefix)

	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(prefix))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			result[string(k[:])] = v
		}
		return nil
	})

	if err != nil {
		return result, err
	}

	return result, nil
}
