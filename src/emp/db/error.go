/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

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
