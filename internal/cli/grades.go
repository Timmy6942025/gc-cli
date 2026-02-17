package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newGradesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "grades", Short: "Manage grades"}
	cmd.AddCommand(newGradesListCmd(app), newGradesSetDraftCmd(app), newGradesSetAssignedCmd(app), newGradesReturnCmd(app))
	return cmd
}

func newGradesListCmd(app *App) *cobra.Command {
	var courseID, courseWorkID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List grades",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudentsReadonly})
			if err != nil {
				return err
			}
			if courseWorkID != "" {
				subs, _, err := client.ListStudentSubmissions(ctx, courseID, courseWorkID, "", classroom.ListParams{PageSize: 200})
				if err != nil {
					return err
				}
				if app.Printer.JSON {
					return app.Printer.PrintJSON(map[string]any{"course_work": courseWorkID, "submissions": subs})
				}
				for _, s := range subs {
					_, _ = fmt.Fprintf(app.Out, "%s\tuser=%s\tdraft=%.2f\tassigned=%.2f\tstate=%s\n", s.Id, s.UserId, s.DraftGrade, s.AssignedGrade, s.State)
				}
				return nil
			}
			work, _, err := client.ListCourseWork(ctx, courseID, classroom.ListParams{PageSize: 200})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				gradebook := make([]map[string]any, 0, len(work))
				for _, item := range work {
					subs, _, err := client.ListStudentSubmissions(ctx, courseID, item.Id, "", classroom.ListParams{PageSize: 200})
					if err != nil {
						return err
					}
					gradebook = append(gradebook, map[string]any{
						"course_work_id": item.Id,
						"title":          item.Title,
						"submissions":    subs,
					})
				}
				return app.Printer.PrintJSON(map[string]any{"course": courseID, "gradebook": gradebook})
			}
			for _, item := range work {
				subs, _, err := client.ListStudentSubmissions(ctx, courseID, item.Id, "", classroom.ListParams{PageSize: 200})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\n", item.Id, item.Title)
				for _, s := range subs {
					_, _ = fmt.Fprintf(app.Out, "  %s\tuser=%s\tdraft=%.2f\tassigned=%.2f\n", s.Id, s.UserId, s.DraftGrade, s.AssignedGrade)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID (optional)")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newGradesSetDraftCmd(app *App) *cobra.Command {
	return newGradesSetCmd(app, "set-draft", true)
}

func newGradesSetAssignedCmd(app *App) *cobra.Command {
	return newGradesSetCmd(app, "set-assigned", false)
}

func newGradesSetCmd(app *App, use string, draft bool) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	var value float64
	cmd := &cobra.Command{
		Use:   use,
		Short: "Set a grade value",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			var draftPtr, assignedPtr *float64
			if draft {
				draftPtr = &value
			} else {
				assignedPtr = &value
			}
			result, err := client.PatchStudentSubmissionGrades(ctx, courseID, courseWorkID, submissionID, draftPtr, assignedPtr)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated grade for submission %s\n", result.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&submissionID, "submission", "", "Submission ID")
	cmd.Flags().Float64Var(&value, "value", 0, "Grade value")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("submission")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func newGradesReturnCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	cmd := &cobra.Command{
		Use:   "return",
		Short: "Return graded work",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			result, err := client.ReturnSubmission(ctx, courseID, courseWorkID, submissionID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "Returned submission %s\n", result.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&submissionID, "submission", "", "Submission ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("submission")
	return cmd
}
