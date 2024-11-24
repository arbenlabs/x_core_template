package service

import (
	"x/core/internal/persist"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/rs/zerolog"
)

type Service struct {
	p   *persist.PGStore
	z   *zerolog.Logger
	cld *cloudinary.Cloudinary
}

func NewService(
	store *persist.PGStore,
	logger *zerolog.Logger,
	cloud *cloudinary.Cloudinary,
) *Service {
	return &Service{
		p:   store,
		z:   logger,
		cld: cloud,
	}
}
