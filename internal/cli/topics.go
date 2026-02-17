package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newTopicsCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{Use: "topics", Short: "Manage classwork topics"}
	cmd.AddCommand(
		newTopicsListCmd(app),
		newTopicsCreateCmd(app),
		newTopicsEditCmd(app),
		newTopicsDeleteCmd(app),
		newTopicsMoveCmd(app),
	)
	return cmd
}

func newTopicsListCmd(app *App) *cobra.Command {
	var courseID, pageToken string
	var pageSize int64
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List topics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeTopicsReadonly})
			if err != nil {
				return err
			}
			topics, next, err := client.ListTopics(ctx, courseID, classroom.ListParams{PageSize: pageSize, PageToken: pageToken})
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]any{"topics": topics, "next_page_token": next})
			}
			for _, t := range topics {
				_, _ = fmt.Fprintf(app.Out, "%s\t%s\n", t.TopicId, t.Name)
			}
			if next != "" {
				_, _ = fmt.Fprintf(app.Out, "next_page_token=%s\n", next)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().Int64Var(&pageSize, "page-size", 50, "Number of topics")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	_ = cmd.MarkFlagRequired("course")
	return cmd
}

func newTopicsCreateCmd(app *App) *cobra.Command {
	var courseID, name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeTopics})
			if err != nil {
				return err
			}
			topic, err := client.CreateTopic(ctx, courseID, name)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(topic)
			}
			_, _ = fmt.Fprintf(app.Out, "Created topic %s (%s)\n", topic.Name, topic.TopicId)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&name, "name", "", "Topic name")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTopicsEditCmd(app *App) *cobra.Command {
	var courseID, topicID, name string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeTopics})
			if err != nil {
				return err
			}
			topic, err := client.UpdateTopic(ctx, courseID, topicID, name)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(topic)
			}
			_, _ = fmt.Fprintf(app.Out, "Updated topic %s\n", topic.TopicId)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&topicID, "topic", "", "Topic ID")
	cmd.Flags().StringVar(&name, "name", "", "New topic name")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTopicsDeleteCmd(app *App) *cobra.Command {
	var courseID, topicID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeTopics})
			if err != nil {
				return err
			}
			if err := client.DeleteTopic(ctx, courseID, topicID); err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(map[string]string{"status": "deleted", "topic": topicID})
			}
			_, _ = fmt.Fprintf(app.Out, "Deleted topic %s\n", topicID)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&topicID, "topic", "", "Topic ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("topic")
	return cmd
}

func newTopicsMoveCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, topicID string
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move classwork into a topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := ctx(cmd)
			client, err := app.ClassroomClient(ctx, []string{auth.ScopeCourseWorkStudents})
			if err != nil {
				return err
			}
			work, err := client.MoveCourseWorkToTopic(ctx, courseID, courseWorkID, topicID)
			if err != nil {
				return err
			}
			if app.Printer.JSON {
				return app.Printer.PrintJSON(work)
			}
			_, _ = fmt.Fprintf(app.Out, "Moved classwork %s to topic %s\n", work.Id, topicID)
			return nil
		},
	}
	cmd.Flags().StringVar(&courseID, "course", "", "Course ID")
	cmd.Flags().StringVar(&courseWorkID, "course-work", "", "CourseWork ID")
	cmd.Flags().StringVar(&topicID, "topic", "", "Topic ID")
	_ = cmd.MarkFlagRequired("course")
	_ = cmd.MarkFlagRequired("course-work")
	_ = cmd.MarkFlagRequired("topic")
	return cmd
}
