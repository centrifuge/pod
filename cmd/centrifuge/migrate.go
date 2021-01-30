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
			err := doMigrate()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(migrateCmd)
}

func doMigrate() error {
	cfg := config.LoadConfiguration(cfgFile)
	runner := migration.NewMigrationRunner()
	return runner.RunMigrations(cfg.GetStoragePath())
}
