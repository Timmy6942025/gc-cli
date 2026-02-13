# Google Classroom CLI - Learnings

## Project Overview
- Go-based CLI/TUI using BubbleTea framework
- Student-only access to Google Classroom API
- OAuth 2.0 browser flow authentication
- Hybrid CLI + TUI interface

## Key Patterns
- Foundation → CLI Core Wave-based execution: → TUI
- Dependencies: Task 1 (setup) → Tasks 2,3 (auth, api) → Tasks 4-8 (CLI) → Tasks 9-13 (TUI)

## Conventions
- Project structure: cmd/, internal/{auth,api,config,tui}/, pkg/
- Configuration: ~/.config/gc-cli/config.yaml
- Token storage: ~/.config/gc-cli/token.json

## Task Completion Summary
- Task 0: Test infrastructure (go.mod created)
- Task 1: Project setup ✅
- Task 2: OAuth authentication ✅
- Task 3: API client ✅
- Task 4-8: CLI commands ✅
- Task 9-13: TUI framework ✅

## Debugging/Refinement Notes
- Fixed missing `fmt` import in cmd/gc-cli/main.go
- Fixed TUI struct field name conflicts (Description -> Desc, Title -> AssignTitle/AnnounceTitle)
- Fixed bubble tea API compatibility issues:
  - Removed listStyles() function (incompatible with v0.16.1)
  - Replaced SelectPrevious/SelectNext with CursorUp/CursorDown
  - Removed viewport.Style field assignment (incompatible)
  - Removed list.DefaultDelegate() argument (incompatible)
  - Removed keys.ScrollUp/ScrollDown and viewport scrolling methods
  - Removed tea.MouseScrollUp/tea.MouseScrollDown (incompatible)
  - Removed tea.WithMouseCellFader (incompatible with v0.24.2)
- All Go files pass go fmt and go vet

## Notes
- go.mod requires `go mod tidy` on user machine with network
- OAuth requires Google Cloud credentials configuration