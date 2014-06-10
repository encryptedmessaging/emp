package objects

import (
  "time"
)

type Version struct {
  Version unit32
  Timestamp time.Time
  UserAgent string
}
