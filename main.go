package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/joshmue/scs-status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type ServerImplementation struct {
	GithubClient   *github.Client
	GithubV4Client *githubv4.Client
	ProjectOwner   string
	ProjectNumber  int64
	ProjectID      string
	ImpactTypes    []string
	LastPhase      string
}

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

func main() {
	addr := flag.String("addr", ":3000", "address to listen on")
	projectOwner := flag.String("github.project.user", "joshmue", "user owning the project")
	projectNumber := flag.Int64("github.project.number", 1, "project number")
	impactTypeList := flag.String("impacttypes", "performance-degration,connectivity-issues", `","-seperated list of impact types`)
	lastPhase := flag.String("last-phase", "Done", "last phase of incidents")
	flag.Parse()

	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))
	server := &ServerImplementation{
		GithubClient:   github.NewClient(httpClient),
		GithubV4Client: githubv4.NewClient(httpClient),
		ProjectOwner:   *projectOwner,
		ProjectNumber:  *projectNumber,
		ImpactTypes:    strings.Split(*impactTypeList, ","),
		LastPhase:      *lastPhase,
	}
	if err := server.fillProjectID(); err != nil {
		log.Fatalln(err)
	}
	if err := server.ensureProjectConfiguration(); err != nil {
		log.Fatalln(err)
	}
	e := echo.New()
	e.Use(middleware.Logger())
	api.RegisterHandlers(e, server)
	log.Fatalln(e.Start(*addr))
}
