# Typa

Terminal typing game with a public leaderboard. You move on an endless grid by typing, grow a trail, collect tokens, and avoid crashing into yourself under a fixed round clock with gross WPM scoring.

**Included in this repository**


| Component         | Stack                                                                                                                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Game              | Go, [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), [Harmonica](https://github.com/charmbracelet/harmonica) (camera) |
| Leaderboard & API | Next.js (`web/`), deployable to Vercel                                                                                                                                                |


---

## Features

- **Rounds** — Two-minute sessions with gross WPM and score tracking  
- **Movement** — Type to advance; turn with arrow keys or directional words (no 180° reversals)  
- **Collectibles** — Pick up **◎** for points  
- **Leaderboard** — Optional sync to your hosted API (best score per anonymized IP when configured)

---

## Run the game

Requires [Go](https://go.dev/).

```bash
go run .
```

