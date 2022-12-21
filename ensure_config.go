package main

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

func (s *ServerImplementation) ensureProjectConfiguration() error {
	// Check final "Status" field
	var query struct {
		User struct {
			ProjectV2 struct {
				StatusField struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"status: field(name: \"Status\")"`
				BeganAtField struct {
					ProjectV2Field struct {
						DataType string
					} `graphql:"... on ProjectV2Field"`
				} `graphql:"beganat: field(name: \"Began At\")"`
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
	// Check "Status" field
	options := query.User.ProjectV2.StatusField.ProjectV2SingleSelectField.Options
	if len(options) == 0 {
		return fmt.Errorf(`expected to have phases encoded as fields of "Status"; not having any`)
	}
	if options[len(options)-1].Name != s.LastPhase {
		return fmt.Errorf(`expected final phase to be "%s"; is "%s"`, s.LastPhase, options[len(options)-1].Name)
	}
	// Check "Began At" field
	if query.User.ProjectV2.BeganAtField.ProjectV2Field.DataType != "TEXT" {
		return fmt.Errorf(`expected field "Began At" to be "TEXT"; is "%s"`, query.User.ProjectV2.BeganAtField.ProjectV2Field.DataType)
	}
	return nil
}
