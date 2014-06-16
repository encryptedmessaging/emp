package db

const (
	EUNINIT = iota
)

type DBError int

func (e DBError) Error() string {
	switch int(e) {
	case EUNINIT:
		return "Database or hash list not initialized! Please call Initialize()"
	default:
		return "Unknown error..."
	}
}