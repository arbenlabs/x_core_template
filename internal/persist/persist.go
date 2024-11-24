package persist

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type PGStore struct {
	DB *gorm.DB
	z  *zerolog.Logger
}

func NewPGStore(db *gorm.DB, log *zerolog.Logger) *PGStore {
	return &PGStore{
		DB: db,
		z:  log,
	}
}
