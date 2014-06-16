package db

import (
	"testing"
	"fmt"
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
	if DBConn == nil  || HashList == nil {
		fmt.Println("ERROR! ERROR! WTF!!!")
	}
	
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	purgeHash := []byte{'a', 'b', 'c', 'd'}
	pubHash := []byte{'e', 'f', 'g', 'h'}

	if Contains(string(purgeHash)) != NOTFOUND {
		fmt.Println("Purge Hash already in list...")
		t.FailNow()
	}
	if Contains(string(pubHash)) != NOTFOUND {
		fmt.Println("Pubkey Hash already in list...")
		t.FailNow()
	}

	err = AddPubkey(log, pubHash, pubHash)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	if Contains(string(pubHash)) != PUBKEY {
		fmt.Println("Pubkey not in hash list")
		time.Sleep(time.Millisecond)
		t.FailNow()
	}

	RemoveHash(log, pubHash)

	if Contains(string(pubHash)) != NOTFOUND {
		fmt.Println("Pubkey stuck in hash list")
		t.FailNow()
	}

	err = AddPubkey(log, purgeHash, purgeHash)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		t.FailNow()
	}

	if Contains(string(purgeHash)) != PURGE {
		fmt.Println("Purge not in hash list")
		t.FailNow()
	}

	RemoveHash(log, purgeHash)

	if Contains(string(purgeHash)) != NOTFOUND {
		fmt.Println("Purge stuck in hash list")
		t.FailNow()
	}

	Cleanup()
}