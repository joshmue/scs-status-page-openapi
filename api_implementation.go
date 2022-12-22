package main

import (
	"context"
	"fmt"
	"strings"

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
	var query struct {
		Node struct {
			ProjectV2 struct {
				Repositories struct {
					Nodes []struct {
						Labels struct {
							Nodes []struct {
								Name        string
								Id          string
								Description string
							}
						} `graphql:"labels(first: 100)"`
					}
				} `graphql:"repositories(first: 100)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectid)"`
	}
	err := s.GithubV4Client.Query(
		context.Background(),
		&query,
		map[string]interface{}{
			"projectid": githubv4.ID(s.ProjectID),
		},
	)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	components := []api.Component{}
	for repo := range query.Node.ProjectV2.Repositories.Nodes {
		for label := range query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes {
			if strings.HasPrefix(query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Name, "component:") {
				component := api.Component{
					Id:          &query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Name,
					DisplayName: &query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Description,
				}
				components = append(components, component)
			}
		}
	}
	return ctx.JSON(200, components)
}
func (s *ServerImplementation) GetImpacttypes(ctx echo.Context) error {
	var query struct {
		User struct {
			ProjectV2 struct {
				Field struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"field(name: \"Impact Type\")"`
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
		ctx.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	impactTypes := []api.IncidentImpactType{}
	for _, phase := range query.User.ProjectV2.Field.ProjectV2SingleSelectField.Options {
		impactTypes = append(impactTypes, phase.Name)
	}
	return ctx.JSON(200, impactTypes)
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
		ctx.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	phases := []api.IncidentPhase{}
	for _, phase := range query.User.ProjectV2.Field.ProjectV2SingleSelectField.Options {
		phases = append(phases, phase.Name)
	}
	return ctx.JSON(200, phases)
}
