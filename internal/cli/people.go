package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newPeopleCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "people", Short: "Manage People"}
	cmd.AddCommand(newPeopleListCmd(app), newPeopleInviteCmd(app), newPeopleRemoveCmd(app))
	return cmd
}

func newPeopleListCmd(app *App) *cobra.Command {
	var courseID, role, pageToken string
	var pageSize int64
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List people in a class",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeRostersReadonly})
			if err != nil {
				return err
			}
			role = strings.ToLower(role)
			if app.Printer.JSON {
				payload := map[string]any{}
				if role == "all" || role == "teachers" || role == "teacher" {
					teachers, next, err := client.ListTeachers(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
					if err != nil {
						return err
					}
					payload["teachers"] = teachers
					payload["teachers_next_page_token"] = next
				}
				if role == "all" || role == "students" || role == "student" {
					students, next, err := client.ListStudents(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
					if err != nil {
						return err
					}
					payload["students"] = students
					payload["students_next_page_token"] = next
				}
				return app.Printer.PrintJSON(payload)
			}
			if role == "all" || role == "teachers" || role == "teacher" {
				teachers, _, err := client.ListTeachers(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(app.Out, "Teachers:")
				for _, teacher := range teachers {
					_, _ = fmt.Fprintf(app.Out, "  %s\t%s\n", teacher.UserId, personName(teacher.Profile))
				}
			}
			if role == "all" || role == "students" || role == "student" {
				students, _, err := client.ListStudents(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(app.Out, "Students:")
				for _, student := range students {
					_, _ = fmt.Fprintf(app.Out, "  %s\t%s\n", student.UserId, personName(student.Profile))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&role, "role", "all", "Role filter: all|teachers|students")
	cmd.Flags().Int64Var(&pageSize, "page-size", 50, "Number of results")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newPeopleInviteCmd(app *App) *cobra.Command {
	var courseID, role, userID string
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a teacher or student",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeRosters})
			if err != nil {
				return err
			}
			role = strings.ToLower(role)
			var result any
			switch role {
			case "teacher", "teachers":
				result, err = client.InviteTeacher(ctx, courseID, userID)
			case "student", "students":
				result, err = client.InviteStudent(ctx, courseID, userID)
			default:
				return fmt.Errorf("invalid role %q; expected teacher or student", role)
			}
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "Created %s invitation for %s\n", role, userID)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&role, "role", "student", "Role: teacher|student")
	cmd.Flags().StringVar(&userID, "user", "", "User ID or email")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

func newPeopleRemoveCmd(app *App) *cobra.Command {
	var courseID, role, userID string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a teacher or student",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeRosters})
			if err != nil {
				return err
			}
			role = strings.ToLower(role)
			switch role {
			case "teacher", "teachers":
				err = client.RemoveTeacher(ctx, courseID, userID)
			case "student", "students":
				err = client.RemoveStudent(ctx, courseID, userID)
			default:
				return fmt.Errorf("invalid role %q; expected teacher or student", role)
			}
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "removed", "role": role, "user": userID})
			}
			_, _ = fmt.Fprintf(app.Out, "Removed %s %s\n", role, userID)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&role, "role", "student", "Role: teacher|student")
	cmd.Flags().StringVar(&userID, "user", "", "User ID or email")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

func personName(profile *gclassroom.UserProfile) string {
	if profile == nil || profile.Name == nil {
		return ""
	}
	if profile.Name.FullName != "" {
		return profile.Name.FullName
	}
	if profile.EmailAddress != "" {
		return profile.EmailAddress
	}
	return profile.Id
}
