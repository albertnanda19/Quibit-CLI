package cmd

import (
	"fmt"
	"os"
	"sync"

	"quibit/internal/db"
	"quibit/internal/persistence"
	"quibit/internal/tui"

	"github.com/spf13/cobra"
)

var migrate bool
var noAnim bool
var splashOnce sync.Once

var rootCmd = &cobra.Command{
	Use:           "quibit",
	Short:         "Quibit is a personal CLI to generate portfolio project ideas.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Global UI motion toggle (spinner/micro-delays). No business logic impact.
		tui.SetMotionEnabled(!noAnim)
		if !migrate {
			splashOnce.Do(func() {
				_ = tui.ShowSplashScreen(os.Stdin, cmd.OutOrStdout())
			})
		}
		if migrate {
			{
				ctx := cmd.Context()
				out := cmd.OutOrStdout()
				spin := tui.StartSpinner(ctx, out, "Running database migrations")
				defer spin.Stop()
				gdb, err := db.Connect(ctx)
				if err != nil {
					return fmt.Errorf("migrate: %w", err)
				}

				sqlDB, err := gdb.DB()
				if err != nil {
					return fmt.Errorf("migrate: get sql db: %w", err)
				}

				if err := persistence.AutoMigrate(ctx, gdb); err != nil {
					return fmt.Errorf("migrate: %w", err)
				}
				if err := sqlDB.Close(); err != nil {
					return fmt.Errorf("migrate: close sql db: %w", err)
				}
				spin.Stop()
				tui.Done(out, "Database migrations completed")
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
	rootCmd.PersistentFlags().BoolVar(&noAnim, "no-anim", false, "Disable subtle CLI animations")
	rootCmd.AddCommand(generateCmd)
}
