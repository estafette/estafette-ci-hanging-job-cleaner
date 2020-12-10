package cleaner

import (
	"context"

	estafetteciapi "github.com/estafette/estafette-ci-hanging-job-cleaner/clients/estafetteciapi"
)

type Service interface {
	Clean(ctx context.Context) (err error)
}

func NewService(estafetteciapiClient estafetteciapi.Client) Service {
	return &service{
		estafetteciapiClient: estafetteciapiClient,
	}
}

type service struct {
	estafetteciapiClient estafetteciapi.Client
}

func (s *service) Clean(ctx context.Context) (err error) {
	return nil
}
