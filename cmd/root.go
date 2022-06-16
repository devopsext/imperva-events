package cmd

import (
	"github.com/devopsext/imperva-events/pkg/common"
	"github.com/devopsext/imperva-events/pkg/imperva"
	"github.com/devopsext/imperva-events/pkg/output"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"sync"
	"time"
)

var (
	impervaID        string
	impervaToken     string
	impervaAccountID string

	isDebug = false

	slackToken   string
	slackChannel string

	grafanaURL    string
	grafanaAPIKey string

	pollInterval = 10
	initInterval = 600

	wg sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "imperva-events",
	Short: "Scrap imperva security events",
	Run: func(cmd *cobra.Command, args []string) {

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.TimeFieldFormat = time.RFC3339
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})

		if isDebug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		if impervaID == "" || impervaToken == "" || impervaAccountID == "" {
			log.Fatal().Msg("IMPERVA_API_ID, IMPERVA_API_TOKEN and IMPERVA_ACCOUNT_ID must be set")
		}

		i, err := imperva.NewImperva(impervaID, impervaToken, impervaAccountID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Imperva client")
		}

		if slackToken != "" && slackChannel != "" {
			i.AddOutput(output.NewSlack(slackToken, slackChannel))
			log.Info().Msg("Slack output enabled")
		}

		if grafanaURL != "" && grafanaAPIKey != "" {
			i.AddOutput(output.NewGrafana(grafanaURL, grafanaAPIKey))
			log.Info().Msg("Grafana output enabled")
		}

		wg.Add(1)
		i.Run(pollInterval, initInterval, &wg)
		wg.Wait()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	impervaID = common.EnvGet("IMPERVA_API_ID", "").(string)
	impervaToken = common.EnvGet("IMPERVA_API_TOKEN", "").(string)
	impervaAccountID = common.EnvGet("IMPERVA_ACCOUNT_ID", "").(string)

	slackToken = common.EnvGet("IMPERVA_SLACK_TOKEN", "").(string)
	slackChannel = common.EnvGet("IMPERVA_SLACK_CHANNEL", "").(string)

	grafanaURL = common.EnvGet("IMPERVA_GRAFANA_URL", "").(string)
	grafanaAPIKey = common.EnvGet("IMPERVA_GRAFANA_API_KEY", "").(string)

	pollInterval = common.EnvGet("IMPERVA_POLL_INTERVAL", pollInterval).(int)
	initInterval = common.EnvGet("IMPERVA_INIT_INTERVAL", initInterval).(int)

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&impervaID, "api-id", impervaID, "Imperva API ID")
	flags.StringVar(&impervaToken, "api-token", impervaToken, "Imperva API Token")
	flags.StringVar(&impervaAccountID, "account-id", impervaAccountID, "Imperva Account ID")

	flags.BoolVar(&isDebug, "debug", isDebug, "Enable debug logging")

	flags.StringVar(&slackToken, "slack-token", slackToken, "Slack token")
	flags.StringVar(&slackChannel, "slack-channel", slackChannel, "Slack channel")

	flags.StringVar(&grafanaURL, "grafana-url", grafanaURL, "Grafana URL")
	flags.StringVar(&grafanaAPIKey, "grafana-api-key", grafanaAPIKey, "Grafana API Key")

	flags.IntVar(&pollInterval, "poll-interval", pollInterval, "Imperva Poll interval (seconds)")
	flags.IntVar(&initInterval, "init-interval", initInterval, "Imperva Init interval (minutes)")
}
