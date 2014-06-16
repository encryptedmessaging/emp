package db

const (
	PUBKEY = iota
	PURGE = iota
	MSG = iota
	NOTFOUND = iota
)

// Hash List
var hashList map[string]int

func Add(hash string, hashType int) {
	if hashList != nil {
		hashList[hash] = hashType
	}
}

func Delete(hash string) {
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