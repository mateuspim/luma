package ddc

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Display represents a detected monitor.
type Display struct {
	Index      int
	Name       string
	Model      string
	Brightness int
	MaxVal     int
}

var vcpRe = regexp.MustCompile(`current value\s*=\s*(\d+),\s*max value\s*=\s*(\d+)`)

// Client wraps an Executor with high-level display operations.
type Client struct {
	exec *Executor
}

// NewClient creates a Client using the given Executor.
func NewClient(exec *Executor) *Client {
	return &Client{exec: exec}
}

// Detect returns all DDC/CI displays visible to ddcutil.
// Returns ErrNotFound if ddcutil is not installed.
func (c *Client) Detect(ctx context.Context) ([]Display, error) {
	out, err := c.exec.Run(ctx, "detect", "--brief")
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ddcutil detect: %w", err)
	}
	return parseDetect(out), nil
}

// GetBrightness fetches current and max brightness for a display by its ddcutil index.
func (c *Client) GetBrightness(ctx context.Context, displayIndex int) (current, max int, err error) {
	out, err := c.exec.Run(ctx, "--display", strconv.Itoa(displayIndex), "getvcp", "10")
	if err != nil {
		if isNotFound(err) {
			return 0, 0, ErrNotFound
		}
		return 0, 0, fmt.Errorf("ddcutil getvcp: %w", err)
	}
	current, max, err = parseVCP(out)
	if err != nil {
		return 0, 0, fmt.Errorf("parse getvcp output: %w", err)
	}
	return current, max, nil
}

// SetBrightness sets brightness for a single display.
func (c *Client) SetBrightness(ctx context.Context, displayIndex, value int) error {
	_, err := c.exec.Run(ctx, "--display", strconv.Itoa(displayIndex), "setvcp", "10", strconv.Itoa(value))
	if err != nil {
		if isNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("ddcutil setvcp: %w", err)
	}
	return nil
}

// SetBrightnessAll sets brightness on all displays sequentially through the executor.
func (c *Client) SetBrightnessAll(ctx context.Context, displays []Display, value int) error {
	for _, d := range displays {
		if err := c.SetBrightness(ctx, d.Index, value); err != nil {
			return fmt.Errorf("display %d: %w", d.Index, err)
		}
	}
	return nil
}

// parseDetect parses `ddcutil detect --brief` output into Display slices.
// Example lines:
//
//	Display 1
//	   Model:   Dell U2723D
func parseDetect(out string) []Display {
	var displays []Display
	var current *Display

	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Display ") {
			if current != nil {
				displays = append(displays, *current)
			}
			indexStr := strings.TrimPrefix(trimmed, "Display ")
			idx, _ := strconv.Atoi(strings.Fields(indexStr)[0])
			current = &Display{Index: idx, Name: trimmed}
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(trimmed, "Model:") {
			current.Model = strings.TrimSpace(strings.TrimPrefix(trimmed, "Model:"))
		}
	}
	if current != nil {
		displays = append(displays, *current)
	}
	return displays
}

// parseVCP parses `ddcutil getvcp 10` output for current/max brightness.
func parseVCP(out string) (current, max int, err error) {
	m := vcpRe.FindStringSubmatch(out)
	if m == nil {
		return 0, 0, fmt.Errorf("no VCP value found in output: %q", out)
	}
	current, err = strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, err
	}
	max, err = strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, err
	}
	return current, max, nil
}

// isNotFound detects "executable file not found" errors.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "executable file not found")
}
