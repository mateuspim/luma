# luma

> Minimal TUI brightness controller for DDC/CI displays

## Requirements

- `ddcutil` installed and accessible in `$PATH`
- Go 1.25+
- `i2c` kernel module loaded (`sudo modprobe i2c-dev`)
- User in `i2c` group (`sudo usermod -aG i2c $USER`)

## Installation

### From Release (Recommended)

Download the latest release for your platform:

```sh
# Linux (amd64)
curl -L https://github.com/mateuspim/luma/releases/download/latest/luma-linux-amd64.tar.gz | tar xz
sudo mv luma /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/mateuspim/luma/releases/download/latest/luma-linux-arm64.tar.gz | tar xz
sudo mv luma /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/mateuspim/luma/releases/download/latest/luma-darwin-arm64.tar.gz | tar xz
sudo mv luma /usr/local/bin/
```

Copy the example config:

```sh
mkdir -p ~/.config/luma
cp config.toml.example ~/.config/luma/config.toml
```

### From Source

```sh
git clone https://github.com/mateuspim/luma
cd luma
make build
sudo mv luma /usr/local/bin/
```

Copy the example config:

```sh
mkdir -p ~/.config/luma
cp config.toml.example ~/.config/luma/config.toml
```

## Usage

```sh
luma
```

The UI fills the terminal window automatically and adapts to any terminal size.

## Keybindings

### List View

| Key       | Action                           |
| --------- | -------------------------------- |
| `↑` / `↓` | Select display                   |
| `+` / `-` | Brightness ±small                |
| `[` / `]` | Brightness ±medium               |
| `{` / `}` | Brightness ±large                |
| `Enter`   | Open slider for selected display |
| `a`       | Open Set All slider              |
| `q`       | Quit                             |

### Slider View (per-display and Set All)

| Key       | Action                     |
| --------- | -------------------------- |
| `←` / `→` | Brightness ±small          |
| `[` / `]` | Brightness ±medium         |
| `{` / `}` | Brightness ±large          |
| `Enter`   | Apply and return to list   |
| `Esc`     | Discard and return to list |

Step sizes shown in the footer reflect your current config values.

## Configuration

Full annotated `config.toml` example:

```toml
[steps]
small  = 1    # +/- keys
medium = 5    # [/] keys
large  = 10   # {/} keys

[display]
refresh_interval_ms = 5000   # 0 = disabled, minimum 2000 if enabled
show_display_name   = true

[guardrails]
debounce_ms       = 300   # wait after last keypress before firing ddcutil [100-2000]
max_ddcutil_procs = 1     # hard cap on concurrent ddcutil processes [1-2]
command_timeout_s = 8     # kill ddcutil if it hangs beyond this [3-30]

[theme]
accent_color = "#9B59B6"
```

### Fields

**`[steps]`**

- `small` (default: 1) — Step size for `+/-` keys
- `medium` (default: 5) — Step size for `[/]` keys
- `large` (default: 10) — Step size for `{/}` keys

**`[display]`**

- `refresh_interval_ms` (default: 5000) — Auto-refresh brightness values. Set to `0` to disable. Minimum `2000` if enabled. Refresh only runs when idle (no ddcutil processes running, no pending debounce, not in slider mode).
- `show_display_name` (default: true) — Display monitor names in the list view

**`[guardrails]`**

- `debounce_ms` (default: 300, range: 100–2000) — Wait time after last keypress before firing ddcutil. Prevents command spam during rapid key presses.
- `max_ddcutil_procs` (default: 1, range: 1–2) — Hard cap on concurrent ddcutil processes. Commands are dropped (non-blocking) if the limit is reached.
- `command_timeout_s` (default: 8, range: 3–30) — Kill ddcutil if it hangs beyond this timeout.

**`[theme]`**

- `accent_color` (default: `"#9B59B6"`) — Hex color for UI accents (header, cursor, selected row, separator, gradient bar fill).

All guardrail values are clamped to their valid ranges — invalid config is silently corrected.

## Performance

By default, `ddcutil` is slow (~1.2s per call). To speed it up, create `~/.config/ddcutil/ddcutilrc`:

```
options: --sleep-multiplier 0.1 --skip-ddc-checks --noverify
```

**Bus caching:** luma runs `ddcutil detect` at startup and caches `--bus N` for each display, dropping per-call overhead from ~1.2s to ~0.124s by avoiding repeated bus enumeration.

## UI Layout

The UI adapts to the terminal size:

- **Wide (≥74 cols):** full bar, 14-char name column, full footer
- **Compact (40–73 cols):** narrower bar, 10-char name column, abbreviated footer
- **Minimal (<40 cols):** no bar, percentage only, compact footer

### List View

```
╭────────────────────────────────────────────────────────────────────────╮
│◈ luma v1.0.0 ⠋                                                         │
├────────────────────────────────────────────────────────────────────────┤
│▸ AOC 27G2G4       ╸██████████████████████──────────────────────╺   50% │
│  CMI GP2711       ╸████████████████████████████████────────────╺   75% │
├────────────────────────────────────────────────────────────────────────┤
│  ↑/↓ Enter · [a]ll · [q]uit ◆ +/- ±1 · [/] ±5 · {/} ±10                │
╰────────────────────────────────────────────────────────────────────────╯
```

- `⠋` spinner appears when ddcutil is running
- `▸` cursor and `█` fill rendered in `accent_color`
- Selected row name and percentage rendered in **bold**
- Bar fill interpolates from `#000000` → `accent_color`

### Slider View

```
╭────────────────────────────────────────────────────────────────────────╮
│◈ luma v1.0.0 · AOC 27G2G4                                     [q]uit   │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│   ╸████████████████████████████████────────────────────────────╺   42% │
│                                                                        │
├────────────────────────────────────────────────────────────────────────┤
│  ←/→ ±1 · [/] ±5 · {/} ±10 · Enter apply · Esc back                    │
╰────────────────────────────────────────────────────────────────────────╯
```

The Set All slider (`a` from the list) shows `· All Displays` in the title instead of the display name.

## Development

To trigger a versioned release, push a tag:

```sh
git tag v1.0.0
git push --tags
```

This automatically builds binaries for all platforms and creates a GitHub Release with an auto-generated changelog.

## License

MIT
