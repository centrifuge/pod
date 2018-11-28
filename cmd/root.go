package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	c "github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	logging "github.com/ipfs/go-log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	gologging "github.com/whyrusleeping/go-logging"
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

//setCentrifugeLoggers sets the loggers.
func setCentrifugeLoggers() {

	var formatter = gologging.MustStringFormatter(utils.GetCentLogFormat())
	gologging.SetFormatter(formatter)
	if verbose {
		logging.SetAllLoggers(gologging.DEBUG)
		return
	}

	logging.SetAllLoggers(gologging.INFO)

}

func runBootstrap(cfgFile string) {
	mb := c.MainBootstrapper{}
	mb.PopulateRunBootstrappers()
	ctx := map[string]interface{}{}
	ctx[config.BootstrappedConfigFile] = cfgFile
	err := mb.Bootstrap(ctx)
	if err != nil {
		// application must not continue to run
		panic(err)
	}
}

func baseBootstrap(cfgFile string) map[string]interface{} {
	mb := c.MainBootstrapper{}
	mb.PopulateBaseBootstrappers()
	ctx := map[string]interface{}{}
	ctx[config.BootstrappedConfigFile] = cfgFile
	err := mb.Bootstrap(ctx)
	if err != nil {
		// application must not continue to run
		panic(err)
	}
	return ctx
}

func commandBootstrap(cfgFile string) (map[string]interface{}, context.CancelFunc, error) {
	ctx := baseBootstrap(cfgFile)
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	// init node with only the queue server which is needed by commands
	n := node.New([]node.Server{queueSrv})
	cx, canc := context.WithCancel(context.Background())
	e := make(chan error)
	go n.Start(cx, e)
	return ctx, canc, nil
}
