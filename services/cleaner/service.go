package cleaner

import (
	"context"
	"time"

	estafetteciapi "github.com/estafette/estafette-ci-hanging-job-cleaner/clients/estafetteciapi"
	kubernetesapi "github.com/estafette/estafette-ci-hanging-job-cleaner/clients/kubernetesapi"
)

type Service interface {
	Init(ctx context.Context) (err error)
	Clean(ctx context.Context) (err error)
}

func NewService(estafetteciapiClient estafetteciapi.Client, kubernetesapiClient kubernetesapi.Client) (Service, error) {
	return &service{
		estafetteciapiClient: estafetteciapiClient,
	}, nil
}

type service struct {
	estafetteciapiClient estafetteciapi.Client
	kubernetesapiClient  kubernetesapi.Client
}

func (s *service) Init(ctx context.Context) (err error) {
	_, err = s.estafetteciapiClient.GetToken(ctx)
	if err != nil {
		return
	}

	return
}

func (s *service) Clean(ctx context.Context) (err error) {

	err = s.cleanBuilds(ctx)
	if err != nil {
		return
	}

	err = s.cleanReleases(ctx)
	if err != nil {
		return
	}

	return nil
}

func (s *service) cleanBuilds(ctx context.Context) (err error) {

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

	return nil
}

func (s *service) cleanReleases(ctx context.Context) (err error) {

	maxAgeMinutes := float64(6*60 - 15)
	pageNumber := 1
	pageSize := 12

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

func (s *service) cleanJobs(ctx context.Context) (err error) {

	maxAgeMinutes := float64(6*60 + 15)

	jobs, err := s.kubernetesapiClient.GetJobs(ctx)
	if err != nil {
		return err
	}

	for _, j := range jobs {
		// jobs that are older than max jwt lifetime missed being canceled properly, delete them
		if time.Now().UTC().Sub(j.CreationTimestamp.Time).Minutes() > maxAgeMinutes {
			err = s.kubernetesapiClient.DeleteJob(ctx, j)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *service) cleanConfigMaps(ctx context.Context) (err error) {

	maxAgeMinutes := float64(6*60 + 15)

	configmaps, err := s.kubernetesapiClient.GetConfigMaps(ctx)
	if err != nil {
		return err
	}

	for _, c := range configmaps {
		// configmaps that are older than max jwt lifetime missed being canceled properly, delete them
		if time.Now().UTC().Sub(c.CreationTimestamp.Time).Minutes() > maxAgeMinutes {
			err = s.kubernetesapiClient.DeleteConfigMap(ctx, c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *service) cleanSecrets(ctx context.Context) (err error) {

	maxAgeMinutes := float64(6*60 + 15)

	secrets, err := s.kubernetesapiClient.GetSecrets(ctx)
	if err != nil {
		return err
	}

	for _, sec := range secrets {
		// secrets that are older than max jwt lifetime missed being canceled properly, delete them
		if time.Now().UTC().Sub(sec.CreationTimestamp.Time).Minutes() > maxAgeMinutes {
			err = s.kubernetesapiClient.DeleteSecret(ctx, sec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
