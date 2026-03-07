### 🛡️ Guardrail Design Principles

Before the plan, here are the **4 core rules** the entire app is built around:

| Rule                     | Description                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------ |
| **Single Executor**      | Only 1 `ddcutil` process allowed at a time, globally. A mutex-guarded queue.         |
| **Debounce on Action**   | Brightness changes only fire after the user **stops** input for `300ms`              |
| **Commit-on-Confirm**    | Slider and input mode **never** call ddcutil until `Enter` is pressed                |
| **Max Concurrent Procs** | Hard cap of `maxProcs = 1` (configurable to max 2) in config, enforced via semaphore |

---

### ⚙️ `config.toml`

```toml
[steps]
small  = 1
medium = 5
large  = 10

[display]
refresh_interval_ms = 5000   # slower = safer, 0 = disabled
show_display_name   = true

[guardrails]
debounce_ms       = 300      # wait after last keypress before firing ddcutil
max_ddcutil_procs = 1        # hard cap, max 2 allowed
command_timeout_s = 8        # kill ddcutil if it hangs beyond this

[theme]
accent_color = "#A78BFA"
```

---

### 🗂️ Directory Structure

```
luma/
├── main.go
├── go.mod
├── go.sum
├── config.toml.example
├── README.md
├── internal/
│   ├── ddc/
│   │   ├── client.go        ← ddcutil shell wrapper
│   │   ├── executor.go      ← single-process queue + semaphore + timeout
│   │   └── client_test.go
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   └── ui/
│       ├── model.go
│       ├── update.go
│       ├── view.go
│       ├── styles.go
│       └── keys.go
└── .github/
    └── workflows/
        └── ci.yml
```

> **New file: `executor.go`** — this is the heart of the guardrail system.

---

### 🔁 Git Iterations

---

#### **Commit 1 — `chore: init project scaffold`**

> _Goal: Runnable empty shell, nothing crashes_

- `go mod init github.com/yourname/luma`
- Add deps: `bubbletea`, `lipgloss`, `bubbles`, `BurntSushi/toml`
- `main.go` with bare `tea.NewProgram` rendering `"luma loading..."`
- `README.md` stub + `.gitignore`

**✅ Deliverable:** `go run .` starts and exits cleanly with `q`

---

#### **Commit 2 — `feat(ddc): implement ddcutil executor with guardrails`**

> _Goal: A safe, controlled execution layer that prevents ddcutil abuse_

- `internal/ddc/executor.go`:
  - `Executor` struct with:
    - `semaphore chan struct{}` — sized to `maxProcs` (from config, max 2)
    - `mu sync.Mutex` — prevents concurrent queue manipulation
    - `inFlight int` — tracks active processes
  - `Run(ctx context.Context, args ...string) (string, error)`:
    - Tries to acquire semaphore, **drops the command silently** if at cap (no queue buildup)
    - Wraps execution in `context.WithTimeout(commandTimeoutS)`
    - Kills the process if timeout exceeded
    - Releases semaphore in `defer`
  - `IsBusy() bool` — UI can check this to show a lock indicator
- All ddcutil calls in the entire app go through this executor, **never** `exec.Command` directly

```
// Guardrail flow:
keystroke → debounce timer → executor.Run() → ddcutil
                                ↓ if busy:
                            drop silently + show ⚠ busy indicator
```

**✅ Deliverable:** `go test ./internal/ddc/...` — concurrent calls proven safe, timeout proven to kill hung processes

---

#### **Commit 3 — `feat(ddc): implement ddcutil client using executor`**

> _Goal: High-level display API built on top of the safe executor_

- `internal/ddc/client.go`:
  - `Detect(ctx) ([]Display, error)` — parses `ddcutil detect`
  - `GetBrightness(ctx, displayIndex) (current, max int, error)` — parses `getvcp 10`
  - `SetBrightness(ctx, displayIndex, value int) error` — calls `setvcp 10 <val>`
  - `SetBrightnessAll(ctx, displays []Display, value int) error` — sequential loop, **not concurrent** (one at a time through executor)
  - `Display` struct: `{ Index, Name, Model, Brightness, MaxVal int }`
- Every method passes `ctx` down to `executor.Run()` for timeout propagation
- Parse errors return typed errors, not raw strings

**✅ Deliverable:** Client methods work correctly, executor enforces the single-process rule end-to-end

---

#### **Commit 4 — `feat(config): implement config loader with guardrail defaults`**

> _Goal: User preferences load, guardrail values have safe defaults_

- `internal/config/config.go`:
  - `LoadOrDefault()` — reads `~/.config/luma/config.toml`, falls back to safe defaults
  - Validates guardrail values on load:
    - `max_ddcutil_procs` clamped to `[1, 2]` — never allow > 2
    - `debounce_ms` clamped to `[100, 2000]` — never allow instant-fire
    - `command_timeout_s` clamped to `[3, 30]`
    - `refresh_interval_ms` — `0` disables auto-refresh, minimum `2000` if enabled
  - Creates `~/.config/luma/` dir + copies `config.toml.example` on first run
- `config_test.go` — test clamping, test partial overrides, test missing file

**✅ Deliverable:** Bad config values are silently corrected, app never runs with unsafe settings

---

#### **Commit 5 — `feat(ui): build core model and display list view`**

> _Goal: Real display data renders in the TUI, executor status visible_

- `internal/ui/model.go`:
  - `Model` struct: `displays`, `selected`, `mode`, `config`, `executor`, `loading`, `err`, `busyIndicator`
  - `Init()` fires a single `tea.Cmd` to run `ddc.Detect()` then fetch all brightness values **sequentially** through the executor
  - `Msg` types: `displaysLoadedMsg`, `brightnessUpdatedMsg`, `errMsg`, `busyMsg`
- `internal/ui/styles.go` — all lipgloss styles, accent from config
- `internal/ui/view.go`:
  - Outer border box
  - Header: `◈ luma` left, `[r]efresh` right + `⚙` spinner when executor is busy
  - Per-display rows: `▸ Display 1  Dell U2723D  ████████████░░░░  72%`
  - Footer keybinding hints
- `internal/ui/keys.go` — all keybindings via `bubbles/key`

**✅ Deliverable:** App launches, shows real displays with brightness bars, spinner shows when ddcutil is running

---

#### **Commit 6 — `feat(ui): implement debounced step controls`**

> _Goal: Nudge keys feel instant in UI but only fire ddcutil after user pauses_

- `internal/ui/update.go`:
  - `↑` / `↓` → change `model.selected` (zero cost, no ddcutil)
  - `+` / `-` → update **local state only** immediately (optimistic UI), reset debounce timer
  - `[` / `]` → same, `±config.Steps.Medium`
  - `{` / `}` → same, `±config.Steps.Large`
  - **Debounce timer** (`tea.Tick` of `config.Guardrails.DebounceMs`):
    - Resets on every keypress
    - Only when timer fires with **no new keypresses** → calls `executor.Run()` with the final value
    - If executor is busy when timer fires → show `⚠ queued` indicator, retry once after 500ms, then drop
  - Value clamped to `[0, display.MaxVal]` before any update
  - `r` key → re-fetches all brightness values (goes through executor, respects busy state)

```
User presses + + + + + (5 times fast)
→ UI shows +5 immediately each time
→ debounce resets each time
→ 300ms after last press: ONE ddcutil setvcp call with final value
```

**✅ Deliverable:** Rapid keypresses feel snappy, only 1 ddcutil call fires per burst

---

#### **Commit 7 — `feat(ui): implement global slider (commit-on-enter only)`**

> _Goal: Slide freely, zero ddcutil calls until Enter is pressed_

- New `Mode`: `ModeSlider`
- `s` → enters slider mode, `sliderVal` = average of all displays
- `←` / `→` → moves `sliderVal` by `Steps.Small` — **pure local state, zero ddcutil**
- `[` / `]` → moves `sliderVal` by `Steps.Medium` — **pure local state, zero ddcutil**
- `Enter` → **one single call** to `ddc.SetBrightnessAll()` through executor, exits slider mode
- `Esc` → cancels, no ddcutil call at all
- Auto-refresh ticker **pauses** while in `ModeSlider`
- View renders:

  ```
  ◈ Set All  [████████████░░░░░░░░]  60%   ← pending, not yet applied
  Press Enter to apply · Esc to cancel
  ```

**✅ Deliverable:** Slider is completely free to move, exactly 1 ddcutil call on confirm

---

#### **Commit 8 — `feat(ui): implement direct number input mode`**

> _Goal: Type exact value, single ddcutil call on Enter_

- New `Mode`: `ModeInput`
- Pressing `0-9` on a selected display enters `ModeInput`
- Renders inline: `▸ Display 1  [ 85_ ]`
- `Enter` → validates range `[0, 100]`, **one** `SetBrightness` call through executor
- `Esc` → cancels, restores previous display value, zero ddcutil calls
- Auto-refresh ticker **pauses** while in `ModeInput`
- Works in both normal mode (per display) and slider mode (set all)

**✅ Deliverable:** Type `85` + `Enter` → exactly 1 ddcutil call, display jumps to 85%

---

#### **Commit 9 — `feat(ui): implement safe auto-refresh`**

> _Goal: Values stay in sync passively without hammering ddcutil_

- `tea.Tick` fires every `config.Display.RefreshIntervalMs` (default 5000ms, 0 = off)
- On tick:
  - **Skip entirely** if executor is busy
  - **Skip entirely** if in `ModeSlider` or `ModeInput`
  - **Skip entirely** if a debounce timer is pending
  - Only proceeds if all clear → fetches brightness values sequentially
- Values only update UI if they actually changed (no flicker)
- Refresh indicator: subtle `↻` in header that flashes briefly when a refresh completes

```
Tick fires → check all guards → if clear: one sequential fetch pass
           → if busy: skip this tick, try next tick
```

**✅ Deliverable:** Auto-refresh never races with user actions, never stacks up calls

---

#### **Commit 10 — `feat(ui): polish error states and edge cases`**

> _Goal: App never crashes, always shows useful feedback_

- `ddcutil` not found → full-screen error with install instructions
- No displays detected → `"No DDC/CI displays found. Check connections."`
- Executor timeout hit → show `"Display N: command timed out"` inline, mark display as `[?]`
- Display goes offline mid-session → mark as `[disconnected]`, skip in all operations
- Executor drops a command (busy) → show brief `⚠` flash on the affected display row
- All errors inline in TUI, never panic or crash to stderr

**✅ Deliverable:** App handles all hardware and process edge cases gracefully

---

#### **Commit 11 — `chore: CI, README, Makefile, and release build`**

> _Goal: Project is shareable and installable_

- `.github/workflows/ci.yml` — `go build` + `go test` on push, build `linux/amd64` + `linux/arm64`
- `README.md` — full docs, TUI screenshot (ASCII), install one-liner, config reference, ddcutil prerequisites
- `config.toml.example` — fully commented
- `Makefile` — `make build`, `make test`, `make install`

**✅ Deliverable:** `go install github.com/yourname/luma@latest` works end-to-end

---

### 📊 Full Commit Summary

| #   | Commit                                                | Guardrail Added                     | What Works After     |
| --- | ----------------------------------------------------- | ----------------------------------- | -------------------- |
| 1   | `chore: init project scaffold`                        | —                                   | App starts and quits |
| 2   | `feat(ddc): ddcutil executor with guardrails`         | ✅ Semaphore, timeout, drop-if-busy | Safe execution layer |
| 3   | `feat(ddc): ddcutil client using executor`            | ✅ Sequential SetAll                | Hardware API ready   |
| 4   | `feat(config): config loader with guardrail defaults` | ✅ Value clamping                   | Safe config always   |
| 5   | `feat(ui): core model and display list view`          | ✅ Busy indicator                   | Real displays render |
| 6   | `feat(ui): debounced step controls`                   | ✅ Debounce 300ms                   | Nudge keys safe      |
| 7   | `feat(ui): global slider commit-on-enter`             | ✅ Zero calls until Enter           | Slider safe          |
| 8   | `feat(ui): direct number input mode`                  | ✅ Zero calls until Enter           | Exact input safe     |
| 9   | `feat(ui): safe auto-refresh`                         | ✅ Skip-if-busy ticker              | Passive sync safe    |
| 10  | `feat(ui): error states and edge cases`               | ✅ Timeout + disconnect handling    | Fully robust         |
| 11  | `chore: CI + README + release`                        | —                                   | Shippable project    |
