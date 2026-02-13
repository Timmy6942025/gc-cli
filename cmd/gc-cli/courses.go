package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/api"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func CoursesCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "courses",
		Usage: "list your enrolled courses",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all enrolled courses",
				Action: handleCoursesList(cfg),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "json",
						Usage: "output as JSON",
					},
				},
			},
		},
	}
}

func handleCoursesList(cfg *config.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		ctx := context.Background()

		token, err := auth.GetValidToken(ctx, auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile))
		if err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		authCfg := auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile)
		client, err := api.NewClientFromToken(ctx, authCfg.OAuth2Config(), token)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		courses, _, err := client.ListCourses(ctx, 100)
		if err != nil {
			return fmt.Errorf("failed to list courses: %w (debug: %+v)", err, err)
		}

		var studentCourses []api.Course
		for _, course := range courses {
			if course.CourseState == "ACTIVE" {
				studentCourses = append(studentCourses, course)
			}
		}

		if c.Bool("json") {
			return outputJSON(studentCourses)
		}
		return outputTable(studentCourses)
	}
}

func outputJSON(courses []api.Course) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(courses)
}

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)
	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

func outputTable(courses []api.Course) error {
	if len(courses) == 0 {
		fmt.Println("No enrolled courses found.")
		return nil
	}

	idWidth := 12
	nameWidth := 40
	sectionWidth := 20
	roomWidth := 15

	for _, c := range courses {
		if len(c.ID) > idWidth {
			idWidth = len(c.ID)
		}
		if len(c.Name) > nameWidth {
			nameWidth = len(c.Name)
		}
		if len(c.Section) > sectionWidth {
			sectionWidth = len(c.Section)
		}
		if len(c.Room) > roomWidth {
			roomWidth = len(c.Room)
		}
	}

	// Print header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(idWidth).Render("ID"),
		headerStyle.Width(nameWidth).Render("Name"),
		headerStyle.Width(sectionWidth).Render("Section"),
		headerStyle.Width(roomWidth).Render("Room"),
	)
	separator := separatorStyle.Render("─")

	fmt.Println(header)
	fmt.Println(lipgloss.JoinHorizontal(
		lipgloss.Left,
		separator+separator+separator+separator,
	))

	for _, c := range courses {
		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cellStyle.Width(idWidth).Render(truncate(c.ID, idWidth)),
			cellStyle.Width(nameWidth).Render(truncate(c.Name, nameWidth)),
			cellStyle.Width(sectionWidth).Render(truncate(c.Section, sectionWidth)),
			cellStyle.Width(roomWidth).Render(truncate(c.Room, roomWidth)),
		)
		fmt.Println(row)
	}

	fmt.Println()
	fmt.Printf("Total: %d course(s)\n", len(courses))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
