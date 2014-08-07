package objects

type ChannelDetail struct {
	String       string `json:"address"`
	Address      []byte `json:"address_bytes"`
	Pubkey       []byte `json:"public_key"`
	Privkey      []byte `json:"private_key"`
}