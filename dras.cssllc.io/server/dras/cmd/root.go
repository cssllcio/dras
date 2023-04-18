package cmd

import (
	"dras.cssllc.io/server/dras/api"
	dbconn "dras.cssllc.io/server/dras/dbcon"
	"fmt"
	"github.com/spf13/cobra"
)

var dialect, hostname, dbName, dbUser, dbPassword string
var serverPort, dbPort int

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "A REST API server with CRUD operations for a Postgres database",
	Long: `A REST API server with CRUD operations for a Postgres database using
the Gin web framework and GORM.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := dbconn.InitDB(dialect, hostname, dbName, dbPort, dbUser, dbPassword)
		if err != nil {
			return err
		}

		r := api.SetupRouter(db)

		err = r.Run(fmt.Sprintf(":%d", serverPort))
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dialect, "dialect", "l", "postgres", "DB dialect - postgres, mssql (default: postgres)")
	rootCmd.PersistentFlags().StringVarP(&hostname, "hostname", "g", "", "host name (required)")
	rootCmd.PersistentFlags().StringVarP(&dbName, "db", "d", "", "Database name (required)")
	rootCmd.PersistentFlags().IntVarP(&dbPort, "dbport", "b", 5432, "Database port (default: 5432)")
	rootCmd.PersistentFlags().IntVarP(&serverPort, "port", "p", 8080, "Server port number (default: 8080)")
	rootCmd.PersistentFlags().StringVarP(&dbUser, "user", "u", "", "Database user (required)")
	rootCmd.PersistentFlags().StringVarP(&dbPassword, "password", "w", "", "Database password (required)")

	_ = rootCmd.MarkPersistentFlagRequired("hostname")
	_ = rootCmd.MarkPersistentFlagRequired("db")
	_ = rootCmd.MarkPersistentFlagRequired("user")
	_ = rootCmd.MarkPersistentFlagRequired("password")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
