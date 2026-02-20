package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newClassworkCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "classwork",
		Aliases: []string{"coursework", "cw"},
		Short:   "Manage Classwork",
		Example: "  gc classwork list -c <course_id>\n  gc cw create -c <course_id> --title \"Worksheet 4\"\n  gc classwork publish -c <course_id> -w <course_work_id>",
	}
	cmd.AddCommand(
		newClassworkListCmd(app),
		newClassworkCreateCmd(app),
		newClassworkEditCmd(app),
		newClassworkPublishCmd(app),
		newClassworkScheduleCmd(app),
		newClassworkDeleteCmd(app),
	)
	return cmd
}

func newClassworkListCmd(app *App) *cobra.Command {
	var courseID string
	var pageSize int64
	var pageToken string
	cmd := &cobra.Command{
		Use:     "list [course_id]",
		Aliases: []string{"ls"},
		Short:   "List classwork items",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudentsReadonly})
			if err != nil {
				return err
			}
			items, next, err := client.ListCourseWork(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"classwork": items, "next_page_token": next})
			}
			for _, item := range items {
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\t%s\t%s\n", item.Id, item.Title, item.WorkType, item.State)
			}
			if next != "" {
				_, _ = fmt.Fprintf(app.Out, "next_page_token=%s\n", next)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().Int64VarP(&pageSize, "page-size", "n", 50, "Number of classwork items to return")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Pagination token")
	return cmd
}

func newClassworkCreateCmd(app *App) *cobra.Command {
	var courseID, title, description, workType, state, topicID, dueRaw string
	var maxPoints float64
	var driveFileIDs []string
	var linkURLs []string
	var uploadPaths []string

	cmd := &cobra.Command{
		Use:     "create [course_id]",
		Aliases: []string{"add", "new"},
		Short:   "Create classwork",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			requiredScopes := []string{auth.ScopeCourseWorkStudents}
			if len(driveFileIDs) > 0 || len(uploadPaths) > 0 {
				requiredScopes = append(requiredScopes, auth.ScopeDriveFile)
			}
			client, err := app.ClassroomClient(ctx, requiredScopes)
			if err != nil {
				return err
			}
			materials, err := buildMaterials(ctx, app, driveFileIDs, linkURLs, uploadPaths)
			if err != nil {
				return err
			}

			cw := &gclassroom.CourseWork{
				Title:       title,
				Description: description,
				WorkType:    workType,
				State:       state,
				TopicId:     topicID,
				MaxPoints:   maxPoints,
				Materials:   materials,
			}
			if dueDate, dueTime, ok, err := parseDueInput(dueRaw); err != nil {
				return err
			} else if ok {
				cw.DueDate = dueDate
				cw.DueTime = dueTime
			}

			created, err := client.CreateCourseWork(ctx, courseID, cw)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(created)
			}
			_, _ = fmt.Fprintf(app.Out, "Created classwork %s (%s)\n", created.Title, created.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVar(&title, "title", "", "Classwork title")
	cmd.Flags().StringVar(&description, "description", "", "Classwork description")
	cmd.Flags().StringVar(&workType, "type", "ASSIGNMENT", "Classwork type")
	cmd.Flags().StringVar(&state, "state", "DRAFT", "Classwork state: DRAFT or PUBLISHED")
	cmd.Flags().StringVar(&topicID, "topic", "", "Topic ID")
	cmd.Flags().Float64Var(&maxPoints, "max-points", 100, "Maximum points")
	cmd.Flags().StringVar(&dueRaw, "due", "", "Due timestamp (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringArrayVar(&driveFileIDs, "drive-file-id", nil, "Attach a Google Drive file ID (repeatable)")
	cmd.Flags().StringArrayVar(&linkURLs, "link", nil, "Attach a link URL (repeatable)")
	cmd.Flags().StringArrayVar(&uploadPaths, "upload-file", nil, "Upload and attach a local file path (repeatable)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func newClassworkEditCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, title, description, topicID, dueRaw string
	var maxPoints float64
	var setMaxPoints bool

	cmd := &cobra.Command{
		Use:     "edit [course_id] [course_work_id]",
		Aliases: []string{"update", "set"},
		Short:   "Edit classwork",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			patch := &gclassroom.CourseWork{}
			mask := make([]string, 0)
			if title != "" {
				patch.Title = title
				mask = append(mask, "title")
			}
			if description != "" {
				patch.Description = description
				mask = append(mask, "description")
			}
			if topicID != "" {
				patch.TopicId = topicID
				mask = append(mask, "topicId")
			}
			if setMaxPoints {
				patch.MaxPoints = maxPoints
				mask = append(mask, "maxPoints")
			}
			if dueDate, dueTime, ok, err := parseDueInput(dueRaw); err != nil {
				return err
			} else if ok {
				patch.DueDate = dueDate
				patch.DueTime = dueTime
				mask = append(mask, "dueDate", "dueTime")
			}
			updated, err := client.PatchCourseWork(ctx, courseID, courseWorkID, patch, mask)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(updated)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated classwork %s\n", updated.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVar(&title, "title", "", "Classwork title")
	cmd.Flags().StringVar(&description, "description", "", "Classwork description")
	cmd.Flags().StringVar(&topicID, "topic", "", "Topic ID")
	cmd.Flags().Float64Var(&maxPoints, "max-points", 0, "Maximum points")
	cmd.Flags().BoolVar(&setMaxPoints, "set-max-points", false, "Apply max-points value")
	cmd.Flags().StringVar(&dueRaw, "due", "", "Due timestamp (RFC3339 or YYYY-MM-DD)")
	return cmd
}

func newClassworkPublishCmd(app *App) *cobra.Command {
	var courseID, courseWorkID string
	cmd := &cobra.Command{
		Use:     "publish [course_id] [course_work_id]",
		Aliases: []string{"pub"},
		Short:   "Publish classwork",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			item, err := client.PublishCourseWork(ctx, courseID, courseWorkID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(item)
			}
			_, _ = fmt.Fprintf(app.Out, "Published classwork %s\n", item.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	return cmd
}

func newClassworkScheduleCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, whenRaw string
	cmd := &cobra.Command{
		Use:     "schedule [course_id] [course_work_id]",
		Aliases: []string{"sched"},
		Short:   "Schedule classwork publication",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			when, err := time.Parse(time.RFC3339, whenRaw)
			if err != nil {
				return fmt.Errorf("parse --when: %w", err)
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			item, err := client.ScheduleCourseWork(ctx, courseID, courseWorkID, when)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(item)
			}
			_, _ = fmt.Fprintf(app.Out, "Scheduled classwork %s at %s\n", item.Id, when.Format(time.RFC3339))
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVar(&whenRaw, "when", "", "Publish timestamp (RFC3339)")
	_ = cmd.MarkFlagRequired("when")
	return cmd
}

func newClassworkDeleteCmd(app *App) *cobra.Command {
	var courseID, courseWorkID string
	cmd := &cobra.Command{
		Use:     "delete [course_id] [course_work_id]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete classwork",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			if err := client.DeleteCourseWork(ctx, courseID, courseWorkID); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "deleted", "course_work": courseWorkID})
			}
			_, _ = fmt.Fprintf(app.Out, "Deleted classwork %s\n", courseWorkID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	return cmd
}

func buildMaterials(ctx context.Context, app *App, driveFileIDs, linkURLs, uploadPaths []string) ([]*gclassroom.Material, error) {
	materials := make([]*gclassroom.Material, 0, len(driveFileIDs)+len(linkURLs)+len(uploadPaths))
	for _, u := range linkURLs {
		materials = append(materials, &gclassroom.Material{Link: &gclassroom.Link{Url: u}})
	}
	if len(driveFileIDs) == 0 && len(uploadPaths) == 0 {
		return materials, nil
	}
	driveClient, err := app.DriveClient(ctx, []string{auth.ScopeDriveFile})
	if err != nil {
		return nil, err
	}
	for _, fileID := range driveFileIDs {
		file, err := driveClient.GetFile(ctx, fileID)
		if err != nil {
			return nil, err
		}
		materials = append(materials, &gclassroom.Material{DriveFile: &gclassroom.SharedDriveFile{DriveFile: &gclassroom.DriveFile{Id: file.Id, Title: file.Name, AlternateLink: file.WebViewLink}}})
	}
	for _, path := range uploadPaths {
		file, err := driveClient.UploadFile(ctx, path)
		if err != nil {
			return nil, err
		}
		title := file.Name
		if title == "" {
			title = filepath.Base(path)
		}
		materials = append(materials, &gclassroom.Material{DriveFile: &gclassroom.SharedDriveFile{DriveFile: &gclassroom.DriveFile{Id: file.Id, Title: title, AlternateLink: file.WebViewLink}}})
	}
	return materials, nil
}

func parseDueInput(raw string) (*gclassroom.Date, *gclassroom.TimeOfDay, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil, false, nil
	}
	var t time.Time
	var err error
	if len(raw) == len("2006-01-02") {
		t, err = time.Parse("2006-01-02", raw)
	} else {
		t, err = time.Parse(time.RFC3339, raw)
	}
	if err != nil {
		return nil, nil, false, fmt.Errorf("parse due value: %w", err)
	}
	date := &gclassroom.Date{Year: int64(t.Year()), Month: int64(t.Month()), Day: int64(t.Day())}
	timeOfDay := &gclassroom.TimeOfDay{Hours: int64(t.Hour()), Minutes: int64(t.Minute()), Seconds: int64(t.Second()), Nanos: int64(t.Nanosecond())}
	return date, timeOfDay, true, nil
}
