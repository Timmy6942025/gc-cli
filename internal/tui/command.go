package tui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type commandResultMsg struct {
	command string
	output  string
	err     error
}

func (m *model) defaultCommandTemplate() string {
	course := m.selectedCourse()
	if m.classMode && course != nil {
		switch m.currentClassTab() {
		case "Stream":
			return fmt.Sprintf("stream list --course %s", course.Id)
		case "Classwork":
			return fmt.Sprintf("classwork list --course %s", course.Id)
		case "People":
			return fmt.Sprintf("people list --course %s --role all", course.Id)
		case "Grades":
			return fmt.Sprintf("grades list --course %s", course.Id)
		}
	}
	if m.currentGlobalView() == "To-do" {
		return "to-do list"
	}
	return "classes list"
}

func (m *model) runCLICommand(raw string) tea.Cmd {
	command := strings.TrimSpace(raw)
	return func() tea.Msg {
		args, err := splitCommandLine(command)
		if err != nil {
			return commandResultMsg{command: command, err: err}
		}
		if len(args) == 0 {
			return commandResultMsg{command: command, err: fmt.Errorf("empty command")}
		}

		first := strings.TrimSpace(args[0])
		if first == "gc-cli" || first == "gc" {
			args = args[1:]
		}
		if len(args) == 0 {
			return commandResultMsg{
				command: command,
				err:     fmt.Errorf("missing subcommand; example: classwork list --course <id>"),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, os.Args[0], args...)
		cmd.Env = os.Environ()
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		runErr := cmd.Run()
		output := strings.TrimSpace(out.String())
		if output == "" {
			output = "(no output)"
		}

		if ctx.Err() == context.DeadlineExceeded {
			return commandResultMsg{
				command: command,
				output:  output,
				err:     fmt.Errorf("command timed out after 120s"),
			}
		}
		if runErr != nil {
			return commandResultMsg{command: command, output: output, err: runErr}
		}
		return commandResultMsg{command: command, output: output}
	}
}

func splitCommandLine(raw string) ([]string, error) {
	var parts []string
	var current strings.Builder

	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			parts = append(parts, current.String())
			current.Reset()
		}
	}

	for _, r := range raw {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			if inSingle {
				current.WriteRune(r)
				continue
			}
			escaped = true
		case '\'':
			if inDouble {
				current.WriteRune(r)
				continue
			}
			inSingle = !inSingle
		case '"':
			if inSingle {
				current.WriteRune(r)
				continue
			}
			inDouble = !inDouble
		case ' ', '\t', '\n':
			if inSingle || inDouble {
				current.WriteRune(r)
				continue
			}
			flush()
		default:
			current.WriteRune(r)
		}
	}

	if escaped {
		return nil, fmt.Errorf("unfinished escape sequence")
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unclosed quote in command")
	}
	flush()
	return parts, nil
}
