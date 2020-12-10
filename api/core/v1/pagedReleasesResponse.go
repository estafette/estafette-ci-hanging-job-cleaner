package core

import (
	contracts "github.com/estafette/estafette-ci-contracts"
)

type PagedReleasesResponse struct {
	Items      []*contracts.Release `json:"items"`
	Pagination contracts.Pagination `json:"pagination"`
}
