package objects



type Message struct {
  Magic unit32
  Type string
  Payload []byte
}
