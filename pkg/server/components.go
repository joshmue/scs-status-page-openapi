package server

import (
	"context"
	"strings"

	"github.com/joshmue/scs-status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/shurcooL/githubv4"
)

func (s *ServerImplementation) GetComponent(ctx echo.Context, componentId string) error {
	var query struct {
		Node struct {
			Label struct {
				Id     string
				Name   string
				Issues struct {
					Nodes []struct {
						Id string
					}
				} `graphql:"issues(first:10)"`
			} `graphql:"... on Label"`
		} `graphql:"node(id: $labelid)"`
	}
	err := s.GithubV4Client.Query(
		context.Background(),
		&query,
		map[string]interface{}{
			"labelid": githubv4.ID(componentId),
		},
	)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	affectedBy := []api.Id{}
	for i := range query.Node.Label.Issues.Nodes {
		affectedBy = append(affectedBy, query.Node.Label.Issues.Nodes[i].Id)
	}
	component := api.Component{
		AffectedBy:  affectedBy,
		DisplayName: strings.TrimPrefix(query.Node.Label.Name, "component:"),
		Id:          componentId,
		Labels:      map[string]string{}, // TODO
	}
	return ctx.JSON(200, component)
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
					Id:          query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Id,
					DisplayName: strings.TrimPrefix(query.Node.ProjectV2.Repositories.Nodes[repo].Labels.Nodes[label].Name, "component:"),
				}
				components = append(components, component)
			}
		}
	}
	return ctx.JSON(200, components)
}
