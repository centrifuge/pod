package cmd

import (
	"fmt"
	"os"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	logging "github.com/ipfs/go-log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	gologging "github.com/whyrusleeping/go-logging"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "centrifuge",
	Short: "Centrifuge protocol node",
	Long:  `POC for centrifuge app`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

var log = logging.Logger("centrifuge-cmd")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = gologging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	logging.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.centrifuge.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		cfgFile = fmt.Sprintf("%s/%s", home, ".centrifuge.yaml")
		if _, err := os.Stat(cfgFile); err != nil {
			log.Error("Config file not provided and default $HOME/.centrifuge.yaml does not exist")
			os.Exit(1)
		}
	}
	// If a config file is found, read it in.
	config.Bootstrap(cfgFile)
}
