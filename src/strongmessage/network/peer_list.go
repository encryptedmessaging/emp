package network

const DEBUG = true

type PeerList struct {
	Peers []Peer `json:"peers"`
}
