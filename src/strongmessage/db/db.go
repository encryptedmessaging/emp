package db

import (
	"fmt"
	"github.com/mxk/go-sqlite/sqlite3"
)

// Database Connection
var dbConn *sqlite3.Conn

func Initialize(log chan string, dbFile string) error {
	var err error
	if dbConn != nil {
		return nil
	}

	// Create Database Connection
	dbConn, err = sqlite3.Open(dbFile)
	if err != nil || dbConn == nil {
		log <- fmt.Sprintf("Error opening sqlite database at %s... %s", dbFile, err)
		dbConn = nil
		return err
	}

	// Create Database Schema
	err = dbConn.Exec("CREATE TABLE IF NOT EXISTS pubkey (hash BLOB NOT NULL UNIQUE, payload BLOB NOT NULL, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up pubkey schema... %s", err)
		dbConn = nil
		return err
	}

	err = dbConn.Exec("CREATE TABLE IF NOT EXISTS purge (hash BLOB NOT NULL UNIQUE, txid BLOB NOT NULL UNIQUE, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up purge schema... %s", err)
		dbConn = nil
		return err
	}

	err = dbConn.Exec("CREATE TABLE IF NOT EXISTS msg (hash BLOB NOT NULL UNIQUE, addrHash BLOB NOT NULL, timestamp INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up msg schema... %s", err)
		dbConn = nil
		return err
	}

	err = dbConn.Exec("CREATE TABLE IF NOT EXISTS peer (ip BLOB NOT NULL, port INTEGER NOT NULL, port_admin INTEGER NOT NULL, last_seen INTEGER NOT NULL, id INTEGER PRIMARY KEY AUTOINCREMENT)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up peer schema... %s", err)
		dbConn = nil
		return err
	}

	err = dbConn.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ip_index ON peer (ip, port, port_admin)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up peer index... %s", err)
		dbConn = nil
		return err
	}

	if hashList == nil {
		hashList = make(map[string]int)
		return populateHashes()
	}

	if dbConn == nil || hashList == nil {
		fmt.Println("ERROR! ERROR! WTF!!! SHOULD BE INITIALIZED!")
	}

	return nil
}

func populateHashes() error {

	for s, err := dbConn.Query("SELECT hash FROM pubkey"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = PUBKEY
	}

	for s, err := dbConn.Query("SELECT hash FROM msg"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = MSG
	}

	for s, err := dbConn.Query("SELECT hash FROM purge"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = PURGE
	}

	return nil
}

func Cleanup() {
	dbConn.Close()
	dbConn = nil
	hashList = nil
}
