package core

import (
	contracts "github.com/estafette/estafette-ci-contracts"
)

type PagedBuildResponse struct {
	Items      []*contracts.Build   `json:"items"`
	Pagination contracts.Pagination `json:"pagination"`
}
