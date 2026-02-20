package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newSubmissionsCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "submissions",
		Aliases: []string{"subs"},
		Short:   "Manage submissions",
		Example: "  gc submissions list -c <course_id> -w <course_work_id> --user me\n  gc submissions turn-in -c <course_id> -w <course_work_id> -s <submission_id>",
	}
	cmd.AddCommand(
		newSubmissionsListCmd(app),
		newSubmissionsShowCmd(app),
		newSubmissionsTurnInCmd(app),
		newSubmissionsUnsubmitCmd(app),
		newSubmissionsGradeCmd(app),
		newSubmissionsReturnCmd(app),
		newSubmissionsReclaimCmd(app),
	)
	return cmd
}

func newSubmissionsListCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, userID, pageToken string
	var pageSize int64
	cmd := &cobra.Command{
		Use:     "list [course_id] [course_work_id]",
		Aliases: []string{"ls"},
		Short:   "List submissions",
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
			scope := auth.ScopeCourseWorkStudentsReadonly
			if userID == "me" {
				scope = auth.ScopeCourseWorkMeReadonly
			}
			client, err := app.ClassroomClient(ctx, []string{scope})
			if err != nil {
				return err
			}
			items, next, err := client.ListStudentSubmissions(ctx, courseID, courseWorkID, userID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"submissions": items, "next_page_token": next})
			}
			for _, sub := range items {
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\t%s\tlate=%t\n", sub.Id, sub.UserId, sub.State, sub.Late)
			}
			if next != "" {
				_, _ = fmt.Fprintf(app.Out, "next_page_token=%s\n", next)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVarP(&userID, "user", "u", "", "User ID filter (use me for own submissions)")
	cmd.Flags().Int64VarP(&pageSize, "page-size", "n", 50, "Number of submissions to return")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Pagination token")
	return cmd
}

func newSubmissionsShowCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	cmd := &cobra.Command{
		Use:     "show [course_id] [course_work_id] [submission_id]",
		Aliases: []string{"get"},
		Short:   "Show a submission",
		Args:    cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID, &submissionID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			if err := requireValue(submissionID, "submission", "submission_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudentsReadonly})
			if err != nil {
				return err
			}
			sub, err := client.GetStudentSubmission(ctx, courseID, courseWorkID, submissionID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(sub)
			}
			_, _ = fmt.Fprintf(app.Out, "Submission %s\nState: %s\nDraft grade: %.2f\nAssigned grade: %.2f\n", sub.Id, sub.State, sub.DraftGrade, sub.AssignedGrade)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVarP(&submissionID, "submission", "s", "", "Submission ID")
	return cmd
}

func newSubmissionsTurnInCmd(app *App) *cobra.Command {
	cmd := newSubmissionActionCmd(app, "turn-in", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.TurnInSubmission(ctx, courseID, courseWorkID, submissionID)
	})
	cmd.Aliases = []string{"submit"}
	return cmd
}

func newSubmissionsUnsubmitCmd(app *App) *cobra.Command {
	cmd := newSubmissionActionCmd(app, "unsubmit", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReclaimSubmission(ctx, courseID, courseWorkID, submissionID)
	})
	cmd.Aliases = []string{"undo-turn-in"}
	return cmd
}

func newSubmissionsReclaimCmd(app *App) *cobra.Command {
	cmd := newSubmissionActionCmd(app, "reclaim", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReclaimSubmission(ctx, courseID, courseWorkID, submissionID)
	})
	cmd.Aliases = []string{"take-back"}
	return cmd
}

func newSubmissionsReturnCmd(app *App) *cobra.Command {
	cmd := newSubmissionActionCmd(app, "return", []string{auth.ScopeCourseWorkStudents}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReturnSubmission(ctx, courseID, courseWorkID, submissionID)
	})
	cmd.Aliases = []string{"send-back"}
	return cmd
}

func newSubmissionsGradeCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	var draftGrade float64
	var assignedGrade float64
	var setDraft, setAssigned bool

	cmd := &cobra.Command{
		Use:     "grade [course_id] [course_work_id] [submission_id]",
		Aliases: []string{"set-grade"},
		Short:   "Set draft and/or assigned grade",
		Args:    cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID, &submissionID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			if err := requireValue(submissionID, "submission", "submission_id"); err != nil {
				return err
			}
			if !setDraft && !setAssigned {
				return fmt.Errorf("provide --set-draft and/or --set-assigned")
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			var draftPtr, assignedPtr *float64
			if setDraft {
				draftPtr = &draftGrade
			}
			if setAssigned {
				assignedPtr = &assignedGrade
			}
			result, err := client.PatchStudentSubmissionGrades(ctx, courseID, courseWorkID, submissionID, draftPtr, assignedPtr)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated grades for submission %s\n", result.Id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVarP(&submissionID, "submission", "s", "", "Submission ID")
	cmd.Flags().Float64Var(&draftGrade, "draft", 0, "Draft grade value")
	cmd.Flags().Float64Var(&assignedGrade, "assigned", 0, "Assigned grade value")
	cmd.Flags().BoolVar(&setDraft, "set-draft", false, "Apply --draft")
	cmd.Flags().BoolVar(&setAssigned, "set-assigned", false, "Apply --assigned")
	return cmd
}

type submissionActionFunc func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error)

func newSubmissionActionCmd(app *App, use string, scopes []string, action submissionActionFunc) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	cmd := &cobra.Command{
		Use:   use + " [course_id] [course_work_id] [submission_id]",
		Short: fmt.Sprintf("%s a submission", strings.ReplaceAll(use, "-", " ")),
		Args:  cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID, &submissionID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			if err := requireValue(submissionID, "submission", "submission_id"); err != nil {
				return err
			}
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, scopes)
			if err != nil {
				return err
			}
			result, err := action(client, ctx, courseID, courseWorkID, submissionID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(result)
			}
			_, _ = fmt.Fprintf(app.Out, "%s submission %s\n", titleCase(use), submissionID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVarP(&submissionID, "submission", "s", "", "Submission ID")
	return cmd
}

func titleCase(v string) string {
	v = strings.ReplaceAll(v, "-", " ")
	if v == "" {
		return v
	}
	return strings.ToUpper(v[:1]) + v[1:]
}
