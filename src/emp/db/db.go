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
	"fmt"
	"github.com/mxk/go-sqlite/sqlite3"
	"sync"
)

// Database Connection
var dbConn *sqlite3.Conn
var mutex *sync.Mutex

func Initialize(log chan string, dbFile string) error {
	var err error
	if dbConn != nil {
		return nil
	}

	mutex = new(sync.Mutex)

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

	err = dbConn.Exec("CREATE TABLE IF NOT EXISTS pub (hash BLOB NOT NULL UNIQUE, addrHash BLOB NOT NULL, timestamp INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY (hash))")
	if err != nil {
		log <- fmt.Sprintf("Error setting up pub schema... %s", err)
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
	mutex.Lock()

	for s, err := dbConn.Query("SELECT hash FROM pubkey"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = PUBKEY
	}

	for s, err := dbConn.Query("SELECT hash FROM msg"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = MSG
	}

	for s, err := dbConn.Query("SELECT hash FROM pub"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = PUB
	}

	for s, err := dbConn.Query("SELECT hash FROM purge"); err == nil; err = s.Next() {
		var hash []byte
		s.Scan(&hash) // Assigns 1st column to rowid, the rest to row
		hashList[string(hash)] = PURGE
	}

	mutex.Unlock()
	return nil
}

func Cleanup() {
	dbConn.Close()
	dbConn = nil
	hashList = nil
}
