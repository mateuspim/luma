package ddc

import (
	"testing"
)

func TestParseDetect(t *testing.T) {
	input := `
Display 1
   I2C bus:  /dev/i2c-4
   DRM_connector:           card1-DP-2
   EDID synopsis:
      Mfg id:               AOC - UNK
      Model:                27G2G4
      Serial number:

Display 2
   I2C bus:  /dev/i2c-6
   EDID synopsis:
      Mfg id:               CMI - C-Media Electronics
      Model:                GP2711
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
	if displays[0].Model != "27G2G4" {
		t.Errorf("display[0].Model = %q, want %q", displays[0].Model, "27G2G4")
	}
	if displays[0].Name != "AOC 27G2G4" {
		t.Errorf("display[0].Name = %q, want %q", displays[0].Name, "AOC 27G2G4")
	}
	if displays[1].Index != 2 {
		t.Errorf("display[1].Index = %d, want 2", displays[1].Index)
	}
	if displays[1].Bus != 6 {
		t.Errorf("display[1].Bus = %d, want 6", displays[1].Bus)
	}
	if displays[1].Name != "CMI GP2711" {
		t.Errorf("display[1].Name = %q, want %q", displays[1].Name, "CMI GP2711")
	}
}

func TestParseDetect_FallbackName(t *testing.T) {
	// No Mfg id or Model → falls back to "Display N"
	input := `
Display 3
   I2C bus:  /dev/i2c-2
`
	displays := parseDetect(input)
	if len(displays) != 1 {
		t.Fatalf("expected 1 display, got %d", len(displays))
	}
	if displays[0].Name != "Display 3" {
		t.Errorf("Name = %q, want %q", displays[0].Name, "Display 3")
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
