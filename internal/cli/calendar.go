package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/platform"
)

func newCalendarCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calendar",
		Aliases: []string{"cal"},
		Short:   "Calendar shortcuts",
		Example: "  gc calendar open",
	}
	cmd.AddCommand(newCalendarOpenCmd(app))
	return cmd
}

func newCalendarOpenCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open",
		Aliases: []string{"launch"},
		Short:   "Open Google Calendar",
		RunE: func(cmd *cobra.Command, args []string) error {
			h := app.Resolver.Resolve("calendar", "", nil)
			if h.Blocked {
				return errors.New(h.Reason)
			}
			if err := platform.OpenURL(h.URL); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(h)
			}
			_, _ = fmt.Fprintln(app.Out, "Opened Google Calendar")
			return nil
		},
	}
	return cmd
}
