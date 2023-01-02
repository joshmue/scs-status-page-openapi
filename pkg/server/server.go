package server

import (
	"context"
	"fmt"
	"strings"

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

func (s *ServerImplementation) EnsureProjectConfiguration() error {
	// Make a single query to assess all relevant factors
	var query struct {
		Node struct {
			ProjectV2 struct {
				Repositories struct {
					Nodes []struct {
						Labels struct {
							Nodes []struct {
								Name string
							}
						} `graphql:"labels(first: 10)"`
					}
				} `graphql:"repositories(first: 10)"`
				StatusField struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"status: field(name: \"Status\")"`
				ImpactTypeField struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"impacttype: field(name: \"Impact Type\")"`
				BeganAtField struct {
					ProjectV2Field struct {
						DataType string
					} `graphql:"... on ProjectV2Field"`
				} `graphql:"beganat: field(name: \"Began At\")"`
				EndedAtField struct {
					ProjectV2Field struct {
						DataType string
					} `graphql:"... on ProjectV2Field"`
				} `graphql:"endedat: field(name: \"Ended At\")"`
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
		return err
	}
	// Check components
	componentFound := false
	for _, repo := range query.Node.ProjectV2.Repositories.Nodes {
		for _, label := range repo.Labels.Nodes {
			if strings.HasPrefix(label.Name, "component:") {
				componentFound = true
			}
		}
	}
	if !componentFound {
		return fmt.Errorf("expected components, got none")
	}
	// Check "Status" field
	phaseOptions := query.Node.ProjectV2.StatusField.ProjectV2SingleSelectField.Options
	if len(phaseOptions) == 0 {
		return fmt.Errorf(`expected to have phases encoded as fields of "Status"; not having any`)
	}
	if phaseOptions[len(phaseOptions)-1].Name != s.LastPhase {
		return fmt.Errorf(`expected final phase to be "%s"; is "%s"`, s.LastPhase, phaseOptions[len(phaseOptions)-1].Name)
	}
	// Check "Impact Type" field
	impactTypeOptions := query.Node.ProjectV2.ImpactTypeField.ProjectV2SingleSelectField.Options
	if len(impactTypeOptions) == 0 {
		return fmt.Errorf(`expected to have impact types encoded as fields of "Impact Type"; not having any`)
	}
	// Check "Began At" field
	if query.Node.ProjectV2.BeganAtField.ProjectV2Field.DataType != "TEXT" {
		return fmt.Errorf(`expected field "Began At" to be "TEXT"; is "%s"`, query.Node.ProjectV2.BeganAtField.ProjectV2Field.DataType)
	}
	// Check "Ended At" field
	if query.Node.ProjectV2.EndedAtField.ProjectV2Field.DataType != "TEXT" {
		return fmt.Errorf(`expected field "Began At" to be "TEXT"; is "%s"`, query.Node.ProjectV2.EndedAtField.ProjectV2Field.DataType)
	}
	return nil
}