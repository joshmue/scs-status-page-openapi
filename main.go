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
	"golang.org/x/oauth2"
)

type ServerImplementation struct {
	GithubClient *github.Client
	RepoOwner    string
	RepoName     string
}

func (s *ServerImplementation) GetComponents(ctx echo.Context) error {
	labels, _, err := s.GithubClient.Issues.ListLabels(
		ctx.Request().Context(),
		s.RepoOwner, s.RepoName, nil,
	)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	components := []*api.Component{}
	for _, label := range labels {
		if !strings.HasPrefix(label.GetName(), "component:") {
			continue
		}
		componentId := fmt.Sprintf("%d", label.GetID())
		componentDisplayName := strings.TrimPrefix(label.GetName(), "component:")
		newComponent := &api.Component{
			Id:          &componentId,
			DisplayName: &componentDisplayName,
		}
		components = append(components, newComponent)
	}
	return ctx.JSON(200, components)
}
func (s *ServerImplementation) GetImpacttypes(ctx echo.Context) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	return fmt.Errorf("not implemented")
}
func (s *ServerImplementation) GetPhases(ctx echo.Context) error {
	return fmt.Errorf("not implemented")
}

func main() {
	addr := flag.String("addr", ":3000", "address to listen on")
	repoOwner := flag.String("repo.owner", "joshmue", "Owner of Github repository")
	repoName := flag.String("repo.name", "statuspage-issues-playground", "Name of Github repository")
	flag.Parse()

	githubClient := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)))
	var server api.ServerInterface = &ServerImplementation{
		GithubClient: githubClient,
		RepoOwner:    *repoOwner,
		RepoName:     *repoName,
	}
	e := echo.New()
	e.Use(middleware.Logger())
	api.RegisterHandlers(e, server)
	log.Fatalln(e.Start(*addr))
}
