# NOTES TUI

I made this little app to read my notes from the terminal. This integrates
very well with the workflow im building around Neovim and other terminal based
Applications. It's far more lightweight than using something like Apple Notes,
and I can keep my devices synched by uploading my notes to a Github Repo.

This is also my first project written in Go and also the first time using a
TUI framework such as Bubbletea.

---

## Testing The Program

The `main` branch supports viewing and editing notes. The editing feature is
new, so it is still worth testing with notes that are backed up or tracked in
Git.

> Currently, the colorscheme is hardcoded and designed to work on a dark
> terminal theme. The TUI components do not render a solid background, so the app
> relies on the terminal background color. If you are using a light terminal
> theme, the colors may not look as intended. I plan to add theme selection in
> the future.

If you have Go installed, you can run the app from the project root:

```bash
go run .
```

You can also test one of the prebuilt binaries from the `dist` directory.
Most Ubuntu computers use the `amd64` build:

```bash
chmod +x dist/notes-tui-linux-amd64
./dist/notes-tui-linux-amd64
```

For ARM-based Ubuntu machines, such as some Raspberry Pi setups, use the
`arm64` build instead:

```bash
chmod +x dist/notes-tui-linux-arm64
./dist/notes-tui-linux-arm64
```

If you are not sure which one to use, check the machine architecture:

```bash
uname -m
```

Use `notes-tui-linux-amd64` for `x86_64`, and use `notes-tui-linux-arm64` for
`aarch64`.

On Apple Silicon Macs, use the macOS ARM build:

```bash
chmod +x dist/notes-tui-darwin-arm64
./dist/notes-tui-darwin-arm64
```

On Windows, use the `.exe` build from PowerShell or Command Prompt:

```powershell
.\dist\notes-tui-windows-amd64.exe
```

On first launch, the app asks for the full local path to your notes directory.
Use an absolute path like `/home/yourname/notes` or `C:\Users\yourname\Notes`.
Do not use `~` because the app does not expand it yet.

---

## Future Improvements

- Adding theme selection menu, or maybe just working from a config file.
- Commit, push and pull from the app, so that the user could 'refresh' the state
  and also save changes into the repo.
