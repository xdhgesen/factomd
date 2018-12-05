package databaseOverlay

import (
	"encoding/binary"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (db *Overlay) SaveDirectoryBlockSignature(dbSig interfaces.IMsg, height uint32) error {
	if dbSig == nil {
		return nil
	}

	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes, height)
	bucket := append(DIRECTORYBLOCKSIGNATURE, heightBytes...)

	exists, err := db.DoesKeyExist(bucket, dbSig.GetRepeatHash().Bytes())
	if err != nil {
		return err
	} else if exists {
		return nil
	}

	batch := []interfaces.Record{}
	batch = append(batch, interfaces.Record{bucket, dbSig.GetRepeatHash().Bytes(), dbSig})

	err = db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) FetchDirectoryBlockSignatures(height uint32) ([]interfaces.IMsg, error) {
	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes, height)
	bucket := append(DIRECTORYBLOCKSIGNATURE, heightBytes...)

	keys, err := db.DB.ListAllKeys(bucket)
	if err != nil {
		return nil, err
	}

	var dbSigs []interfaces.IMsg
	for _, key := range keys {
		dbSig, err := db.DB.Get(bucket, key, new(messages.DirectoryBlockSignature))
		if err != nil {
			return dbSigs, err
		} else if dbSig == nil {
			return dbSigs, nil
		}
		dbSigs = append(dbSigs, dbSig.(interfaces.IMsg))
	}

	return dbSigs, nil
}

func (db *Overlay) DropDirectoryBlockSignatures(height uint32) error {
	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes, height)
	bucket := append(DIRECTORYBLOCKSIGNATURE, heightBytes...)
	err := db.DB.Clear(bucket)
	return err
}
