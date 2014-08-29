/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

// Package localdb provdes a local SQLite3 Database for the EMPLocal client.
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

// Initialize database with mutexes from file.
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

// Close and Cleanup database.
func Cleanup() {
	LocalDB.Close()
	LocalDB = nil
	hashList = nil
}

// Hash Types
const (
	INBOX    = iota // Incoming Messages
	OUTBOX   = iota // Outgoing, Unsent Messages
	SENDBOX  = iota // Outgoing, Sent Messages
	ADDRESS  = iota // EMP Addresses
	NOTFOUND = iota // Not Found in DB
)

// Hash List
var hashList map[string]int

// Add to global Hash List.
func Add(hashObj objects.Hash, hashType int) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		hashList[hash] = hashType
	}
}

// Delete hash from Hash List
func Del(hashObj objects.Hash) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		delete(hashList, hash)
	}
}

// Get type of object in Hash List
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
