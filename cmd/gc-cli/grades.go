package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/api"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/urfave/cli/v2"
)

type GradeEntry struct {
	Assignment string
	Grade      string
	MaxPoints  string
	Feedback   string
}

func GradesCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "grades",
		Usage: "view your grades for a course",
		Action: func(c *cli.Context) error {
			return handleGrades(c, cfg)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "course",
				Usage:    "course ID to view grades for",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "output as JSON",
			},
		},
	}
}

func handleGrades(c *cli.Context, cfg *config.Config) error {
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

	coursework, _, err := client.ListCourseWork(ctx, courseID, 100)
	if err != nil {
		return fmt.Errorf("failed to list coursework: %w", err)
	}

	var publishedCoursework []api.CourseWork
	for _, cw := range coursework {
		if cw.State == "PUBLISHED" {
			publishedCoursework = append(publishedCoursework, cw)
		}
	}

	var grades []GradeEntry
	for _, cw := range publishedCoursework {
		submission, err := client.GetMySubmission(ctx, courseID, cw.ID)
		if err != nil {
			continue
		}

		if submission.AssignedGrade > 0 || submission.DraftGrade > 0 {
			grade := submission.AssignedGrade
			if grade == 0 && submission.DraftGrade > 0 {
				grade = submission.DraftGrade
			}

			feedback := "Not returned"
			if !submission.ReturnTimestamp.IsZero() {
				feedback = "Returned"
			} else if submission.State == "TURNED_IN" {
				feedback = "Graded"
			}

			grades = append(grades, GradeEntry{
				Assignment: cw.Title,
				Grade:      fmt.Sprintf("%.1f", grade),
				MaxPoints:  fmt.Sprintf("%d", cw.MaxPoints),
				Feedback:   feedback,
			})
		}
	}

	if c.Bool("json") {
		return outputGradesJSON(grades)
	}
	return outputGradesTable(grades)
}

func outputGradesJSON(grades []GradeEntry) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(grades)
}

func outputGradesTable(grades []GradeEntry) error {
	if len(grades) == 0 {
		fmt.Println("No grades yet")
		return nil
	}

	sort.Slice(grades, func(i, j int) bool {
		return grades[i].Assignment < grades[j].Assignment
	})

	assignmentWidth := 40
	gradeWidth := 10
	maxPointsWidth := 12
	feedbackWidth := 15

	for _, g := range grades {
		if len(g.Assignment) > assignmentWidth {
			assignmentWidth = len(g.Assignment)
		}
		if len(g.Grade) > gradeWidth {
			gradeWidth = len(g.Grade)
		}
		if len(g.MaxPoints) > maxPointsWidth {
			maxPointsWidth = len(g.MaxPoints)
		}
		if len(g.Feedback) > feedbackWidth {
			feedbackWidth = len(g.Feedback)
		}
	}

	if assignmentWidth < 40 {
		assignmentWidth = 40
	}
	if gradeWidth < 10 {
		gradeWidth = 10
	}
	if maxPointsWidth < 12 {
		maxPointsWidth = 12
	}
	if feedbackWidth < 15 {
		feedbackWidth = 15
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(assignmentWidth).Render("Assignment"),
		headerStyle.Width(gradeWidth).Render("Grade"),
		headerStyle.Width(maxPointsWidth).Render("Max Points"),
		headerStyle.Width(feedbackWidth).Render("Feedback"),
	)
	separator := separatorStyle.Render("─")

	fmt.Println(header)
	fmt.Println(lipgloss.JoinHorizontal(
		lipgloss.Left,
		separator+separator+separator+separator,
	))

	for _, g := range grades {
		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cellStyle.Width(assignmentWidth).Render(truncate(g.Assignment, assignmentWidth)),
			cellStyle.Width(gradeWidth).Render(g.Grade),
			cellStyle.Width(maxPointsWidth).Render(g.MaxPoints),
			cellStyle.Width(feedbackWidth).Render(g.Feedback),
		)
		fmt.Println(row)
	}

	fmt.Println()
	fmt.Printf("Total: %d grade(s)\n", len(grades))
	return nil
}
