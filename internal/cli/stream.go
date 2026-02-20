package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newStreamCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stream",
		Aliases: []string{"announcements", "ann"},
		Short:   "Manage Stream announcements",
		Example: "  gc stream list -c <course_id>\n  gc stream post -c <course_id> --text \"Reminder: quiz Friday\"",
	}
	cmd.AddCommand(
		newStreamListCmd(app),
		newStreamPostCmd(app),
		newStreamEditCmd(app),
		newStreamDeleteCmd(app),
	)
	return cmd
}

func newStreamListCmd(app *App) *cobra.Command {
	var courseID string
	var pageSize int64
	var pageToken string
	cmd := &cobra.Command{
		Use:     "list [course_id]",
		Aliases: []string{"ls"},
		Short:   "List announcements in a class stream",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeAnnouncementsReadonly})
			if err != nil {
				return err
			}
			items, next, err := client.ListAnnouncements(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"announcements": items, "next_page_token": next})
			}
			for _, ann := range items {
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\n", ann.Id, truncate(ann.Text, 90))
			}
			if next != "" {
				_, _ = fmt.Fprintf(app.Out, "next_page_token=%s\n", next)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().Int64VarP(&pageSize, "page-size", "n", 50, "Number of announcements to return")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Pagination token")
	return cmd
}

func newStreamPostCmd(app *App) *cobra.Command {
	var courseID string
	var text string
	var state string
	cmd := &cobra.Command{
		Use:     "post [course_id]",
		Aliases: []string{"add", "create"},
		Short:   "Post an announcement",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeAnnouncements})
			if err != nil {
				return err
			}
			ann := &gclassroom.Announcement{Text: text, State: state}
			created, err := client.CreateAnnouncement(ctx, courseID, ann)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(created)
			}
			_, _ = fmt.Fprintf(app.Out, "Posted announcement %s\n", created.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVar(&text, "text", "", "Announcement text")
	cmd.Flags().StringVar(&state, "state", "PUBLISHED", "Announcement state: PUBLISHED or DRAFT")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func newStreamEditCmd(app *App) *cobra.Command {
	var courseID, announcementID, text string
	cmd := &cobra.Command{
		Use:     "edit [course_id] [announcement_id]",
		Aliases: []string{"update", "set"},
		Short:   "Edit an announcement",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &announcementID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(announcementID, "announcement", "announcement_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeAnnouncements})
			if err != nil {
				return err
			}
			updated, err := client.UpdateAnnouncement(ctx, courseID, announcementID, &gclassroom.Announcement{Text: text}, []string{"text"})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(updated)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated announcement %s\n", updated.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&announcementID, "announcement", "a", "", "Announcement ID")
	cmd.Flags().StringVar(&text, "text", "", "Updated text")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func newStreamDeleteCmd(app *App) *cobra.Command {
	var courseID, announcementID string
	cmd := &cobra.Command{
		Use:     "delete [course_id] [announcement_id]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete an announcement",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &announcementID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(announcementID, "announcement", "announcement_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeAnnouncements})
			if err != nil {
				return err
			}
			if err := client.DeleteAnnouncement(ctx, courseID, announcementID); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "deleted", "announcement": announcementID})
			}
			_, _ = fmt.Fprintf(app.Out, "Deleted announcement %s\n", announcementID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&announcementID, "announcement", "a", "", "Announcement ID")
	return cmd
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}
