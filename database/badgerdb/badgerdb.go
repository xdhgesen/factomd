// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package badgerdb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/dgraph-io/badger"
)

const BadgerDirectoryName = "factoid_badger.db"

type BadgerDB struct {
	db     *badger.DB // Pointer to the badger db
	closed bool
}

var _ interfaces.IDatabase = (*BadgerDB)(nil)

func NewBadgerDB(filename string) (*BadgerDB, error) {
	db := new(BadgerDB)
	err := db.Init(filename)
	return db, err
}

func NewAndCreateBadgerDB(filedir string) (*BadgerDB, error) {
	err := os.MkdirAll(filepath.Dir(filedir), 0750)
	if err != nil {
		if err != nil {
			panic("Database could not be created, " + err.Error())
		}
	}
	return NewBadgerDB(filedir)
}

/***************************************
 *       Methods
 ***************************************/

func (db *BadgerDB) ListAllBuckets() ([][]byte, error) {
	return nil, fmt.Errorf("Unable to fetch buckets due to BadgerDB design")
}

// We don't care if delete works or not.  If the key isn't there, that's ok
func (db *BadgerDB) Delete(bucket []byte, key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(leveldb.CombineBucketAndKey(bucket, key))
	})
}

// Can't trim a real database
func (db *BadgerDB) Trim() {
}

func (db *BadgerDB) Close() error {
	if db.closed {
		return nil
	}
	db.closed = true
	return db.db.Close()
}

func (db *BadgerDB) Get(bucket []byte, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(leveldb.CombineBucketAndKey(bucket, key))
		if err == badger.ErrKeyNotFound {
			destination = nil
			return nil
		}
		if err != nil {
			return err
		}

		data, err := item.Value()
		if err != nil {
			return err
		}
		return destination.UnmarshalBinary(data)
	})
	return destination, err
}

func (db *BadgerDB) Put(bucket []byte, key []byte, data interfaces.BinaryMarshallable) error {
	hex, err := data.MarshalBinary()
	if err != nil {
		return err
	}

	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(leveldb.CombineBucketAndKey(bucket, key), hex)
	})
}

func (db *BadgerDB) PutInBatch(records []interfaces.Record) error {
	return db.db.Update(func(txn *badger.Txn) error {
		for _, v := range records {
			hex, err := v.Data.MarshalBinary()
			if err != nil {
				return err
			}
			err = txn.Set(leveldb.CombineBucketAndKey(v.Bucket, v.Key), hex)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *BadgerDB) Clear(bucket []byte) error {

	return fmt.Errorf("Not implemented")
}

func (db *BadgerDB) ListAllKeys(bucket []byte) (keys [][]byte, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, k)
		}
		return nil
	})
	return
}

func (db *BadgerDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	keys := make([][]byte, 0)
	result := make([]interfaces.BinaryMarshallableAndCopyable, 0)
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(bucket)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}

			tmp := sample.New()
			err = tmp.UnmarshalBinary(v)
			if err != nil {
				return err
			}
			keys = append(keys, item.Key())
			result = append(result, tmp)
		}
		return nil
	})
	return result, nil, err
}

// We have to make accommodation for many Init functions.  But what we really
// want here is:
//
//      Init(bucketList [][]byte, filename string)
//
func (db *BadgerDB) Init(filedir string) error {
	opts := badger.DefaultOptions
	opts.Dir = filedir
	opts.ValueDir = filedir

	bdb, err := badger.Open(opts)
	if err != nil {
		return err
	}
	db.db = bdb
	return err
}

func (db *BadgerDB) DoesKeyExist(bucket, key []byte) (bool, error) {
	var result = false
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(leveldb.CombineBucketAndKey(bucket, key))
		if err == badger.ErrKeyNotFound {
			return nil // default to false
		}
		if err != nil {
			return err
		}
		result = true
		return nil
	})
	return result, err
}
