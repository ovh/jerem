package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ovh/jerem/src/core"
	"github.com/ovh/jerem/src/runner"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yml)")
	RootCmd.PersistentFlags().Int32("log-level", 4, "set logging level between 0 and 5")
}

func initConfig() {
	// Environment variables management
	viper.SetEnvPrefix("jerem")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default
	viper.SetDefault("runner.period", 10*60*time.Second)
	viper.SetDefault("api.listen", "127.0.0.1:8080")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath("./")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	level := viper.GetInt("log-level")
	log.SetLevel(log.InfoLevel) // level == 4
	if level >= 0 && level < len(log.AllLevels) {
		log.SetLevel(log.AllLevels[level])
	}
	log.SetLevel(log.AllLevels[5])
}

// RootCmd main command of jerem
var RootCmd = &cobra.Command{
	Use:   "jerem",
	Short: "Jerem bring observability to Jira",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := core.LoadConfig()
		if err != nil {
			log.WithError(err).Fatal("Fail to load config")
		}

		// Jerem status handler
		go func() {
			e := echo.New()
			e.HideBanner = true
			address := viper.GetString("api.listen")
			e.GET("/health", func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			err := e.Start(address)
			if err != nil && err != http.ErrServerClosed {
				log.WithError(err).Error("cannot start HTTP server")
			}
		}()

		// Start Jerem JIRA epic and sprint collectors
		epicRunner := core.NewRunner(func() {
			runner.EpicRunner(config)
		}, viper.GetDuration("runner.period")+1*time.Second)

		sprintRunner := core.NewRunner(func() {
			runner.SprintRunner(config)
		}, viper.GetDuration("runner.period"))

		var gracefulStop = make(chan os.Signal, 1)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, os.Interrupt)
		<-gracefulStop

		epicRunner.Stop()
		sprintRunner.Stop()
	},
}
