package cmd

import (
	"fmt"
	"os"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	logging "github.com/ipfs/go-log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	gologging "github.com/whyrusleeping/go-logging"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)

//global flags
var cfgFile string
var verbose bool

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	logging.SetAllLoggers(gologging.INFO)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCentrifuge)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.centrifuge.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set loglevel to debug")
}

// initCentrifuge reads in config file and ENV variables if set.
func initCentrifuge() {
	if verbose {
		logging.SetAllLoggers(gologging.DEBUG)
	} else {
		logging.SetAllLoggers(gologging.INFO)
	}
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
	// If a config file is found, read it in.
	config.Bootstrap(cfgFile)
}

func defaultBootstrap() {
	mb := cc.MainBootstrapper{}
	mb.PopulateDefaultBootstrappers()
	err := mb.Bootstrap(map[string]interface{}{})
	if err != nil {
		// application must not continue to run
		panic(err)
	}
}
