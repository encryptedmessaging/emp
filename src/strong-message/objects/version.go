package objects

import (
  "time"
)

type Version struct {
  Version uint32
  Timestamp time.Time
  UserAgent string
}
