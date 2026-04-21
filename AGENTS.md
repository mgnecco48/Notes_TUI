# Notes_TUI

- Single-module Go app; the executable entrypoint is `main.go` at the repo root.
- UI stack: Bubble Tea + Bubbles (`list`, `spinner`, `textinput`, `viewport`), styling via Lip Gloss, Markdown rendering via Glamour.
- The notes directory is persisted in `os.UserConfigDir()/notes_tui/path.txt`; first launch prompts for a full local path if that file is missing.
- `~` expansion is not supported for the notes path. Use an absolute path.
- `q` / `Ctrl+C` quit the app; when viewing a note, `q` exits note view first instead of quitting immediately.
- Run checks from the repo root with `go test ./...` and `go build ./...`.
- There is no repo-local task runner, CI config, or generated-code workflow in this repo.
- Keep changes ASCII unless a file already clearly uses non-ASCII.
