package main

import (
	"fmt"
	"os"

	"github.com/centrifuge/pod/utils"
	"github.com/common-nighthawk/go-figure"
	logging "github.com/ipfs/go-log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	gologging "github.com/whyrusleeping/go-logging"
)

var cfgFile string
var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "centrifuge",
	Short: "Centrifuge protocol node",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

var log = logging.Logger("centrifuge-cmd")

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	printStartMessage()
	logging.SetAllLoggers(logging.LevelInfo)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func printStartMessage() {
	c := figure.NewFigure("Centrifuge  OS", "banner", true)
	fmt.Println()
	c.Print()
	fmt.Println()
	fmt.Println("Centrifuge OS and this client implementation are beta software. For more information refer to the disclaimer on https://developer.centrifuge.io.")
}

func init() {
	cobra.OnInitialize(setCentrifugeLoggers)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.centrifuge.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set loglevel to debug")
}

// ensureConfigFile ensures a config file is provided
func ensureConfigFile() string {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		cfgFile = fmt.Sprintf("%s/%s", home, ".centrifuge.yaml")
		if _, err := os.Stat(cfgFile); err != nil {
			log.Info("Config file not provided and default $HOME/.centrifuge.yaml does not exist")
			cfgFile = ""
		}
	}
	return cfgFile
}

// setCentrifugeLoggers sets the loggers.
func setCentrifugeLoggers() {
	var formatter = gologging.MustStringFormatter(utils.GetCentLogFormat())
	gologging.SetFormatter(formatter)
	if verbose {
		logging.SetAllLoggers(logging.LevelDebug)
		return
	}

	logging.SetAllLoggers(logging.LevelInfo)
}
