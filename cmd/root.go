package cmd

import (
	"fmt"
	"os"

	"quibit/internal/db"

	"github.com/spf13/cobra"
)

var migrate bool

var rootCmd = &cobra.Command{
	Use:           "quibit",
	Short:         "Quibit is a personal CLI to generate portfolio project ideas.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if migrate {
			{
				ctx := cmd.Context()
				fmt.Fprintln(cmd.OutOrStdout(), "Running database migrations...")
				gdb, err := db.Connect(ctx)
				if err != nil {
					return fmt.Errorf("migrate: %w", err)
				}

				sqlDB, err := gdb.DB()
				if err != nil {
					return fmt.Errorf("migrate: get sql db: %w", err)
				}

				if err := db.AutoMigrate(ctx, gdb); err != nil {
					return fmt.Errorf("migrate: %w", err)
				}
				if err := sqlDB.Close(); err != nil {
					return fmt.Errorf("migrate: close sql db: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Database migrations completed.")
			}

			os.Exit(0)
			return nil
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&migrate, "migrate", false, "Run database migrations")
	rootCmd.AddCommand(generateCmd)
}
