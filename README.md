# gc-cli

Google Classroom CLI/TUI built in Go, packaged for npm/bun so users can run `gc-cli` (or `gc`) directly.

## Install

### npm

```bash
npm install -g gc-classroom-cli
```

### bun

```bash
bun add -g gc-classroom-cli
```

Then run:

```bash
gc-cli auth login
gc-cli
```

`gc` is also installed as an alias.

## TUI navigation

In the TUI:

```bash
tab / shift+tab   # move focus between Global Views, Classes, Class Tabs, Content
up/down (or j/k)  # move inside the focused pane
[ / ]             # quick switch global view (or class tab when class is open)
:                 # open in-TUI command bar (run any gc-cli subcommand)
enter             # open class tabs for selected class
esc               # return from class tabs to global mode
r                 # refresh from Classroom
o                 # open current context in browser
```

`Stream`, `Classwork`, `People`, `Grades`, and `To-do` load real Classroom API data in the content pane.
Use command bar examples:

```bash
stream post --course <course_id> --text "Reminder: quiz Friday"
classwork create --course <course_id> --title "Worksheet 4"
people invite --course <course_id> --role student --user student@example.com
submissions turn-in --course <course_id> --course-work <work_id> --submission <submission_id>
```

## Seamless OAuth

The binary ships with a default OAuth Desktop client configuration, so users can run `gc-cli auth login` without manual config.

Optional overrides:

```bash
export GC_OAUTH_CLIENT_ID="..."
export GC_OAUTH_CLIENT_SECRET="..."
```

## Command surface

```bash
gc-cli                          # launch TUI
gc-cli auth login|status|logout

gc-cli classes list|show|create|update|archive|restore|delete
gc-cli stream list|post|edit|delete
gc-cli classwork list|create|edit|publish|schedule|delete
gc-cli submissions list|show|turn-in|unsubmit|grade|return|reclaim
gc-cli people list|invite|remove
gc-cli grades list|set-draft|set-assigned|return
gc-cli topics list|create|edit|delete|move
gc-cli to-do list
gc-cli calendar open
gc-cli handoff open <feature> --course <id>
```

Use `--json` with any command for machine-readable output.

## Package maintainers

Build npm dist binaries:

```bash
npm run build:dist
```

If your OAuth Desktop client requires a secret for token exchange, inject it at build time:

```bash
GC_CLI_DEFAULT_OAUTH_CLIENT_SECRET="..." npm run build:dist
```

Create package tarball:

```bash
npm pack
```

Publish (after npm login):

```bash
npm publish --access public
```

## Development

```bash
go test ./...
```

See parity details in [`docs/parity-matrix.md`](docs/parity-matrix.md).
