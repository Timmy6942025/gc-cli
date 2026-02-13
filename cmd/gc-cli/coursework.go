package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/api"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func CourseworkCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "coursework",
		Aliases: []string{"classwork", "cw"},
		Usage:   "manage classwork for a class",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list classwork for a class",
				Action: handleCourseworkList(cfg),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "course",
						Usage:    "course ID to list classwork for",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "output as JSON",
					},
					&cli.BoolFlag{
						Name:  "all",
						Usage: "include all classwork (including draft)",
					},
				},
			},
		},
	}
}

func handleCourseworkList(cfg *config.Config) func(*cli.Context) error {
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

		courseID := c.String("course")
		if courseID == "" {
			return fmt.Errorf("course ID is required (use --course flag)")
		}

		if _, err := client.GetCourse(ctx, courseID); err != nil {
			return fmt.Errorf("course %s not found or access denied: %w", courseID, err)
		}

		coursework, _, err := client.ListCourseWork(ctx, courseID, 100)
		if err != nil {
			return fmt.Errorf("failed to list classwork: %w", err)
		}

		filteredCoursework := coursework
		if !c.Bool("all") {
			filteredCoursework = []api.CourseWork{}
			for _, cw := range coursework {
				if cw.State == "PUBLISHED" {
					filteredCoursework = append(filteredCoursework, cw)
				}
			}
		}

		sort.Slice(filteredCoursework, func(i, j int) bool {
			dateI := getDueDate(filteredCoursework[i])
			dateJ := getDueDate(filteredCoursework[j])

			if dateI.IsZero() && dateJ.IsZero() {
				return false
			}
			if dateI.IsZero() {
				return false
			}
			if dateJ.IsZero() {
				return true
			}
			return dateI.Before(dateJ)
		})

		if c.Bool("json") {
			return outputCourseworkJSON(filteredCoursework)
		}
		return outputCourseworkTable(filteredCoursework)
	}
}

func getDueDate(cw api.CourseWork) time.Time {
	if cw.DueDate == nil {
		return time.Time{}
	}
	return time.Date(cw.DueDate.Year, time.Month(cw.DueDate.Month), cw.DueDate.Day, 0, 0, 0, 0, time.UTC)
}

func getStatus(cw api.CourseWork) string {
	if cw.State == "DRAFT" {
		return "Draft"
	}

	if cw.DueDate != nil {
		dueDate := getDueDate(cw)
		var dueTime time.Time
		if cw.DueTime != nil {
			dueTime = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(),
				cw.DueTime.Hours, cw.DueTime.Minutes, cw.DueTime.Seconds, 0, time.UTC)
		} else {
			dueTime = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 23, 59, 59, 0, time.UTC)
		}

		if time.Now().After(dueTime) {
			return "Overdue"
		}
	}

	return "Pending"
}

func formatDueDate(cw api.CourseWork) string {
	if cw.DueDate == nil {
		return "-"
	}
	date := fmt.Sprintf("%d/%02d/%02d", cw.DueDate.Year, cw.DueDate.Month, cw.DueDate.Day)
	if cw.DueTime != nil {
		date += fmt.Sprintf(" %02d:%02d", cw.DueTime.Hours, cw.DueTime.Minutes)
	}
	return date
}

func outputCourseworkJSON(coursework []api.CourseWork) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(coursework)
}

func outputCourseworkTable(coursework []api.CourseWork) error {
	if len(coursework) == 0 {
		fmt.Println("No classwork found.")
		return nil
	}

	idWidth := 12
	titleWidth := 40
	dueDateWidth := 16
	statusWidth := 12

	for _, cw := range coursework {
		if len(cw.ID) > idWidth {
			idWidth = len(cw.ID)
		}
		if len(cw.Title) > titleWidth {
			titleWidth = len(cw.Title)
		}
		dueStr := formatDueDate(cw)
		if len(dueStr) > dueDateWidth {
			dueDateWidth = len(dueStr)
		}
		status := getStatus(cw)
		if len(status) > statusWidth {
			statusWidth = len(status)
		}
	}

	if idWidth < 12 {
		idWidth = 12
	}
	if titleWidth < 40 {
		titleWidth = 40
	}
	if dueDateWidth < 16 {
		dueDateWidth = 16
	}
	if statusWidth < 12 {
		statusWidth = 12
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(idWidth).Render("ID"),
		headerStyle.Width(titleWidth).Render("Title"),
		headerStyle.Width(dueDateWidth).Render("Due Date"),
		headerStyle.Width(statusWidth).Render("Status"),
	)
	separator := separatorStyle.Render("─")

	fmt.Println(header)
	fmt.Println(lipgloss.JoinHorizontal(
		lipgloss.Left,
		separator+separator+separator+separator,
	))

	for _, cw := range coursework {
		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cellStyle.Width(idWidth).Render(truncate(cw.ID, idWidth)),
			cellStyle.Width(titleWidth).Render(truncate(cw.Title, titleWidth)),
			cellStyle.Width(dueDateWidth).Render(formatDueDate(cw)),
			cellStyle.Width(statusWidth).Render(getStatus(cw)),
		)
		fmt.Println(row)
	}

	fmt.Println()
	fmt.Printf("Total: %d classwork item(s)\n", len(coursework))
	return nil
}
