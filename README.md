# gc-cli

Google Classroom CLI for students - manage your classes, classwork, and grades from the terminal.

## Features

- List and view classes
- List classwork and assignments
- View grades
- View announcements
- Submit assignments
- Interactive TUI mode

## Installation

```bash
go install github.com/timboy697/gc-cli@latest
```

## Setup

### 1. Configure OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project (or select existing)
3. Enable the Google Classroom API
4. Go to Credentials → Create OAuth Client ID
5. Set application type to "Desktop app"
6. Download the credentials JSON
7. Copy `client_id` and `client_secret` to your config

### 2. Create Config File

Create `~/.config/gc-cli/config.yaml`:

```yaml
auth:
  client_id: YOUR_CLIENT_ID
  client_secret: YOUR_CLIENT_SECRET
  token_file: ~/.config/gc-cli/token.json

google_classroom:
  course_id: optional-default-course-id
```

### 3. Authenticate

```bash
gc-cli auth login
```

This opens a browser window for Google sign-in.

## Usage

```bash
# Check auth status
gc-cli auth status

# List your classes
gc-cli classes list

# List classwork for a class
gc-cli classwork list --course COURSE_ID

# List grades
gc-cli grades list --course COURSE_ID

# List announcements
gc-cli announcements list --course COURSE_ID

# Submit an assignment
gc-cli submit --course COURSE_ID --assignment ASSIGNMENT_ID --file submission.pdf

# Launch interactive TUI
gc-cli tui
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with Google |
| `auth status` | Check authentication status |
| `classes list` | List all enrolled classes |
| `classwork list` | List classwork for a class |
| `grades list` | List grades for a class |
| `announcements list` | List announcements for a class |
| `submit` | Submit an assignment |
| `tui` | Launch interactive TUI |

## Configuration

Default config path: `~/.config/gc-cli/config.yaml`

| Option | Description |
|--------|-------------|
| `auth.client_id` | Google OAuth client ID |
| `auth.client_secret` | Google OAuth client secret |
| `auth.token_file` | Path to store auth token |
| `google_classroom.course_id` | Default course ID |

## Development

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run locally
go run ./cmd/gc-cli
```
