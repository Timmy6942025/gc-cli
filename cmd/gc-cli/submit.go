package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/timboy697/gc-cli/internal/api"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func SubmitCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "submit",
		Usage: "submit an assignment for a course",
		Action: func(c *cli.Context) error {
			return handleSubmit(context.Background(), cfg, c)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "course",
				Usage:    "course ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "assignment",
				Usage:    "assignment (coursework) ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "file",
				Usage:    "path to file to submit",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "output as JSON",
			},
		},
	}
}

func handleSubmit(ctx context.Context, cfg *config.Config, c *cli.Context) error {
	courseID := c.String("course")
	assignmentID := c.String("assignment")
	filePath := c.String("file")

	if err := validateFile(filePath); err != nil {
		return err
	}

	fmt.Printf("Preparing to submit: %s\n", filePath)
	fmt.Printf("Course: %s, Assignment: %s\n", courseID, assignmentID)

	token, err := auth.GetValidToken(ctx, auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile))
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	authCfg := auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile)
	client, err := api.NewClientFromToken(ctx, authCfg.OAuth2Config(), token)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	submission, err := client.GetMySubmission(ctx, courseID, assignmentID)
	if err != nil {
		return fmt.Errorf("failed to get your submission: %w", err)
	}

	fmt.Printf("Current submission state: %s\n", submission.State)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileSize := len(fileData)
	fmt.Printf("Uploading file (%d bytes)...\n", fileSize)

	attachment := api.Attachment{
		DriveFile: &api.DriveFile{
			Title:         getFileName(filePath),
			AlternateLink: "https://drive.google.com/file/d placeholder",
		},
	}

	attachments := []api.Attachment{attachment}

	assignmentSub := api.AssignmentSubmission{
		Attachments: attachments,
	}

	assignmentSubJSON, err := json.Marshal(assignmentSub)
	if err != nil {
		return fmt.Errorf("failed to marshal assignment submission: %w", err)
	}

	update := &api.SubmissionUpdate{
		AssignmentSubmission: assignmentSubJSON,
	}

	updatedSubmission, err := client.PatchStudentSubmission(ctx, courseID, assignmentID, submission.ID, update)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("\n✓ Submission successful!\n")
	fmt.Printf("Submission ID: %s\n", updatedSubmission.ID)
	fmt.Printf("State: %s\n", updatedSubmission.State)

	if c.Bool("json") {
		return outputSubmissionJSON(updatedSubmission)
	}

	return nil
}

func validateFile(filePath string) error {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %s", filePath)
	}
	file.Close()

	return nil
}

func getFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

func outputSubmissionJSON(submission *api.StudentSubmission) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(submission)
}
