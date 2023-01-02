package server

import (
	"github.com/shurcooL/githubv4"
)

type ServerImplementation struct {
	GithubV4Client    *githubv4.Client
	ProjectOwner      string
	ProjectOwnerIsOrg bool
	ProjectNumber     int64
	ProjectID         string
	ImpactTypes       []string
	LastPhase         string
}
