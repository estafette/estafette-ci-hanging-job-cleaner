package cleaner

import (
	"context"
	"time"

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

	maxAgeMinutes := float64(6*60 - 15)

	pageNumber := 1
	pageSize := 12
	for {
		pagedBuilds, err := s.estafetteciapiClient.GetRunningBuilds(ctx, pageNumber, pageSize)
		if err != nil {
			return err
		}

		// cancel builds close to 6 hours old (max lifetime of their jwt and last chance to send their logs to the api)
		for _, b := range pagedBuilds.Items {
			if b == nil {
				continue
			}
			if time.Now().UTC().Sub(b.InsertedAt).Minutes() > maxAgeMinutes {
				err = s.estafetteciapiClient.CancelBuild(ctx, b)
				if err != nil {
					return err
				}
			}
		}

		if pagedBuilds.Pagination.TotalPages <= pageNumber {
			break
		}

		pageNumber++
	}

	pageNumber = 1
	pageSize = 12
	for {
		pagedReleases, err := s.estafetteciapiClient.GetRunningReleases(ctx, pageNumber, pageSize)
		if err != nil {
			return err
		}

		// cancel releases close to 6 hours old (max lifetime of their jwt and last chance to send their logs to the api)
		for _, r := range pagedReleases.Items {
			if r == nil || r.InsertedAt == nil {
				continue
			}
			if time.Now().UTC().Sub(*r.InsertedAt).Minutes() > maxAgeMinutes {
				err = s.estafetteciapiClient.CancelRelease(ctx, r)
				if err != nil {
					return err
				}
			}
		}

		if pagedReleases.Pagination.TotalPages <= pageNumber {
			break
		}

		pageNumber++
	}

	return nil
}
