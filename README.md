# NOTES TUI

I made this little app to read my notes from the terminal. This integrates
very well with the workflow im building around Neovim and other terminal based
Applications. It's far more lightweight than using something like Apple Notes,
and I can keep my devices synched by uploading my notes to a Github Repo.

This is also my first project written in Go and also the first time using a
TUI framework such as Bubbletea.

---

## Future Improvements

- Right now the notes directory is hardcoded in main.go, ideally the first
  time a user launches the app, they will be prompted to enter the local directory.
- Add an editing mode so the notes could be altered inside of the app. Perhaps just
  opening the file in the system $EDITOR could be a clean solution.
- Adding theme selection menu, or maybe just working from a config file.
- Commit, push and pull from the app, so that the user could 'refresh' the state
  and also save changes into the repo.

---
