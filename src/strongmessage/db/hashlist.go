package db

const (
	PUBKEY = iota
	PURGE = iota
	MSG = iota
	NOTFOUND = iota
)

// Hash List
var HashList map[string]int

func Add(hash string, hashType int) {
	if HashList != nil {
		HashList[hash] = hashType
	}
}

func Delete(hash string) {
	if HashList != nil {
		delete(HashList, hash)
	}
}

func Contains(hash string) int {
	if HashList != nil {
		hashType, ok := HashList[hash]
		if ok {
			return hashType
		} else {
			return NOTFOUND
		}
	}
	return NOTFOUND
}