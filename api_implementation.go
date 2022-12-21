package main

import (
	"context"
	"fmt"

	"github.com/joshmue/scs-status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/shurcooL/githubv4"
)


func (s *ServerImplementation) fillProjectID() error {
	var query struct {
		User struct {
			ProjectV2 struct {
				Id string `graphql:"id"`
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
	return nil
}

func (s *ServerImplementation) GetComponents(ctx echo.Context) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetImpacttypes(ctx echo.Context) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetPhases(ctx echo.Context) error {
	var query struct {
		User struct {
			ProjectV2 struct {
				Field struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"field(name: \"Status\")"`
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
	phases := []api.IncidentPhase{}
	for _, phase := range query.User.ProjectV2.Field.ProjectV2SingleSelectField.Options {
		phases = append(phases, phase.Name)
	}
	return ctx.JSON(200, phases)
}
