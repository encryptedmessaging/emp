package db

import (
	"github.com/mxk/go-sqlite/sqlite3"
	"fmt"
)

// Database Connection
var DBConn *sqlite3.Conn

func Initialize(log chan string, dbFile string) error {
	if DBConn != nil {
		return nil
	}

	// Create Database Connection
	DBConn, err := sqlite3.Open(dbFile)
	if err != nil || DBConn == nil{
		log <- fmt.Sprintf("Error opening sqlite database at %s... %s", dbFile, err)
		DBConn = nil
		return err
	}


	// Create Database Schema
	err = DBConn.Exec("CREATE TABLE IF NOT EXISTS pubkey (hash BLOB NOT NULL UNIQUE, payload BLOB NOT NULL, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up pubkey schema... %s", err)
		DBConn = nil
		return err
	}

	err = DBConn.Exec("CREATE TABLE IF NOT EXISTS purge (hash BLOB NOT NULL UNIQUE, txid BLOB NOT NULL UNIQUE, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up purge schema... %s", err)
		DBConn = nil
		return err
	}

	err = DBConn.Exec("CREATE TABLE IF NOT EXISTS msg (hash BLOB NOT NULL UNIQUE, addrHash BLOB NOT NULL, timestamp INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up msg schema... %s", err)
		DBConn = nil
		return err
	}

	err = DBConn.Exec("CREATE TABLE IF NOT EXISTS peer (ip BLOB NOT NULL, port INTEGER NOT NULL, port_admin INTEGER NOT NULL, last_seen INTEGER NOT NULL, id INTEGER PRIMARY KEY AUTOINCREMENT)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up peer schema... %s", err)
		DBConn = nil
		return err
	}

	err = DBConn.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ip_index ON peer (ip, port, port_admin)")
	if err != nil {
		log <- fmt.Sprintf("Error setting up peer index... %s", err)
		DBConn = nil
		return err
	}

	if HashList == nil {
		HashList = make(map[string]int)
		//return populateHashes()
	}

	if DBConn == nil  || HashList == nil {
		fmt.Println("ERROR! ERROR! WTF!!! SHOULD BE INITIALIZED!")
	}

	return nil
}

func populateHashes() error {

	for s, err := DBConn.Query("SELECT hash FROM pubkey"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash)     // Assigns 1st column to rowid, the rest to row
		HashList[string(hash)] = PUBKEY
	}

	for s, err := DBConn.Query("SELECT hash FROM msg"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash)     // Assigns 1st column to rowid, the rest to row
		HashList[string(hash)] = MSG
	}

	for s, err := DBConn.Query("SELECT hash FROM purge"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(hash)     // Assigns 1st column to rowid, the rest to row
		HashList[string(hash)] = PURGE
	}

	return nil
}

func Cleanup() {
	DBConn.Close()
	DBConn = nil
	HashList = nil
}