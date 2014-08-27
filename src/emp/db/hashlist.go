/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package db

import "emp/objects"

const (
	PUBKEY   = iota // Encrypted Public Key
	PURGE    = iota // Purged Message
	MSG      = iota // Basic Message
	PUBKEYRQ = iota // Public Key Request
	PUB      = iota // Published Message
	NOTFOUND = iota // Object not in hash list.
)

// Hash List
var hashList map[string]int

// Add an object to the hash list with a given type.
func Add(hashObj objects.Hash, hashType int) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		hashList[hash] = hashType
	}
}

// Remove an object from the hash list.
func Delete(hashObj objects.Hash) {
	hash := string(hashObj.GetBytes())
	if hashList != nil {
		delete(hashList, hash)
	}
}

// Return the type the item in the hash list (see constants).
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

// List of all hashes in the hash list.
func ObjList() *objects.Obj {
	if hashList == nil {
		return nil
	}

	ret := new(objects.Obj)
	ret.HashList = make([]objects.Hash, 0, 0)

	hash := new(objects.Hash)

	for key, _ := range hashList {
		hash.FromBytes([]byte(key))
		ret.HashList = append(ret.HashList, *hash)
	}
	return ret
}
