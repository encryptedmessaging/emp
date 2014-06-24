package localdb

import (
	"fmt"
	"github.com/mxk/go-sqlite/sqlite3"
)

// Database Connection
var LocalDB *sqlite3.Conn

func Initialize(log chan string, dbFile string) error {
	var err error
	if LocalDB != nil {
		return nil
	}

	// Create Database Connection
	LocalDB, err = sqlite3.Open(dbFile)
	if err != nil || LocalDB == nil {
		log <- fmt.Sprintf("Error opening sqlite database at %s... %s", dbFile, err)
		LocalDB = nil
		return err
	}

	// Create Database Schema

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS addressbook (hash BLOB NOT NULL UNIQUE, address BLOB NOT NULL UNIQUE, registered INTEGER NOT NULL, pubkey BLOB, privkey BLOB, PRIMARY KEY (hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up addressbook schema... %s", err)
		LocalDB = nil
		return err
	}

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS msg (txid_hash BLOB NOT NULL UNIQUE, encrypted BLOB, decrypted BLOB, PRIMARY KEY (txid_hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up msg schema... %s", err)
		LocalDB = nil
		return err
	}

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS inbox (txid_hash BLOB NOT NULL, timestamp INTEGER NOT NULL, sender BLOB, recipient BLOB NOT NULL, UNIQUE(txid_hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up inbox schema... %s", err)
		LocalDB = nil
		return err
	}

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS outbox (txid_hash BLOB NOT NULL, timestamp INTEGER NOT NULL, sender BLOB, recipient BLOB NOT NULL, UNIQUE(txid_hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up outbox schema... %s", err)
		LocalDB = nil
		return err
	}

	err = LocalDB.Exec("CREATE TABLE IF NOT EXISTS sendbox (txid_hash BLOB NOT NULL, timestamp INTEGER NOT NULL, sender BLOB, recipient BLOB NOT NULL, UNIQUE(txid_hash) ON CONFLICT REPLACE)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up sendbox schema... %s", err)
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

	for s, err := LocalDB.Query("SELECT txid_hash FROM inbox"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = INBOX
	}

	for s, err := LocalDB.Query("SELECT txid_hash FROM outbox"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = OUTBOX
	}

	for s, err := LocalDB.Query("SELECT txid_hash FROM sendbox"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = SENDBOX
	}
	for s, err := LocalDB.Query("SELECT hash FROM addressbook"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = ADDRESS
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

func Add(hash string, hashType int) {
	if hashList != nil {
		hashList[hash] = hashType
	}
}

func Del(hash string) {
	if hashList != nil {
		delete(hashList, hash)
	}
}

func Contains(hash string) int {
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
