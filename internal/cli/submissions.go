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
	cmd := &cobra.Command{Use: "submissions", Short: "Manage submissions"}
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
		Use:   "list",
		Short: "List submissions",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&userID, "user", "", "User ID filter (use me for own submissions)")
	cmd.Flags().Int64Var(&pageSize, "page-size", 50, "Number of submissions to return")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	return cmd
}

func newSubmissionsShowCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a submission",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&submissionID, "submission", "", "Submission ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("submission")
	return cmd
}

func newSubmissionsTurnInCmd(app *App) *cobra.Command {
	return newSubmissionActionCmd(app, "turn-in", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.TurnInSubmission(ctx, courseID, courseWorkID, submissionID)
	})
}

func newSubmissionsUnsubmitCmd(app *App) *cobra.Command {
	return newSubmissionActionCmd(app, "unsubmit", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReclaimSubmission(ctx, courseID, courseWorkID, submissionID)
	})
}

func newSubmissionsReclaimCmd(app *App) *cobra.Command {
	return newSubmissionActionCmd(app, "reclaim", []string{auth.ScopeCourseWorkMe}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReclaimSubmission(ctx, courseID, courseWorkID, submissionID)
	})
}

func newSubmissionsReturnCmd(app *App) *cobra.Command {
	return newSubmissionActionCmd(app, "return", []string{auth.ScopeCourseWorkStudents}, func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error) {
		return client.ReturnSubmission(ctx, courseID, courseWorkID, submissionID)
	})
}

func newSubmissionsGradeCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	var draftGrade float64
	var assignedGrade float64
	var setDraft, setAssigned bool

	cmd := &cobra.Command{
		Use:   "grade",
		Short: "Set draft and/or assigned grade",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&submissionID, "submission", "", "Submission ID")
	cmd.Flags().Float64Var(&draftGrade, "draft", 0, "Draft grade value")
	cmd.Flags().Float64Var(&assignedGrade, "assigned", 0, "Assigned grade value")
	cmd.Flags().BoolVar(&setDraft, "set-draft", false, "Apply --draft")
	cmd.Flags().BoolVar(&setAssigned, "set-assigned", false, "Apply --assigned")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("submission")
	return cmd
}

type submissionActionFunc func(client classroom.ClassroomClient, ctx context.Context, courseID, courseWorkID, submissionID string) (any, error)

func newSubmissionActionCmd(app *App, use string, scopes []string, action submissionActionFunc) *cobra.Command {
	var courseID, courseWorkID, submissionID string
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("%s a submission", strings.ReplaceAll(use, "-", " ")),
		RunE: func(cmd *cobra.Command, args []string) error {
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
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&submissionID, "submission", "", "Submission ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("submission")
	return cmd
}

func titleCase(v string) string {
	v = strings.ReplaceAll(v, "-", " ")
	if v == "" {
		return v
	}
	return strings.ToUpper(v[:1]) + v[1:]
}
