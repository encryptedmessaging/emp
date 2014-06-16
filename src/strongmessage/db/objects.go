package db

import (
	"fmt"
	"strongmessage/objects"
	"time"
)

func AddPubkey(log chan string, hash, payload []byte) error {
	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(string(hash)) == PUBKEY {
		return nil
	}

	err := dbConn.Exec("INSERT INTO pubkey VALUES (?, ?)", hash, payload)
	if err != nil {
		log <- fmt.Sprintf("Error inserting pubkey into db... %s", err)
		return err
	}

	Add(string(hash), PUBKEY)
	return nil
}

func GetPubkey(log chan string, hash []byte) ([]byte, error) {
	if hashList == nil || dbConn == nil {
		return nil, DBError(EUNINIT)
	}
	if hashList[string(hash)] != PUBKEY {
		return nil, nil
	}

	for s, err := dbConn.Query("SELECT payload FROM pubkey WHERE hash=?", hash); err == nil; err = s.Next() {
		var payload []byte
		s.Scan(payload) // Assigns 1st column to rowid, the rest to row
		return payload, nil
	}
	// Not Found
	return nil, nil
}

func AddPurge(log chan string, hash, txid []byte) error {
	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(string(hash)) == PURGE {
		return nil
	}

	err := dbConn.Exec("INSERT INTO purge VALUES (?, ?)", hash, txid)
	if err != nil {
		log <- fmt.Sprintf("Error inserting purge into db... %s", err)
		return err
	}

	Add(string(hash), PURGE)
	return nil
}

func GetPurge(log chan string, hash []byte) ([]byte, error) {
	if hashList == nil || dbConn == nil {
		return nil, DBError(EUNINIT)
	}
	if hashList[string(hash)] != PURGE {
		return nil, nil
	}

	for s, err := dbConn.Query("SELECT txid FROM purge WHERE hash=?", hash); err == nil; err = s.Next() {
		var txid []byte
		s.Scan(txid) // Assigns 1st column to rowid, the rest to row
		return txid, nil
	}
	// Not Found
	return nil, nil
}

func AddMessage(log chan string, msg *objects.Message) error {
	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}
	if Contains(string(msg.TxidHash)) == MSG {
		return nil
	}

	err := dbConn.Exec("INSERT INTO purge VALUES (?, ?, ?, ?)", msg.TxidHash, msg.AddrHash, msg.Timestamp.Unix(), msg.Content.GetBytes())
	if err != nil {
		log <- fmt.Sprintf("Error inserting purge into db... %s", err)
		return err
	}

	Add(string(msg.TxidHash), PURGE)
	return nil

}

func GetMessage(log chan string, hash []byte) (*objects.Message, error) {
	if hashList == nil || dbConn == nil {
		return nil, DBError(EUNINIT)
	}
	if hashList[string(hash)] != MSG {
		return nil, nil
	}

	msg := new(objects.Message)

	for s, err := dbConn.Query("SELECT * FROM msg WHERE hash=?", hash); err == nil; err = s.Next() {
		var timestamp int64
		var encrypted []byte
		s.Scan(msg.TxidHash, msg.AddrHash, &timestamp, encrypted)

		msg.Timestamp = time.Unix(timestamp, 0)
		msg.Content = objects.EncryptedFromBytes(encrypted)

		return msg, nil
	}
	// Not Found
	return nil, nil
}

func RemoveHash(log chan string, hash []byte) error {
	if hashList == nil || dbConn == nil {
		return DBError(EUNINIT)
	}

	var sql string

	switch Contains(string(hash)) {
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

	Delete(string(hash))
	return nil
}
