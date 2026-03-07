# luma

> Minimal TUI brightness controller for DDC/CI displays

## Requirements

- `ddcutil` installed and accessible in `$PATH`
- Go 1.21+
- `i2c` kernel module loaded (`sudo modprobe i2c-dev`)
- User in `i2c` group (`sudo usermod -aG i2c $USER`)

## Installation

```sh
git clone https://github.com/pym/luma
cd luma
make build
```

Copy the example config:

```sh
mkdir -p ~/.config/luma
cp config.toml.example ~/.config/luma/config.toml
```

## DDC/CI Optimization

By default, `ddcutil` is slow (~1.2s per call). To dramatically improve performance, create `~/.config/ddcutil/ddcutilrc`:

```toml
[ddcutil]
options: --sleep-multiplier 0.1 --skip-ddc-checks --noverify
```

**Bus caching:** luma runs `ddcutil detect` at startup and caches `--bus N` for each display. This drops per-call scan overhead from ~1.2s to ~0.124s by avoiding repeated bus enumeration.

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

### Configuration Fields

**`[steps]`**

- `small` (default: 1) — Step size for `+/-` keys
- `medium` (default: 5) — Step size for `[/]` keys
- `large` (default: 10) — Step size for `{/}` keys

**`[display]`**

- `refresh_interval_ms` (default: 5000) — Auto-refresh brightness values. Set to `0` to disable. Minimum `2000` if enabled. Refresh only occurs when idle (no ddcutil processes running, no pending debounce timers, not in slider mode).
- `show_display_name` (default: true) — Display monitor names in list view

**`[guardrails]`**

- `debounce_ms` (default: 300, range: 100-2000) — Wait time after last keypress before firing ddcutil. Prevents command spam during rapid key presses.
- `max_ddcutil_procs` (default: 1, range: 1-2) — Hard cap on concurrent ddcutil processes. Commands are dropped (non-blocking) if limit is reached.
- `command_timeout_s` (default: 8, range: 3-30) — Kill ddcutil if it hangs beyond this timeout

**`[theme]`**

- `accent_color` (default: "#9B59B6") — Hex color for UI accents (header, cursor, selected row, separator, gradient bar fill)

All guardrail values are clamped to their valid ranges — invalid config is silently corrected.

## Usage

```sh
./luma
```

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

**Note:** Step sizes shown in footer are dynamic — they reflect current config values.

## UI Layout

### List View

```
╭────────────────────────────────────────────────────────────────────────╮
│◈ luma ⠋                                                                │
├────────────────────────────────────────────────────────────────────────┤
│▸ AOC 27G2G4       ╸██████████████████████──────────────────────╺   50% │
│  CMI GP2711       ╸██████████████████████──────────────────────╺   50% │
├────────────────────────────────────────────────────────────────────────┤
│  ↑/↓ select · Enter slider · [a]ll · [q]uit ◆ +/- ±1 · [/] ±5 · {/} ±10│
╰────────────────────────────────────────────────────────────────────────╯
```

- `⠋` spinner appears in header when ddcutil is running (cycles through `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`)
- `▸` and filled `█` blocks rendered in `accent_color`
- Selected row name and percentage rendered in **bold**
- Bar fill color interpolates from `#000000` → `accent_color` left to right
- Auto-detects connected displays at startup, no hardcoded display count

### Per-display Slider View

```
╭────────────────────────────────────────────────────────────────────────╮
│◈ luma · AOC 27G2G4                                                     │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│   ╸████████████████████████████████████──────────────────────╺   42%   │
│                                                                        │
├────────────────────────────────────────────────────────────────────────┤
│  ←/→ ±1 · [/] ±5 · {/} ±10 · Enter apply · Esc back                    │
╰────────────────────────────────────────────────────────────────────────╯
```

### Set All Slider View

```
╭────────────────────────────────────────────────────────────────────────╮
│◈ luma · All Displays                                                   │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│   ╸████████████████████████████████████──────────────────────╺   42%   │
│                                                                        │
├────────────────────────────────────────────────────────────────────────┤
│  ←/→ ±1 · [/] ±5 · {/} ±10 · Enter apply · Esc back                    │
╰────────────────────────────────────────────────────────────────────────╯
```

- Slider footer step values are dynamic from config
- Bar gradient same as list view: `#000000` → `accent_color`
- Comfortable padding: 3 spaces left and right of bar

## License

MIT
