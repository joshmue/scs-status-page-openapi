package server

import (
	"context"
	"fmt"

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

func (s *ServerImplementation) FillProjectID() error {
	// TODO
	// Make this also accept organizations
	if s.ProjectOwnerIsOrg {
		return fmt.Errorf("support for organizations owning projects not yet implemented")
	}
	var query struct {
		User struct {
			ProjectV2 struct {
				Id     string
				Number int64
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $user)"`
	}
	err := s.GithubV4Client.Query(
		context.Background(),
		&query,
		map[string]interface{}{
			"user":   githubv4.String(s.ProjectOwner),
			"number": githubv4.Int(s.ProjectNumber),
		},
	)
	if err != nil {
		return err
	}
	s.ProjectID = query.User.ProjectV2.Id
	s.ProjectNumber = query.User.ProjectV2.Number
	return nil
}