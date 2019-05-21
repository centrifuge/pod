package main

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/migration"
	"github.com/spf13/cobra"
)

func init() {

	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Runs node migrations",
		Long:  ``,
		Run: func(c *cobra.Command, args []string) {
			cfg := config.LoadConfiguration(cfgFile)
			runner := migration.NewMigrationRunner()
			err := runner.RunMigrations(cfg.GetStoragePath())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(migrateCmd)
}


