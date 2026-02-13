# gc-cli

Google Classroom CLI for students - manage your courses, coursework, and grades from the terminal.

## Features

- List and view courses
- List coursework and assignments
- View grades
- View announcements
- Submit assignments
- Interactive TUI mode

## Installation

```bash
go install github.com/timboy697/gc-cli@latest
```

## Quick Start

```bash
# Authenticate with your Google account
gc-cli auth login

# List your courses
gc-cli courses list
```

That's it! No configuration needed - credentials are built-in.

## Usage

```bash
# Check auth status
gc-cli auth status

# List coursework for a course
gc-cli coursework list --course COURSE_ID

# List grades
gc-cli grades list --course COURSE_ID

# List announcements
gc-cli announcements list --course COURSE_ID

# Submit an assignment
gc-cli submit --course COURSE_ID --coursework COURSEWORK_ID --file submission.pdf

# Launch interactive TUI
gc-cli tui
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with Google |
| `auth status` | Check authentication status |
| `courses list` | List all enrolled courses |
| `coursework list` | List coursework for a course |
| `grades list` | List grades for a course |
| `announcements list` | List announcements for a course |
| `submit` | Submit an assignment |
| `tui` | Launch interactive TUI |

## Configuration (Optional)

The CLI works out of the box without any configuration. If you need to customize, create `~/.config/gc-cli/config.yaml`:

```yaml
auth:
  token_file: ~/.config/gc-cli/token.json

google_classroom:
  course_id: optional-default-course-id
```

Default config path: `~/.config/gc-cli/config.yaml`

## Development

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run locally
go run ./cmd/gc-cli
```
