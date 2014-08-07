package db

import "emp/objects"

const (
	PUBKEY   = iota
	PURGE    = iota
	MSG      = iota
	PUBKEYRQ = iota
	PUB      = iota
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

func Delete(hashObj objects.Hash) {
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
