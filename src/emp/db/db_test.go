/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package db

import (
	"fmt"
	"os/exec"
	"emp/objects"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	// Start Logger
	log := make(chan string, 100)
	go func() {
		for {
			log_stmt := <-log
			fmt.Println(log_stmt)
		}
	}()

	err := Initialize(log, "testdb.db")
	if dbConn == nil || hashList == nil {
		fmt.Println("ERROR! ERROR! WTF!!!")
	}

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	txid := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p'}
	purgeHash := objects.MakeHash(txid)
	pubHash := objects.MakeHash([]byte{'e', 'f', 'g', 'h'})

	if Contains(purgeHash) != NOTFOUND {
		fmt.Println("Purge Hash already in list...")
		t.FailNow()
	}
	if Contains(pubHash) != NOTFOUND {
		fmt.Println("Pubkey Hash already in list...")
		t.FailNow()
	}

	pub := new(objects.EncryptedPubkey)
	pub.AddrHash = pubHash
	pub.IV = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	pub.Payload = []byte{'a', 'b', 'c', 'd'}

	err = AddPubkey(log, *pub)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	if Contains(pubHash) != PUBKEY {
		fmt.Println("Pubkey not in hash list")
		time.Sleep(time.Millisecond)
		t.FailNow()
	}

	RemoveHash(log, pubHash)

	if Contains(pubHash) != NOTFOUND {
		fmt.Println("Pubkey stuck in hash list")
		t.FailNow()
	}

	purge := new(objects.Purge)
	copy(purge.Txid[:], txid)

	err = AddPurge(log, *purge)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	if Contains(purgeHash) != PURGE {
		fmt.Println("Purge not in hash list")
		t.FailNow()
	}

	RemoveHash(log, purgeHash)

	if Contains(purgeHash) != NOTFOUND {
		fmt.Println("Purge stuck in hash list")
		t.FailNow()
	}

	Cleanup()

	// Remove DB
	err = exec.Command("rm", "testdb.db").Run()

}
