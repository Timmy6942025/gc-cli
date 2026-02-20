package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/output"
	"github.com/timothy/gc-cli/internal/platform"
)

func newHandoffCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "handoff",
		Aliases: []string{"web"},
		Short:   "Open Classroom web handoff links",
		Example: "  gc handoff open invite-code -c <course_id>\n  gc web open stream-moderation -c <course_id>",
	}
	cmd.AddCommand(newHandoffOpenCmd(app))
	return cmd
}

func newHandoffOpenCmd(app *App) *cobra.Command {
	var courseID string
	var customURL string
	var openBrowser bool
	cmd := &cobra.Command{
		Use:     "open <feature> [course_id]",
		Aliases: []string{"go", "launch"},
		Short:   "Resolve and open a web handoff feature",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			feature := args[0]
			if len(args) > 1 {
				fillPositional(args[1:], &courseID)
			}
			handoff := app.Resolver.Resolve(feature, courseID, map[string]string{"url": customURL})
			if handoff.Blocked {
				return output.CLIError{Code: "handoff_blocked", Reason: handoff.Reason, ActionableHint: "Use a supported feature and provide required IDs"}
			}
			if openBrowser {
				if err := platform.OpenURL(handoff.URL); err != nil {
					return output.CLIError{Code: "handoff_open_failed", Reason: err.Error(), ActionableHint: "Open the URL manually", WebHandoffURL: handoff.URL}
				}
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(handoff)
			}
			if openBrowser {
				_, _ = fmt.Fprintf(app.Out, "Opened: %s\n", handoff.URL)
			} else {
				_, _ = fmt.Fprintf(app.Out, "%s\n", handoff.URL)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVar(&customURL, "url", "", "Custom fallback URL for unknown features")
	cmd.Flags().BoolVar(&openBrowser, "open-browser", true, "Open the handoff in browser")
	return cmd
}
