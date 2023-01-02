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
	// TODO
	// Make this also accept organizations
	if s.ProjectOwnerIsOrg {
		return fmt.Errorf("support for organizations owning projects not yet implemented")
	}
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

func (s *ServerImplementation) GetComponent(ctx echo.Context, componentId string) error {
	return fmt.Errorf("not implemented")
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
					Id:          query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Name,
					DisplayName: query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Description,
				}
				components = append(components, component)
			}
		}
	}
	return ctx.JSON(200, components)
}
func (s *ServerImplementation) GetImpacttypes(ctx echo.Context) error {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Field struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"field(name: \"Impact Type\")"`
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
	impactTypes := []api.IncidentImpactType{}
	for _, phase := range query.Node.ProjectV2.Field.ProjectV2SingleSelectField.Options {
		impactTypes = append(impactTypes, phase.Name)
	}
	return ctx.JSON(200, impactTypes)
}
func (s *ServerImplementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	type projectItem struct {
		Type    string
		Content struct {
			Issue struct {
				Id    string
				Title string
			} `graphql:"... on Issue"`
		}
		Phase struct {
			ProjectV2ItemFieldSingleSelectValue struct {
				Name string
			} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
		} `graphql:"phase: fieldValueByName(name: \"Status\")"`
		ImpactType struct {
			ProjectV2ItemFieldSingleSelectValue struct {
				Name string
			} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
		} `graphql:"impacttype: fieldValueByName(name: \"Impact Type\")"`
		BeganAt struct {
			ProjectV2ItemFieldTextValue struct {
				Text string
			} `graphql:"... on ProjectV2ItemFieldTextValue"`
		} `graphql:"beganat: fieldValueByName(name: \"Began At\")"`
		EndedAt struct {
			ProjectV2ItemFieldTextValue struct {
				Text string
			} `graphql:"... on ProjectV2ItemFieldTextValue"`
		} `graphql:"endedat: fieldValueByName(name: \"Ended At\")"`
		Labels struct {
			ProjectV2ItemFieldLabelValue struct {
				Labels struct {
					Nodes []struct {
						Id string
					}
				} `graphql:"labels(first:10)"`
			} `graphql:"... on ProjectV2ItemFieldLabelValue"`
		} `graphql:"labels: fieldValueByName(name: \"Labels\")"`
	}
	var query struct {
		Node struct {
			ProjectV2 struct {
				Items struct {
					Nodes []projectItem
				} `graphql:"items(first: 10)"`
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

	// Map GraphQL output to OpenAPI Spec
	incidents := []api.Incident{}
	for itemKey := range query.Node.ProjectV2.Items.Nodes {
		beganAt, err := ParseTimeOrNil(query.Node.ProjectV2.Items.Nodes[itemKey].BeganAt.ProjectV2ItemFieldTextValue.Text)
		if err != nil {
			ctx.Logger().Warn(err)
		}
		endedAt, err := ParseTimeOrNil(query.Node.ProjectV2.Items.Nodes[itemKey].EndedAt.ProjectV2ItemFieldTextValue.Text)
		if err != nil {
			ctx.Logger().Warn(err)
		}
		incident := api.Incident{
			Affects:    []string{},
			Id:         query.Node.ProjectV2.Items.Nodes[itemKey].Content.Issue.Id,
			Title:      query.Node.ProjectV2.Items.Nodes[itemKey].Content.Issue.Title,
			ImpactType: query.Node.ProjectV2.Items.Nodes[itemKey].ImpactType.ProjectV2ItemFieldSingleSelectValue.Name,
			Phase:      query.Node.ProjectV2.Items.Nodes[itemKey].Phase.ProjectV2ItemFieldSingleSelectValue.Name,
			BeganAt:    beganAt,
			EndedAt:    endedAt,
		}
		for componentKey := range query.Node.ProjectV2.Items.Nodes[itemKey].Labels.ProjectV2ItemFieldLabelValue.Labels.Nodes {
			incident.Affects = append(
				incident.Affects,
				query.Node.ProjectV2.Items.Nodes[itemKey].Labels.ProjectV2ItemFieldLabelValue.Labels.Nodes[componentKey].Id,
			)
		}
		incidents = append(incidents, incident)
	}
	return ctx.JSON(200, incidents)
}
func (s *ServerImplementation) GetIncident(ctx echo.Context, incidentId string) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetPhases(ctx echo.Context) error {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Field struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"field(name: \"Status\")"`
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
	phases := []api.IncidentPhase{}
	for _, phase := range query.Node.ProjectV2.Field.ProjectV2SingleSelectField.Options {
		phases = append(phases, phase.Name)
	}
	return ctx.JSON(200, phases)
}
