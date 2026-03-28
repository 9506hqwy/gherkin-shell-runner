package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/cobra"

	"github.com/9506hqwy/gherkin-shell-runner/pkg/testing"
)

var version = "<version>"
var commit = "<commit>"

var rootCmd = &cobra.Command{
	Use:     "gherkin-shell-runner",
	Short:   "Gherkin Shell Runner",
	Long:    "Gherkin Shell Runner",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		showStepDefinitions, _ := cmd.Flags().GetBool("show-steps")
		randomize, _ := cmd.Flags().GetInt64("random")
		stopOnFailure, _ := cmd.Flags().GetBool("stop-on-failture")
		noColors, _ := cmd.Flags().GetBool("no-colors")
		tags, _ := cmd.Flags().GetString("tags")
		format, _ := cmd.Flags().GetString("format")
		concurrency, _ := cmd.Flags().GetInt("concurrency")

		opts := godog.Options{
			ShowStepDefinitions: showStepDefinitions,
			Randomize:           randomize,
			StopOnFailure:       stopOnFailure,
			NoColors:            noColors,
			Tags:                tags,
			Format:              format,
			Concurrency:         concurrency,
			Paths:               args,
			Output:              colors.Colored(os.Stdout),
		}

		suite := godog.TestSuite{
			Name:                 cmd.Name(),
			TestSuiteInitializer: testing.InitializeTestSuite,
			ScenarioInitializer:  testing.InitializeScenario,
			Options:              &opts,
		}

		status := suite.Run()

		//revive:disable:deep-exit
		os.Exit(status)
		//revive:enable:deep-exit
	},
}

//revive:disable:add-constant
//revive:disable:line-length-limit

func init() {
	rootCmd.Flags().Int("concurrency", 1, "Run scenario concurrency.")
	rootCmd.Flags().String("format", "progress", "Report format.")
	rootCmd.Flags().Bool("no-colors", false, "Disable ansi color.")
	rootCmd.Flags().Int64("random", -1, "Randamize scenario order.")
	rootCmd.Flags().Bool("stop-on-failture", false, "Stop on first failed scenario.")
	rootCmd.Flags().String("tags", "~@ignore", "Filter scenario.")
	rootCmd.Flags().Bool("show-steps", false, "Show avaiblae step definitions.")
}

//revive:enable:line-length-limit
//revive:enable:add-constant

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
