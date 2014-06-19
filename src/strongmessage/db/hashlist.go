package db

const (
	PUBKEY   = iota
	PURGE    = iota
	MSG      = iota
	PUBKEYRQ = iota
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

func HashCache() []string {
	if hashList == nil {
		return nil
	}
	ret := make([]string, 0, len(hashList))
	for key, _ := range hashList {
		ret = append(ret, key)
	}
	return ret
}

func HashCopy() map[string]int {
	cpy := make(map[string]int)
	for k, v := range hashList {
		cpy[k] = v
	}

	return cpy
}
