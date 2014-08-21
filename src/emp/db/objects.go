/**
    Copyright 2014 JARST, LLC
    
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
**/

package db

import (
	"crypto/sha512"
	"fmt"
	"emp/objects"
	"time"
)

func SweepMessages(duration time.Duration) error {
	mutex.Lock()
	defer mutex.Unlock()

	deadline := time.Now().Add(-duration).Unix()

	return dbConn.Exec("DELETE FROM msg WHERE timestamp <= ?", deadline)
}

func AddPubkey(log chan string, pubkey objects.EncryptedPubkey) error {
	mutex.Lock()
	defer mutex.Unlock()

	hash := pubkey.AddrHash.GetBytes()
	payload := append(pubkey.IV[:], pubkey.Payload...)

	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(pubkey.AddrHash) == PUBKEY {
		return nil
	}

	err := dbConn.Exec("INSERT INTO pubkey VALUES (?, ?)", hash, payload)
	if err != nil {
		log <- fmt.Sprintf("Error inserting pubkey into db... %s", err)
		return err
	}

	Add(pubkey.AddrHash, PUBKEY)
	return nil
}

func GetPubkey(log chan string, addrHash objects.Hash) *objects.EncryptedPubkey {
	mutex.Lock()
	defer mutex.Unlock()

	hash := addrHash.GetBytes()

	if hashList == nil || dbConn == nil {
		return nil
	}
	if hashList[string(hash)] != PUBKEY {
		return nil
	}

	for s, err := dbConn.Query("SELECT payload FROM pubkey WHERE hash=?", hash); err == nil; err = s.Next() {
		var payload []byte
		s.Scan(&payload) // Assigns 1st column to rowid, the rest to row
		pub := new(objects.EncryptedPubkey)
		pub.AddrHash = addrHash
		copy(pub.IV[:], payload[:16])
		pub.Payload = payload[16:]
		return pub
	}
	// Not Found
	return nil
}

func AddPurge(log chan string, p objects.Purge) error {
	mutex.Lock()
	defer mutex.Unlock()

	txid := p.GetBytes()
	hashArr := sha512.Sum384(txid)
	hash := hashArr[:]

	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	hashObj := new(objects.Hash)
	hashObj.FromBytes(hash)

	if Contains(*hashObj) == PURGE {
		return nil
	}

	err := dbConn.Exec("INSERT INTO purge VALUES (?, ?)", hash, txid)
	if err != nil {
		log <- fmt.Sprintf("Error inserting purge into db... %s", err)
		return err
	}

	Add(*hashObj, PURGE)
	return nil
}

func GetPurge(log chan string, txidHash objects.Hash) *objects.Purge {
	mutex.Lock()
	defer mutex.Unlock()

	hash := txidHash.GetBytes()

	if hashList == nil || dbConn == nil {
		return nil
	}
	if hashList[string(hash)] != PURGE {
		return nil
	}

	for s, err := dbConn.Query("SELECT txid FROM purge WHERE hash=?", hash); err == nil; err = s.Next() {
		var txid []byte
		s.Scan(&txid) // Assigns 1st column to rowid, the rest to row
		p := new(objects.Purge)
		p.FromBytes(txid)
		return p
	}
	// Not Found
	return nil
}

func AddPub(log chan string, msg *objects.Message) error {
	mutex.Lock()
	defer mutex.Unlock()

	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(msg.TxidHash) == MSG {
		return nil
	}

	err := dbConn.Exec("INSERT INTO pub VALUES (?, ?, ?, ?)", msg.TxidHash.GetBytes(), msg.AddrHash.GetBytes(), msg.Timestamp.Unix(), msg.Content.GetBytes())
	if err != nil {
		log <- fmt.Sprintf("Error inserting message into db... %s", err)
		return err
	}

	Add(msg.TxidHash, PUB)
	return nil
}

func AddMessage(log chan string, msg *objects.Message) error {
	mutex.Lock()
	defer mutex.Unlock()

	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(msg.TxidHash) == MSG {
		return nil
	}

	err := dbConn.Exec("INSERT INTO msg VALUES (?, ?, ?, ?)", msg.TxidHash.GetBytes(), msg.AddrHash.GetBytes(), msg.Timestamp.Unix(), msg.Content.GetBytes())
	if err != nil {
		log <- fmt.Sprintf("Error inserting message into db... %s", err)
		return err
	}

	Add(msg.TxidHash, MSG)
	return nil

}

func GetMessage(log chan string, txidHash objects.Hash) *objects.Message {
	mutex.Lock()
	defer mutex.Unlock()

	hash := txidHash.GetBytes()

	if hashList == nil || dbConn == nil {
		return nil
	}
	if hashList[string(hash)] != MSG && hashList[string(hash)] != PUB {
		return nil
	}

	msg := new(objects.Message)

	for s, err := dbConn.Query("SELECT * FROM msg WHERE hash=?", hash); err == nil; err = s.Next() {
		var timestamp int64
		encrypted := make([]byte, 0, 0)
		txidhash := make([]byte, 0, 0)
		addrhash := make([]byte, 0, 0)
		s.Scan(&txidhash, &addrhash, &timestamp, &encrypted)

		msg.TxidHash.FromBytes(txidhash)
		msg.AddrHash.FromBytes(addrhash)
		msg.Timestamp = time.Unix(timestamp, 0)
		msg.Content.FromBytes(encrypted)

		return msg
	}
	// Not Found
	return nil
}

func RemoveHash(log chan string, hashObj objects.Hash) error {
	mutex.Lock()
	defer mutex.Unlock()

	hash := hashObj.GetBytes()

	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}

	var sql string

	switch Contains(hashObj) {
	case PUBKEY:
		sql = "DELETE FROM pubkey WHERE hash=?"
	case MSG:
		sql = "DELETE FROM msg WHERE hash=?"
	case PURGE:
		sql = "DELETE FROM purge WHERE hash=?"
	default:
		return nil
	}

	err := dbConn.Exec(sql, hash)
	if err != nil {
		log <- fmt.Sprintf("Error deleting hash from db... %s", err)
		return nil
	}

	Delete(hashObj)
	return nil
}
