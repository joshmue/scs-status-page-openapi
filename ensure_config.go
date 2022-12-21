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
				Field struct {
					ProjectV2SingleSelectField struct {
						Options []struct {
							Name string
						}
					} `graphql:"... on ProjectV2SingleSelectField"`
				} `graphql:"field(name: $status)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $user)"`
	}
	err := s.GithubV4Client.Query(
		context.Background(),
		&query,
		map[string]interface{}{
			"user":   githubv4.String(s.ProjectOwner),
			"number": githubv4.Int(s.ProjectNumber),
			"status": githubv4.String("Status"),
		},
	)
	if err != nil {
		return err
	}
	options := query.User.ProjectV2.Field.ProjectV2SingleSelectField.Options
	if options[len(options)-1].Name != s.LastPhase {
		return fmt.Errorf("expected final phase to be %s; is %s", s.LastPhase, options[len(options)-1].Name)
	}

	// Check "Began At"
	var queryBeganAt struct {
		User struct {
			ProjectV2 struct {
				Field struct {
					ProjectV2Field struct {
						DataType string
					} `graphql:"... on ProjectV2Field"`
				} `graphql:"field(name: $beganat)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $user)"`
	}
	err = s.GithubV4Client.Query(
		context.Background(),
		&queryBeganAt,
		map[string]interface{}{
			"user":    githubv4.String(s.ProjectOwner),
			"number":  githubv4.Int(s.ProjectNumber),
			"beganat": githubv4.String("Began At"),
		},
	)
	if err != nil {
		return err
	}
	beganAtDataType := queryBeganAt.User.ProjectV2.Field.ProjectV2Field.DataType
	if beganAtDataType != "TEXT" {
		return fmt.Errorf(`expected "Began At" data type to be TEXT; is %s`, beganAtDataType)
	}
	return nil
}
