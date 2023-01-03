package server

import (
	"context"
	"fmt"
	"time"

	"github.com/joshmue/scs-status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/shurcooL/githubv4"
)

func ParseTimeOrNil(timeString string) (*time.Time, error) {
	beganAt, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return nil, err
	}
	return &beganAt, nil
}

func (s *ServerImplementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	type projectItem struct {
		Id string
		Content struct {
			Issue struct {
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
			Id:         query.Node.ProjectV2.Items.Nodes[itemKey].Id,
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
