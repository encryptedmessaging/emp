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

package localdb

import (
	"emp/objects"
	"fmt"
	"github.com/mxk/go-sqlite/sqlite3"
	"sync"
)

// Database Connection
var LocalDB *sqlite3.Conn
var localMutex *sync.Mutex

func Initialize(log chan string, dbFile string) error {
	var err error
	if LocalDB != nil {
		return nil
	}

	localMutex = new(sync.Mutex)

	// Create Database Connection
	LocalDB, err = sqlite3.Open(dbFile)
	if err != nil || LocalDB == nil {
		log <- fmt.Sprintf("Error opening sqlite database at %s... %s", dbFile, err)
		LocalDB = nil
		return err
	}

	// Create Database Schema

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS addressbook (hash BLOB NOT NULL UNIQUE, address BLOB NOT NULL UNIQUE, registered INTEGER NOT NULL, pubkey BLOB, privkey BLOB, label TEXT, subscribed INTEGER NOT NULL, PRIMARY KEY (hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up addressbook schema... %s", err)
		LocalDB = nil
		return err
	}

	// Migration, Ignore error
	LocalDB.Exec("ALTER TABLE addressbook ADD COLUMN subscribed INTEGER NOT NULL DEFAULT 0")
	LocalDB.Exec("ALTER TABLE addressbook ADD COLUMN encprivkey BLOB")

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS msg (txid_hash BLOB NOT NULL, recipient BLOB, timestamp INTEGER, box INTEGER, encrypted BLOB, decrypted BLOB, purged INTEGER, sender BLOB, PRIMARY KEY (txid_hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up msg schema... %s", err)
		LocalDB = nil
		return err
	}

	if hashList == nil {
		hashList = make(map[string]int)
		return populateHashes()
	}

	if LocalDB == nil || hashList == nil {
		fmt.Println("ERROR! ERROR! WTF!!! SHOULD BE INITIALIZED!")
	}

	return nil
}

func populateHashes() error {
	for s, err := LocalDB.Query("SELECT hash FROM addressbook"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = ADDRESS
	}

	for s, err := LocalDB.Query("SELECT txid_hash, box FROM msg"); err == nil; err = s.Next() {
		var hash []byte
		var box int
		s.Scan(&hash, &box) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = box
	}

	return nil
}

func Cleanup() {
	LocalDB.Close()
	LocalDB = nil
	hashList = nil
}

const (
	INBOX    = iota
	OUTBOX   = iota
	SENDBOX  = iota
	ADDRESS  = iota
	NOTFOUND = iota
)

// Hash List
var hashList map[string]int

func Add(hashObj objects.Hash, hashType int) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		hashList[hash] = hashType
	}
}

func Del(hashObj objects.Hash) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		delete(hashList, hash)
	}
}

func Contains(hashObj objects.Hash) int {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		hashType, ok := hashList[hash]
		if ok {
			return hashType
		} else {
			return NOTFOUND
		}
	}
	return NOTFOUND
}
