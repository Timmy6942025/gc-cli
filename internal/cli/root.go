package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/tui"
)

func Execute() int {
	app, err := NewApp(os.Stdout, os.Stderr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to initialize app:", err)
		return 1
	}
	defer app.Close()

	cmd := NewRootCmd(app)
	if err := cmd.Execute(); err != nil {
		_ = app.Printer.PrintError(err)
		return 1
	}
	return 0
}

func NewRootCmd(app *App) *cobra.Command {
	var jsonOutput bool
	useName := filepath.Base(os.Args[0])
	if useName == "" {
		useName = "gc-cli"
	}
	cmd := &cobra.Command{
		Use:   useName,
		Short: "Google Classroom CLI/TUI",
		Long:  "A Google Classroom terminal client with API-first parity and seamless web handoff.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			client, err := app.ClassroomClient(ctx, auth.DefaultReadScopes)
			if err != nil {
				return err
			}
			return tui.Run(client, app.Resolver)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			app.SetJSONOutput(jsonOutput)
		},
		SilenceUsage: true,
	}
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output machine-readable JSON")

	cmd.AddCommand(
		newAuthCmd(app),
		newClassesCmd(app),
		newStreamCmd(app),
		newClassworkCmd(app),
		newSubmissionsCmd(app),
		newPeopleCmd(app),
		newGradesCmd(app),
		newTopicsCmd(app),
		newTodoCmd(app),
		newCalendarCmd(app),
		newHandoffCmd(app),
	)
	return cmd
}
