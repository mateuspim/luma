# luma

A safe, keyboard-driven TUI for controlling monitor brightness via DDC/CI.

```
╭──────────────────────────────────────────────────────────╮
│◈ luma                              [r]efresh  [s]lider  [q]uit│
├──────────────────────────────────────────────────────────┤
│▸ Dell U2723D         ████████████░░░░░░░░  60%           │
│  LG ULTRAFINE        ██████████████░░░░░░  72%           │
├──────────────────────────────────────────────────────────┤
│  ↑/↓ select · +/- [/] {/} adjust · s slider             │
╰──────────────────────────────────────────────────────────╯
```

## Requirements

- Linux with `ddcutil` installed (`pacman -S ddcutil` / `apt install ddcutil`)
- DDC/CI enabled in your monitor OSD
- User in the `i2c` group: `sudo usermod -aG i2c $USER`

## Installation

```sh
go install github.com/pym/luma@latest
```

Or build from source:

```sh
git clone https://github.com/pym/luma
cd luma
make install
```

## Usage

```sh
luma
```

### Keybindings

| Key            | Action                                      |
| -------------- | ------------------------------------------- |
| `↑` / `↓`     | Select display                              |
| `+` / `-`      | Brightness ±small (debounced 300ms)         |
| `[` / `]`      | Brightness ±medium                          |
| `{` / `}`      | Brightness ±large                           |
| `0–9`          | Enter direct number input mode              |
| `s`            | Global slider (all displays, commit-on-Enter) |
| `r`            | Refresh brightness values                   |
| `Esc`          | Cancel current mode                         |
| `Enter`        | Confirm slider / input                      |
| `q` / `Ctrl+C` | Quit                                        |

## Guardrail Design

luma is built around 4 safety rules:

| Rule                | Description                                              |
| ------------------- | -------------------------------------------------------- |
| **Single Executor** | Only 1 `ddcutil` process at a time (semaphore-guarded)   |
| **Debounce**        | Step keys only fire ddcutil 300ms after the last press   |
| **Commit-on-Enter** | Slider and input never call ddcutil until Enter is pressed |
| **Hard Cap**        | Max 2 concurrent ddcutil processes, enforced via config  |

## Configuration

Copy `config.toml.example` to `~/.config/luma/config.toml`:

```sh
mkdir -p ~/.config/luma
cp config.toml.example ~/.config/luma/config.toml
```

```toml
[steps]
small  = 1
medium = 5
large  = 10

[display]
refresh_interval_ms = 5000   # 0 = disabled
show_display_name   = true

[guardrails]
debounce_ms       = 300   # [100–2000]
max_ddcutil_procs = 1     # [1–2]
command_timeout_s = 8     # [3–30]

[theme]
accent_color = "#A78BFA"
```

All guardrail values are clamped to their valid ranges — bad config is silently corrected.

## License

MIT
