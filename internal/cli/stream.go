package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newStreamCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "stream", Short: "Manage Stream announcements"}
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
		Use:   "list",
		Short: "List announcements in a class stream",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().Int64Var(&pageSize, "page-size", 50, "Number of announcements to return")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newStreamPostCmd(app *App) *cobra.Command {
	var courseID string
	var text string
	var state string
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post an announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&text, "text", "", "Announcement text")
	cmd.Flags().StringVar(&state, "state", "PUBLISHED", "Announcement state: PUBLISHED or DRAFT")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func newStreamEditCmd(app *App) *cobra.Command {
	var courseID, announcementID, text string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&announcementID, "announcement", "", "Announcement ID")
	cmd.Flags().StringVar(&text, "text", "", "Updated text")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("announcement")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func newStreamDeleteCmd(app *App) *cobra.Command {
	var courseID, announcementID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&announcementID, "announcement", "", "Announcement ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("announcement")
	return cmd
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}
