package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
)

func newTodoCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "to-do", Short: "View To-do work"}
	cmd.AddCommand(newTodoListCmd(app))
	return cmd
}

func newTodoListCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List student To-do items",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkMeReadonly, auth.ScopeCoursesReadonly})
			if err != nil {
				return err
			}
			items, err := client.BuildTodo(ctx)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"to_do": items})
			}
			if len(items) == 0 {
				_, _ = fmt.Fprintln(app.Out, "No To-do items.")
				return nil
			}
			for _, item := range items {
				due := "no due date"
				if !item.DueDate.IsZero() {
					due = item.DueDate.Format("2006-01-02 15:04")
				}
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\t%s\tdue=%s\tlate=%t\n", item.CourseName, item.CourseWorkTitle, item.SubmissionState, due, item.Late)
			}
			return nil
		},
	}
	return cmd
}
