# gc-cli

Google Classroom CLI built in Go, packaged for npm/bun so users can run `gc-cli` (or `gc`) directly.

## Install

### npm

```bash
npm install -g google-classroom-cli
```

### bun

```bash
bun add -g google-classroom-cli
```

Then run:

```bash
gc-cli auth login
gc-cli classes list
```

`gc` is also installed as an alias.

## Seamless OAuth

The binary ships with a default OAuth Desktop client configuration, so users can run `gc-cli auth login` without manual config.

Optional overrides:

```bash
export GC_OAUTH_CLIENT_ID="..."
export GC_OAUTH_CLIENT_SECRET="..."
```

## Command surface

```bash
gc-cli                          # show command help
gc-cli auth login|status|logout

gc-cli classes|courses list|show|get|create|add|update|edit|archive|restore|delete|rm
gc-cli stream|announcements list|ls|post|add|edit|update|delete|rm
gc-cli classwork|coursework|cw list|ls|create|add|edit|update|publish|schedule|delete|rm
gc-cli submissions|subs list|ls|show|get|turn-in|submit|unsubmit|grade|return|reclaim
gc-cli people|roster list|ls|invite|add|remove|rm
gc-cli grades|gradebook list|ls|set-draft|draft|set-assigned|assigned|return
gc-cli topics|topic list|ls|create|add|edit|update|delete|rm|move
gc-cli to-do|todo [list]
gc-cli calendar open
gc-cli handoff|web open <feature> [course_id]
```

Use `--json` with any command for machine-readable output.

Positional ID shortcuts are supported for common commands (flags still work):

```bash
gc-cli classes show <course_id>
gc-cli classwork publish <course_id> <course_work_id>
gc-cli submissions turn-in <course_id> <course_work_id> <submission_id>
gc-cli people invite <course_id> <user_id_or_email> --role student
```

Common short flags for quicker usage:

```bash
-c  # --course
-w  # --course-work
-s  # --submission
-u  # --user
```

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
