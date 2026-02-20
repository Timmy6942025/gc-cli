package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timothy/gc-cli/internal/auth"
	"github.com/timothy/gc-cli/internal/classroom"
)

func newTopicsCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "topics",
		Aliases: []string{"topic"},
		Short:   "Manage classwork topics",
		Example: "  gc topics list -c <course_id>\n  gc topics move -c <course_id> -w <course_work_id> -t <topic_id>",
	}
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
		Use:     "list [course_id]",
		Aliases: []string{"ls"},
		Short:   "List topics",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
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
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().Int64VarP(&pageSize, "page-size", "n", 50, "Number of topics")
	cmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Pagination token")
	return cmd
}

func newTopicsCreateCmd(app *App) *cobra.Command {
	var courseID, name string
	cmd := &cobra.Command{
		Use:     "create [course_id]",
		Aliases: []string{"add"},
		Short:   "Create a topic",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
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
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVar(&name, "name", "", "Topic name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTopicsEditCmd(app *App) *cobra.Command {
	var courseID, topicID, name string
	cmd := &cobra.Command{
		Use:     "edit [course_id] [topic_id]",
		Aliases: []string{"update", "set"},
		Short:   "Edit a topic",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &topicID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(topicID, "topic", "topic_id"); err != nil {
				return err
			}
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
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&topicID, "topic", "t", "", "Topic ID")
	cmd.Flags().StringVar(&name, "name", "", "New topic name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTopicsDeleteCmd(app *App) *cobra.Command {
	var courseID, topicID string
	cmd := &cobra.Command{
		Use:     "delete [course_id] [topic_id]",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a topic",
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &topicID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(topicID, "topic", "topic_id"); err != nil {
				return err
			}
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
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&topicID, "topic", "t", "", "Topic ID")
	return cmd
}

func newTopicsMoveCmd(app *App) *cobra.Command {
	var courseID, courseWorkID, topicID string
	cmd := &cobra.Command{
		Use:     "move [course_id] [course_work_id] [topic_id]",
		Aliases: []string{"set-topic"},
		Short:   "Move classwork into a topic",
		Args:    cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fillPositional(args, &courseID, &courseWorkID, &topicID)
			if err := requireValue(courseID, "course", "course_id"); err != nil {
				return err
			}
			if err := requireValue(courseWorkID, "course-work", "course_work_id"); err != nil {
				return err
			}
			if err := requireValue(topicID, "topic", "topic_id"); err != nil {
				return err
			}
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
	cmd.Flags().StringVarP(&courseID, "course", "c", "", "Course ID")
	cmd.Flags().StringVarP(&courseWorkID, "course-work", "w", "", "CourseWork ID")
	cmd.Flags().StringVarP(&topicID, "topic", "t", "", "Topic ID")
	return cmd
}
