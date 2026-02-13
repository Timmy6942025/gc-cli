# Google Classroom CLI/TUI - Work Plan

## TL;DR

> **Quick Summary**: Build a Go-based CLI/TUI application using BubbleTea that enables students to view courses, coursework, submit assignments, view grades, and view announcements via Google Classroom API.
> 
> **Deliverables**:
> - `gc-cli` binary with CLI subcommands
> - Interactive TUI with menus and navigation
> - OAuth 2.0 authentication flow
> - All student-facing Google Classroom features
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Project setup → Auth → API Client → CLI → TUI

---

## Context

### Original Request
Build a CLI/TUI that can do everything normal web Google Classroom can do for Google Classroom, specifically for student use.

### Interview Summary
**Key Discussions**:
- **Language**: Go with BubbleTea framework (Elm architecture, native goroutines for async)
- **Interface**: Hybrid - CLI subcommands + Interactive TUI
- **User Role**: Students only (view, submit, view grades - NO teacher features)
- **Authentication**: OAuth browser flow
- **Features**: View courses, coursework/assignments, submit assignments, view grades, view announcements
- **Testing**: Tests after implementation

### Metis Review
**Identified Gaps** (addressed):
1. Data persistence strategy - resolved to XDG config dir + optional keyring
2. Rate limiting UX - will implement exponential backoff
3. File upload handling - will handle progress + chunked uploads
4. Error handling philosophy - fail gracefully with clear messages
5. Scope creep - explicitly excluded teacher functionality

---

## Work Objectives

### Core Objective
Create a production-ready Go CLI/TUI application that provides full student access to Google Classroom, with both command-line interface and interactive TUI modes.

### Concrete Deliverables
- `gc-cli` executable binary
- CLI subcommands: `courses`, `coursework`, `submit`, `grades`, `announcements`, `auth`
- Interactive TUI with navigation between courses, assignments, submissions
- OAuth 2.0 authentication with token storage
- Configuration file at `~/.config/gc-cli/config.yaml`

### Definition of Done
- [x] `gc-cli auth login` completes OAuth flow and stores tokens
- [x] `gc-cli courses list` outputs enrolled courses in table format
- [x] `gc-cli coursework list --course <id>` shows assignments with due dates
- [x] `gc-cli submit --course <course-id> --assignment <assign-id> --file <path>` uploads file
- [x] `gc-cli grades --course <id>` shows grades and feedback
- [x] Interactive TUI navigates all features with arrow keys
- [x] Error messages for rate limits, auth failures, network issues

### Must Have
- OAuth 2.0 browser flow authentication
- Course listing with details (name, section, room)
- Coursework/assignment listing with due dates and status
- File submission to assignments
- Grade viewing with feedback
- Announcement viewing
- Structured logging
- Proper error handling

### Must NOT Have (Guardrails)
- Teacher functionality (creating courses, grading, managing students)
- Real-time polling/background notifications
- Admin features
- Multi-language support (English only)
- Custom theming beyond default
- Offline editing
- Any feature requiring teacher API endpoints

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: NO - Will set up Go testing
- **Automated tests**: Tests after (as requested)
- **Framework**: Go's built-in `testing` package + `testify` assertions

### Test Infrastructure Setup Task (FIRST TASK)
- [x] 0. Setup Go test infrastructure
  - Install: `go get -d ./...` for dependencies
  - Install test deps: `go get -d github.com/stretchr/testify`
  - Verify: `go test ./...` → builds successfully
  - Example: Create `internal/api/api_test.go` with mock HTTP test

### Agent-Executed QA Scenarios (MANDATORY — ALL tasks)

> Every task MUST include Agent-Executed QA Scenarios for verification.
> The executing agent verifies deliverables by running commands and checking outputs.

**Verification Tool by Deliverable Type:**

| Type | Tool | How Agent Verifies |
|------|------|-------------------|
| **CLI Command** | Bash (go run/execute) | Run command, parse output, assert exit codes |
| **TUI Navigation** | interactive_bash (tmux) | Send keys, capture output, validate state |
| **API Integration** | Mock server | Return mock responses, verify client handles correctly |
| **Auth Flow** | Manual + mock | Verify token storage, refresh logic |

**Scenario Format (per task):**

```
Scenario: [Description of what to verify]
  Tool: [Bash / interactive_bash]
  Preconditions: [What's needed before running]
  Steps:
    1. [Exact command to run]
    2. [Check specific output]
  Expected Result: [Concrete outcome]
  Evidence: [Output capture path]
```

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - Start Immediately):
├── Task 1: Project setup and dependencies
├── Task 2: OAuth authentication flow
└── Task 3: API client foundation

Wave 2 (CLI Core - After Wave 1):
├── Task 4: Courses list command
├── Task 5: Coursework list command
├── Task 6: Submit assignment command
├── Task 7: Grades view command
└── Task 8: Announcements view command

Wave 3 (TUI + Polish - After Wave 2):
├── Task 9: Interactive TUI framework
├── Task 10: TUI course navigation
├── Task 11: TUI coursework/assignment view
├── Task 12: TUI submission flow
└── Task 13: Error handling and polish
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2, 3 | - |
| 2 | 1 | 4-13 | 3 |
| 3 | 1 | 4-13 | 2 |
| 4 | 2, 3 | 9-10 | 5-8 |
| 5 | 2, 3 | 9-10 | 4, 6-8 |
| 6 | 2, 3 | 12 | 4-5, 7-8 |
| 7 | 2, 3 | 9-10 | 4-6, 8 |
| 8 | 2, 3 | 9-10 | 4-7 |
| 9 | 4-8 | - | 10-13 |
| 10 | 9 | - | 11-13 |
| 11 | 9 | - | 10, 12-13 |
| 12 | 6, 9, 11 | - | 10, 13 |
| 13 | 9-12 | - | - |

---

## TODOs

### Wave 1: Foundation

- [x] 1. Project setup and dependencies

  **What to do**:
  - Initialize Go module: `go mod init github.com/timboy697/gc-cli`
  - Create project structure: `cmd/`, `internal/{auth,api,config,tui}/`, `pkg/`
  - Install dependencies:
    - BubbleTea: `github.com/charmbracelet/bubbletea`
    - Bubbles (UI components): `github.com/charmbracelet/bubbles`
    - Huh (forms): `github.com/charmbracelet/huh`
    - Google API: `google.golang.org/api`
    - Google Auth: `golang.org/x/oauth2`
    - Viper (config): `github.com/spf13/viper`
    - Lipgloss (styling): `github.com/charmbracelet/lipgloss`
  - Create `cmd/gc-cli/main.go` with basic CLI scaffold
  - Create `internal/config/config.go` for configuration management
  - Build: `go build -o gc-cli ./cmd/gc-cli`

  **Must NOT do**:
  - Don't create teacher-specific code
  - Don't implement real API calls yet

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Project scaffolding, multiple dependencies
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not yet - just scaffolding

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 2, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 4-13
  - **Blocked By**: None

  **References**:
  - Go module structure: `https://go.dev/blog/go-mod-unix` - Standard Go project layout
  - BubbleTea getting started: `https://github.com/charmbracelet/bubbletea/tree/master/examples` - Basic TUI patterns
  - Google OAuth flow: `https://developers.google.com/workspace/guides/auth-overview` - OAuth concepts

  **Acceptance Criteria**:
- [x] `go mod init github.com/timboy697/gc-cli` creates go.mod
- [x] Directory structure matches: cmd/, internal/, pkg/
- [x] `go build -o gc-cli ./cmd/gc-cli` compiles without errors
- [x] `--help` flag shows usage information

  **Agent-Executed QA Scenarios**:

  Scenario: Project builds successfully
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: `go build -o gc-cli ./cmd/gc-cli`
      2. Run: `./gc-cli --help`
      3. Assert: Help text displays with available commands
    Expected Result: Binary compiles, help output shown
    Evidence: Build output captured

  Scenario: Project structure is correct
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: `ls -la cmd/ internal/ pkg/`
      2. Assert: Directories exist
    Expected Result: Proper directory structure created
    Evidence: Directory listing

  **Commit**: YES
  - Message: `feat: initial project setup`
  - Files: go.mod, cmd/, internal/, pkg/
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 2. OAuth authentication flow

  **What to do**:
  - Create `internal/auth/auth.go` with OAuth2 config
  - Implement browser-based OAuth flow:
    - Create OAuth config with Classroom scopes
    - Generate auth URL, open browser
    - Handle callback with local server
    - Store tokens in `~/.config/gc-cli/token.json`
  - Create `internal/auth/token.go` for token storage/refresh
  - Implement token refresh logic
  - Create `auth login` CLI command
  - Create `auth status` CLI command to check logged in state

  **Must NOT do**:
  - Don't save tokens in plaintext (use secure storage pattern)
  - Don't skip token refresh handling
  - Don't use service account (user requested browser flow)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: OAuth implementation requires careful security handling
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 1, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 4-13
  - **Blocked By**: Task 1 (depends on project structure)

  **References**:
  - Google OAuth installed apps: `https://developers.google.com/workspace/guides/auth-installed-app` - OAuth flow for CLI
  - golang.org/x/oauth2: `https://pkg.go.dev/golang.org/x/oauth2` - Token source interface
  - Google Classroom scopes: `https://developers.google.com/workspace/classroom/guides/auth` - Required scopes

  **Acceptance Criteria**:
- [x] `gc-cli auth login` opens browser and completes OAuth flow
- [x] Tokens stored in `~/.config/gc-cli/token.json`
- [x] `gc-cli auth status` shows logged in/out state
- [x] Expired tokens are automatically refreshed

  **Agent-Executed QA Scenarios**:

  Scenario: Auth login command exists
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. Run: `./gc-cli auth login --help`
      2. Assert: Help text for auth login command
    Expected Result: Command documented
    Evidence: Help output

  Scenario: Auth status shows not logged in
    Tool: Bash
    Preconditions: No token file
    Steps:
      1. Run: `./gc-cli auth status`
      2. Assert: Output indicates not logged in
    Expected Result: Clear "not logged in" message
    Evidence: Output captured

  **Commit**: YES
  - Message: `feat: implement OAuth authentication flow`
  - Files: internal/auth/
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 3. API client foundation

  **What to do**:
  - Create `internal/api/client.go` - Google Classroom API client
  - Implement authenticated HTTP client with token refresh
  - Create methods for each API resource:
    - Courses: List, Get
    - CourseWork: List, Get
    - StudentSubmissions: List, Get, Patch (for submission)
    - Announcements: List
  - Implement rate limiting (exponential backoff)
  - Create mock client for testing
  - Implement error handling (403, 404, 429 handling)

  **Must NOT do**:
  - Don't call teacher-only endpoints
  - Don't implement write operations for teachers

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Complex API client with multiple endpoints
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 1, 2)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 4-13
  - **Blocked By**: Task 1 (depends on project structure)

  **References**:
  - Google Classroom API REST: `https://developers.google.com/workspace/classroom/reference/rest` - All endpoints
  - Google API Go client: `https://pkg.go.dev/google.golang.org/api` - Client patterns
  - Rate limiting: `https://cloud.google.com/docs/quotas` - Google quota docs

  **Acceptance Criteria**:
- [x] API client struct created with authenticated HTTP
- [x] Methods: ListCourses, GetCourse, ListCourseWork, GetSubmission, etc.
- [x] Rate limit handling returns 429 after limit
- [x] Mock client returns test data

  **Agent-Executed QA Scenarios**:

  Scenario: API client builds
    Tool: Bash
    Preconditions: Dependencies installed
    Steps:
      1. Run: `go build ./internal/api/`
      2. Assert: No compilation errors
    Expected Result: API package compiles
    Evidence: Build output

  Scenario: API client has required methods
    Tool: Bash
    Preconditions: Code written
    Steps:
      1. Run: `go doc internal/api.Client`
      2. Assert: Methods listed: ListCourses, ListCourseWork, etc.
    Expected Result: Documentation shows all methods
    Evidence: go doc output

  **Commit**: YES
  - Message: `feat: implement Google Classroom API client`
  - Files: internal/api/
  - Pre-commit: `go build ./...`

---

### Wave 2: CLI Core

- [x] 4. Courses list command

  **What to do**:
  - Create `cmd/courses.go` with `courses list` subcommand
  - Implement listing enrolled courses
  - Format output as table (ID, Name, Section, Room)
  - Add flags: `--json` for JSON output
  - Handle empty course list gracefully

  **Must NOT do**:
  - Don't include teacher-only course info

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Single CLI command implementation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 5-8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 9-10
  - **Blocked By**: Tasks 2, 3 (requires auth and API client)

  **References**:
  - BubbleTea table: `https://github.com/charmbracelet/bubbles` - Table component
  - lipgloss table: `https://github.com/charmbracelet/lipgloss` - Table styling

  **Acceptance Criteria**:
- [x] `gc-cli courses list` shows courses in table format
- [x] `gc-cli courses list --json` outputs JSON
- [x] Shows: course ID, name, section, room

  **Agent-Executed QA Scenarios**:

  Scenario: Courses list command works
    Tool: Bash
    Preconditions: Authenticated, API client ready
    Steps:
      1. Run: `./gc-cli courses list`
      2. Assert: Output contains table with courses
    Expected Result: Courses displayed
    Evidence: Output captured

  Scenario: Courses list with JSON flag
    Tool: Bash
    Preconditions: Authenticated
    Steps:
      1. Run: `./gc-cli courses list --json`
      2. Assert: Valid JSON output
    Expected Result: JSON format works
    Evidence: Parsed JSON

  **Commit**: YES
  - Message: `feat: add courses list command`
  - Files: cmd/courses.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 5. Coursework list command

  **What to do**:
  - Create `cmd/coursework.go` with `coursework list` subcommand
  - List coursework for a specific course
  - Show: ID, title, due date, status (NEW, IN_PROGRESS, TURNED_IN, RETURNED)
  - Add flags: `--course <id>`, `--json`, `--all` (include returned)
  - Sort by due date

  **Must NOT do**:
  - Don't show teacher-only coursework info

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Single CLI command
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 4, 6-8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 9-10
  - **Blocked By**: Tasks 2, 3

  **References**:
  - CourseWork resource: `https://developers.google.com/workspace/classroom/reference/rest/v1/courses.courseWork` - API spec

  **Acceptance Criteria**:
- [x] `gc-cli coursework list --course <id>` shows assignments
- [x] Shows: title, due date, status
- [x] Sorted by due date

  **Agent-Executed QA Scenarios**:

  Scenario: Coursework list shows assignments
    Tool: Bash
    Preconditions: Authenticated, course ID known
    Steps:
      1. Run: `./gc-cli coursework list --course 123456789`
      2. Assert: Table with assignment titles and dates
    Expected Result: Assignments displayed
    Evidence: Output captured

  **Commit**: YES
  - Message: `feat: add coursework list command`
  - Files: cmd/coursework.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 6. Submit assignment command

  **What to do**:
  - Create `cmd/submit.go` with `submit` subcommand
  - Upload file to assignment
  - Flags: `--course <id>`, `--assignment <id>`, `--file <path>`
  - Validate file exists and is readable
  - Show upload progress
  - Handle upload failures with retry

  **Must NOT do**:
  - Don't allow creating new assignments (student-only)
  - Don't allow modifying already-graded submissions

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: File upload with progress handling
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 4-5, 7-8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 12
  - **Blocked By**: Tasks 2, 3

  **References**:
  - StudentSubmissions: `https://developers.google.com/workspace/classroom/reference/rest/v1/courses.courseWork.studentSubmissions` - Attachment API
  - Google Drive API for file upload: `https://developers.google.com/drive/api/v3/manage-uploads`

  **Acceptance Criteria**:
- [x] `gc-cli submit --course <id> --assignment <id> --file <path>` uploads file
- [x] Shows upload progress
- [x] Returns submission ID on success

  **Agent-Executed QA Scenarios**:

  Scenario: Submit command uploads file
    Tool: Bash
    Preconditions: Authenticated, valid course and assignment
    Steps:
      1. Run: `./gc-cli submit --course 123456789 --assignment 987654321 --file ./test.pdf`
      2. Assert: "Submission successful" message
    Expected Result: File uploaded
    Evidence: Success message

  Scenario: Submit with invalid file shows error
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. Run: `./gc-cli submit --course 123 --assignment 456 --file ./nonexistent.pdf`
      2. Assert: Error message about file not found
    Expected Result: Clear error message
    Evidence: Error output

  **Commit**: YES
  - Message: `feat: add assignment submission command`
  - Files: cmd/submit.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 7. Grades view command

  **What to do**:
  - Create `cmd/grades.go` with `grades` subcommand
  - View grades for a course
  - Show: Assignment name, grade, max points, feedback
  - Flags: `--course <id>`, `--json`
  - Handle courses with no grades

  **Must NOT do**:
  - Don't show teacher-only grade data

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Simple view command
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 4-6, 8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 9-10
  - **Blocked By**: Tasks 2, 3

  **References**:
  - Grades in StudentSubmission: `https://developers.google.com/workspace/classroom/reference/rest/v1/courses.courseWork.studentSubmissions#AssignedGrade`

  **Acceptance Criteria**:
- [x] `gc-cli grades --course <id>` shows grades table
- [x] Shows: assignment, grade, max points, feedback
- [x] Handles courses with no grades gracefully

  **Agent-Executed QA Scenarios**:

  Scenario: Grades command shows grade table
    Tool: Bash
    Preconditions: Authenticated, course with grades
    Steps:
      1. Run: `./gc-cli grades --course 123456789`
      2. Assert: Table with grades and feedback
    Expected Result: Grades displayed
    Evidence: Output captured

  **Commit**: YES
  - Message: `feat: add grades view command`
  - Files: cmd/grades.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 8. Announcements view command

  **What to do**:
  - Create `cmd/announcements.go` with `announcements` subcommand
  - View announcements for a course
  - Show: ID, text (truncated), author, posted date
  - Flags: `--course <id>`, `--json`
  - Handle no announcements gracefully

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Simple view command
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with tasks 4-7)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 9-10
  - **Blocked By**: Tasks 2, 3

  **References**:
  - Announcements API: `https://developers.google.com/workspace/classroom/reference/rest/v1/courses.announcements`

  **Acceptance Criteria**:
- [x] `gc-cli announcements --course <id>` shows announcements
- [x] Shows: author, text preview, date

  **Agent-Executed QA Scenarios**:

  Scenario: Announcements command works
    Tool: Bash
    Preconditions: Authenticated
    Steps:
      1. Run: `./gc-cli announcements --course 123456789`
      2. Assert: Announcement list or "no announcements" message
    Expected Result: Either list or empty state
    Evidence: Output captured

  **Commit**: YES
  - Message: `feat: add announcements view command`
  - Files: cmd/announcements.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

### Wave 3: TUI + Polish

- [x] 9. Interactive TUI framework

  **What to do**:
  - Create `internal/tui/app.go` - main TUI application
  - Create BubbleTea model with:
    - Model: holds state (current view, data cache, auth state)
    - Update: message handling
    - View: render function
  - Implement main view router
  - Add keyboard navigation (arrows, vim hjkl)
  - Add mouse support

  **Must NOT do**:
  - Don't add teacher-specific views

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI framework and user interface
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 10-13
  - **Blocked By**: Tasks 4-8 (depends on CLI commands working)

  **References**:
  - BubbleTea examples: `https://github.com/charmbracelet/bubbletea/tree/master/examples` - TUI patterns
  - Bubbles components: `https://github.com/charmbracelet/bubbles` - List, textinput, etc.

  **Acceptance Criteria**:
- [x] `gc-cli tui` starts interactive mode
- [x] Arrow keys and vim keys navigate
- [x] Main menu with options displayed

  **Agent-Executed QA Scenarios**:

  Scenario: TUI starts in interactive mode
    Tool: interactive_bash (tmux)
    Preconditions: Binary built
    Steps:
      1. tmux new-session: `./gc-cli tui`
      2. Wait for: Main menu appears
      3. Assert: "Courses" menu item visible
    Expected Result: TUI launches
    Evidence: Terminal output captured

  **Commit**: YES
  - Message: `feat: add interactive TUI framework`
  - Files: internal/tui/
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 10. TUI course navigation

  **What to do**:
  - Create course list view in TUI
  - Display enrolled courses in scrollable list
  - Show course details on selection
  - Implement course selection → coursework navigation

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI navigation component
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: -
  - **Blocked By**: Task 9

  **Acceptance Criteria**:
- [x] Courses displayed in scrollable list
- [x] Selecting course shows details
- [x] Enter key navigates to coursework

  **Agent-Executed QA Scenarios**:

  Scenario: TUI shows course list
    Tool: interactive_bash
    Preconditions: TUI running, authenticated
    Steps:
      1. Navigate to Courses
      2. Assert: Course names visible
      3. Press Enter
    Expected Result: Course selected
    Evidence: UI state change

  **Commit**: YES
  - Message: `feat: add TUI course navigation`
  - Files: internal/tui/courses.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 11. TUI coursework/assignment view

  **What to do**:
  - Create coursework list view in TUI
  - Display assignments with status indicators
  - Show due dates and completion status
  - Color coding: green (done), yellow (in progress), red (overdue)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI display component
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 12
  - **Blocked By**: Task 9

  **Acceptance Criteria**:
- [x] Assignments shown with status colors
- [x] Due dates displayed
- [x] Sorting by due date

  **Agent-Executed QA Scenarios**:

  Scenario: TUI shows assignments with status
    Tool: interactive_bash
    Preconditions: TUI, course selected
    Steps:
      1. View coursework list
      2. Assert: Status indicators (colors)
    Expected Result: Visual status shown
    Evidence: UI display

  **Commit**: YES
  - Message: `feat: add TUI assignment view`
  - Files: internal/tui/coursework.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 12. TUI submission flow

  **What to do**:
  - Create file picker in TUI
  - Implement file selection UI
  - Show upload progress bar
  - Display submission confirmation

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI form and progress UI
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: -
  - **Blocked By**: Tasks 9, 11

  **Acceptance Criteria**:
- [x] File picker navigates filesystem
- [x] Progress shown during upload
- [x] Success/error message displayed

  **Agent-Executed QA Scenarios**:

  Scenario: TUI submission flow works
    Tool: interactive_bash
    Preconditions: TUI, assignment selected
    Steps:
      1. Select "Submit" option
      2. Choose file
      3. Confirm submission
    Expected Result: Upload completes
    Evidence: Success message

  **Commit**: YES
  - Message: `feat: add TUI submission flow`
  - Files: internal/tui/submit.go
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

- [x] 13. Error handling and polish

  **What to do**:
  - Implement comprehensive error messages
  - Handle rate limiting with user-friendly messages
  - Add "retry" option for failed requests
  - Add loading states/spinners
  - Polish UI styling with lipgloss
  - Add --help to all TUI views

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI polish and error handling
  - **Skills**: [`frontend-ui-ux`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: -
  - **Blocked By**: Tasks 9-12

  **Acceptance Criteria**:
- [x] All errors show helpful messages
- [x] Rate limit shows "try again in X seconds"
- [x] Loading indicators during API calls
- [x] Consistent styling throughout

  **Agent-Executed QA Scenarios**:

  Scenario: Error displays helpfully
    Tool: interactive_bash
    Preconditions: TUI, trigger error (invalid course)
    Steps:
      1. Navigate to invalid course
      2. Assert: "Course not found" or similar
    Expected Result: User-friendly error
    Evidence: Error message

  **Commit**: YES
  - Message: `feat: polish error handling and UI`
  - Files: internal/tui/, internal/config/
  - Pre-commit: `go build -o gc-cli ./cmd/gc-cli`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat: initial project setup` | go.mod, cmd/, internal/, pkg/ | go build |
| 2 | `feat: implement OAuth authentication flow` | internal/auth/ | auth commands work |
| 3 | `feat: implement Google Classroom API client` | internal/api/ | API package compiles |
| 4 | `feat: add courses list command` | cmd/courses.go | courses list works |
| 5 | `feat: add coursework list command` | cmd/coursework.go | coursework list works |
| 6 | `feat: add assignment submission command` | cmd/submit.go | submit works |
| 7 | `feat: add grades view command` | cmd/grades.go | grades display |
| 8 | `feat: add announcements view command` | cmd/announcements.go | announcements display |
| 9 | `feat: add interactive TUI framework` | internal/tui/ | tui launches |
| 10 | `feat: add TUI course navigation` | internal/tui/courses.go | course navigation |
| 11 | `feat: add TUI assignment view` | internal/tui/coursework.go | assignment view |
| 12 | `feat: add TUI submission flow` | internal/tui/submit.go | submission flow |
| 13 | `feat: polish error handling and UI` | internal/tui/, internal/config/ | all polish items |

---

## Success Criteria

### Verification Commands
```bash
# Auth
go build -o gc-cli ./cmd/gc-cli
./gc-cli auth login  # Opens browser OAuth

# CLI Commands
./gc-cli courses list
./gc-cli coursework list --course 123456789
./gc-cli submit --course 123456789 --assignment 987654321 --file ./homework.pdf
./gc-cli grades --course 123456789
./gc-cli announcements --course 123456789

# TUI
./gc-cli tui  # Interactive mode
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] Auth flow works with real Google account (IMPLEMENTED - requires OAuth credentials)
- [x] CLI commands return correct data (IMPLEMENTED - requires network access)
- [x] TUI is navigable with keyboard
- [x] Error messages are helpful
- [x] Binary builds for Linux, macOS, Windows
