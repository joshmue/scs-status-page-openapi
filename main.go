package main

import (
	"context"
	"flag"
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
	GithubClient      *github.Client
	GithubV4Client    *githubv4.Client
	ProjectOwner      string
	ProjectOwnerIsOrg bool
	ProjectNumber     int64
	ProjectID         string
	ImpactTypes       []string
	LastPhase         string
}

func main() {
	addr := flag.String("addr", ":3000", "address to listen on")
	projectOwner := flag.String("github.project.owner", "joshmue", "user owning the project")
	projectOwnerIsOrg := flag.Bool("github.project.owner.is-org", false, "sets whether the owner of the github project is an org instead of an user")
	projectNumber := flag.Int64("github.project.number", 1, "project number")
	impactTypeList := flag.String("impacttypes", "performance-degration,connectivity-issues", `","-seperated list of impact types`)
	lastPhase := flag.String("last-phase", "Done", "last phase of incidents")
	flag.Parse()

	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))
	server := &ServerImplementation{
		GithubClient:      github.NewClient(httpClient),
		GithubV4Client:    githubv4.NewClient(httpClient),
		ProjectOwner:      *projectOwner,
		ProjectOwnerIsOrg: *projectOwnerIsOrg,
		ProjectNumber:     *projectNumber,
		ImpactTypes:       strings.Split(*impactTypeList, ","),
		LastPhase:         *lastPhase,
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
