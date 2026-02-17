package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newClassesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "classes", Short: "Manage classes"}
	cmd.AddCommand(
		newClassesListCmd(app),
		newClassesShowCmd(app),
		newClassesCreateCmd(app),
		newClassesUpdateCmd(app),
		newClassesArchiveCmd(app),
		newClassesRestoreCmd(app),
		newClassesDeleteCmd(app),
	)
	return cmd
}

func newClassesListCmd(app *App) *cobra.Command {
	var pageSize int64
	var pageToken string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List classes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCoursesReadonly})
			if err != nil {
				return err
			}
			courses, next, err := client.ListCourses(ctx, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"courses": courses, "next_page_token": next})
			}
			for _, c := range courses {
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\t%s\n", c.Id, c.Name, c.CourseState)
			}
			if next != "" {
				_, _ = fmt.Fprintf(app.Out, "next_page_token=%s\n", next)
			}
			return nil
		},
	}
	cmd.Flags().Int64Var(&pageSize, "page-size", 50, "Number of classes to return")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	return cmd
}

func newClassesShowCmd(app *App) *cobra.Command {
	var courseID string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show class details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCoursesReadonly})
			if err != nil {
				return err
			}
			course, err := client.GetCourse(ctx, courseID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(course)
			}
			_, _ = fmt.Fprintf(app.Out, "ID: %s\nName: %s\nSection: %s\nState: %s\n", course.Id, course.Name, course.Section, course.CourseState)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newClassesCreateCmd(app *App) *cobra.Command {
	var name, section, description, room, ownerID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a class",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourses})
			if err != nil {
				return err
			}
			course := &gclassroom.Course{
				Name:               name,
				Section:            section,
				DescriptionHeading: description,
				Description:        description,
				Room:               room,
				OwnerId:            ownerID,
			}
			created, err := client.CreateCourse(ctx, course)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(created)
			}
			_, _ = fmt.Fprintf(app.Out, "Created class %s (%s)\n", created.Name, created.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Class name")
	cmd.Flags().StringVar(&section, "section", "", "Class section")
	cmd.Flags().StringVar(&description, "description", "", "Class description")
	cmd.Flags().StringVar(&room, "room", "", "Class room")
	cmd.Flags().StringVar(&ownerID, "owner", "", "Owner user ID (optional)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newClassesUpdateCmd(app *App) *cobra.Command {
	var courseID, name, section, description, room string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update class metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourses})
			if err != nil {
				return err
			}
			patch := &gclassroom.Course{}
			mask := make([]string, 0)
			if name != "" {
				patch.Name = name
				mask = append(mask, "name")
			}
			if section != "" {
				patch.Section = section
				mask = append(mask, "section")
			}
			if description != "" {
				patch.DescriptionHeading = description
				patch.Description = description
				mask = append(mask, "descriptionHeading", "description")
			}
			if room != "" {
				patch.Room = room
				mask = append(mask, "room")
			}
			updated, err := client.UpdateCourse(ctx, courseID, patch, mask)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(updated)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated class %s (%s)\n", updated.Name, updated.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&name, "name", "", "Class name")
	cmd.Flags().StringVar(&section, "section", "", "Class section")
	cmd.Flags().StringVar(&description, "description", "", "Class description")
	cmd.Flags().StringVar(&room, "room", "", "Class room")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newClassesArchiveCmd(app *App) *cobra.Command {
	return newClassStateCmd(app, "archive", "ARCHIVED")
}

func newClassesRestoreCmd(app *App) *cobra.Command {
	return newClassStateCmd(app, "restore", "ACTIVE")
}

func newClassStateCmd(app *App, use, state string) *cobra.Command {
	var courseID string
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Set class state to %s", state),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourses})
			if err != nil {
				return err
			}
			course, err := client.SetCourseState(ctx, courseID, state)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(course)
			}
			_, _ = fmt.Fprintf(app.Out, "Class %s state -> %s\n", course.Id, course.CourseState)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newClassesDeleteCmd(app *App) *cobra.Command {
	var courseID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a class",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourses})
			if err != nil {
				return err
			}
			if err := client.DeleteCourse(ctx, courseID); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "deleted", "course": courseID})
			}
			_, _ = fmt.Fprintf(app.Out, "Deleted class %s\n", courseID)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func ctx(cmd *cobra.Command) context.Context {
	if cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}
