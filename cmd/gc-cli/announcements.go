package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/api"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func AnnouncementsCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "announcements",
		Usage: "list announcements for a course",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "course",
				Usage:    "course ID to fetch announcements from",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "output as JSON",
			},
		},
		Action: handleAnnouncements(cfg),
	}
}

func handleAnnouncements(cfg *config.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		ctx := context.Background()

		courseID := c.String("course")
		if courseID == "" {
			return fmt.Errorf("course ID is required (use --course flag)")
		}

		token, err := auth.GetValidToken(ctx, auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile))
		if err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		authCfg := auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile)
		client, err := api.NewClientFromToken(ctx, authCfg.OAuth2Config(), token)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		announcements, _, err := client.ListAnnouncements(ctx, courseID, 100)
		if err != nil {
			return fmt.Errorf("failed to list announcements: %w", err)
		}

		if c.Bool("json") {
			return outputAnnouncementsJSON(announcements)
		}
		return outputAnnouncementsTable(announcements)
	}
}

func outputAnnouncementsJSON(announcements []api.Announcement) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(announcements)
}

func outputAnnouncementsTable(announcements []api.Announcement) error {
	if len(announcements) == 0 {
		fmt.Println("No announcements")
		return nil
	}

	idWidth := 12
	textWidth := 50
	authorWidth := 15
	dateWidth := 20

	for _, a := range announcements {
		if len(a.ID) > idWidth {
			idWidth = len(a.ID)
		}
		textLen := len(strings.TrimSpace(stripHTML(a.Text)))
		if textLen > textWidth {
			textWidth = textLen
		}
		authorLen := len(a.CreatorUserID)
		if authorLen > authorWidth {
			authorWidth = authorLen
		}
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(idWidth).Render("ID"),
		headerStyle.Width(textWidth).Render("Text"),
		headerStyle.Width(authorWidth).Render("Author"),
		headerStyle.Width(dateWidth).Render("Posted Date"),
	)
	separator := separatorStyle.Render("─")

	fmt.Println(header)
	fmt.Println(lipgloss.JoinHorizontal(
		lipgloss.Left,
		separator+separator+separator+separator,
	))

	for _, a := range announcements {
		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cellStyle.Width(idWidth).Render(truncate(a.ID, idWidth)),
			cellStyle.Width(textWidth).Render(truncate(strings.TrimSpace(stripHTML(a.Text)), textWidth)),
			cellStyle.Width(authorWidth).Render(truncate(a.CreatorUserID, authorWidth)),
			cellStyle.Width(dateWidth).Render(a.CreationTime.Format("2006-01-02 15:04")),
		)
		fmt.Println(row)
	}

	fmt.Println()
	fmt.Printf("Total: %d announcement(s)\n", len(announcements))
	return nil
}

func stripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", " ")
	s = strings.ReplaceAll(s, "<br/>", " ")
	s = strings.ReplaceAll(s, "<br />", " ")
	s = strings.ReplaceAll(s, "<p>", " ")
	s = strings.ReplaceAll(s, "</p>", " ")
	s = strings.ReplaceAll(s, "<li>", " - ")
	s = strings.ReplaceAll(s, "</li>", " ")
	s = strings.ReplaceAll(s, "<ul>", " ")
	s = strings.ReplaceAll(s, "</ul>", " ")
	s = strings.ReplaceAll(s, "<b>", "")
	s = strings.ReplaceAll(s, "</b>", "")
	s = strings.ReplaceAll(s, "<i>", "")
	s = strings.ReplaceAll(s, "</i>", "")
	s = strings.ReplaceAll(s, "<a href=\"", "")
	s = strings.ReplaceAll(s, "</a>", "")
	inTag := false
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result = append(result, r)
		}
	}
	return strings.TrimSpace(string(result))
}
