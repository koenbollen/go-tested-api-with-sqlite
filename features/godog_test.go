package features_test

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/koenbollen/go-tested-api-with-sqlite/features/steps"
	"github.com/koenbollen/go-tested-api-with-sqlite/internal"
	"github.com/koenbollen/go-tested-api-with-sqlite/internal/routes"
	"github.com/koenbollen/go-tested-api-with-sqlite/internal/util/timeutil"
)

var AllRoutes = []internal.Route{
	routes.Redirections,
}

type stepCollection interface {
	InitializeSuite(suite *godog.TestSuiteContext) error
	InitializeScenario(scenario *godog.ScenarioContext) error
}

func TestMain(m *testing.M) {
	if _, found := os.LookupEnv("SKIP_FEATURE_TESTS"); found {
		return
	}
	os.Setenv("ENV", "test")

	opts := godog.Options{
		Paths:     []string{"."},
		Tags:      "~@wip && ~@todo",
		Randomize: -1,
	}
	godog.BindCommandLineFlags("godog.", &opts)
	flag.Parse()
	if len(flag.Args()) > 0 {
		opts.Paths = flag.Args()
	}

	httpSteps := &steps.HTTPSteps{}
	databaseSteps := &steps.DatabaseSteps{}
	stepCollections := []stepCollection{
		httpSteps,
		databaseSteps,
	}

	var mux *http.ServeMux

	suite := godog.TestSuite{
		Name:    "go-tested-api-with-sqlite",
		Options: &opts,
		TestSuiteInitializer: func(suite *godog.TestSuiteContext) {
			for _, step := range stepCollections {
				if err := step.InitializeSuite(suite); err != nil {
					panic(err)
				}
			}
			httpSteps.ApplicationMux = func() http.Handler {
				return mux
			}
		},
		ScenarioInitializer: func(scenario *godog.ScenarioContext) {
			for _, step := range stepCollections {
				if err := step.InitializeScenario(scenario); err != nil {
					panic(err)
				}
			}
			scenario.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				t, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
				ctx = timeutil.WithTime(ctx, t)

				deps, err := internal.Setup(ctx, internal.DefaultConfig())
				if err != nil {
					panic(err)
				}
				mux, err = internal.SetupRoutes(ctx, deps, AllRoutes...)
				if err != nil {
					panic(err)
				}
				databaseSteps.DB = deps.DB
				return ctx, nil
			})
		},
	}

	status := suite.Run()
	os.Exit(status)
}
