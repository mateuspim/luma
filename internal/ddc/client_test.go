package ddc

import (
	"testing"
)

func TestParseDetect(t *testing.T) {
	input := `
Display 1
   I2C bus:  /dev/i2c-4
   DRM connector: card1-DP-3
   Model:   Dell U2723D
   Serial number: ...

Display 2
   I2C bus:  /dev/i2c-6
   Model:   LG ULTRAFINE
`
	displays := parseDetect(input)
	if len(displays) != 2 {
		t.Fatalf("expected 2 displays, got %d", len(displays))
	}
	if displays[0].Index != 1 {
		t.Errorf("display[0].Index = %d, want 1", displays[0].Index)
	}
	if displays[0].Bus != 4 {
		t.Errorf("display[0].Bus = %d, want 4", displays[0].Bus)
	}
	if displays[0].Model != "Dell U2723D" {
		t.Errorf("display[0].Model = %q, want %q", displays[0].Model, "Dell U2723D")
	}
	if displays[1].Index != 2 {
		t.Errorf("display[1].Index = %d, want 2", displays[1].Index)
	}
	if displays[1].Bus != 6 {
		t.Errorf("display[1].Bus = %d, want 6", displays[1].Bus)
	}
	if displays[1].Model != "LG ULTRAFINE" {
		t.Errorf("display[1].Model = %q, want %q", displays[1].Model, "LG ULTRAFINE")
	}
}

func TestParseDetect_Empty(t *testing.T) {
	displays := parseDetect("No DDC/CI displays found\n")
	if len(displays) != 0 {
		t.Errorf("expected 0 displays, got %d", len(displays))
	}
}

func TestParseVCP(t *testing.T) {
	cases := []struct {
		input   string
		current int
		max     int
		wantErr bool
	}{
		{
			input:   `VCP code 0x10 (Brightness                    ): current value =   72, max value =  100`,
			current: 72, max: 100,
		},
		{
			input:   `VCP code 0x10 (Brightness): current value =    0, max value =  100`,
			current: 0, max: 100,
		},
		{
			input:   `VCP code 0x10 (Brightness): current value =  100, max value =  100`,
			current: 100, max: 100,
		},
		{
			input:   "some garbage output",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		cur, mx, err := parseVCP(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseVCP(%q): expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseVCP(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if cur != tc.current {
			t.Errorf("parseVCP(%q): current = %d, want %d", tc.input, cur, tc.current)
		}
		if mx != tc.max {
			t.Errorf("parseVCP(%q): max = %d, want %d", tc.input, mx, tc.max)
		}
	}
}
