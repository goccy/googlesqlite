package value

import "testing"

func Test_GeographyNormalizeLongitude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0, 0},
		{"within range positive", 45, 45},
		{"within range negative", -45, -45},
		{"exact 180", 180, 180},
		{"exact -180", -180, -180},
		{"over 180", 190, -170},
		{"under -180", -190, 170},
		{"full rotation positive", 360, 0},
		{"full rotation negative", -360, 0},
		{"multiple rotations", 540, -180},
		{"large positive", 1080 + 30, 30},
		{"large negative", -1080 - 30, -30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := geographyNormalizeLongitude(tt.input)
			if res != tt.expected {
				t.Fatalf(
					"unexpected result for %f: got %f, expected %f",
					tt.input, res, tt.expected,
				)
			}
		})
	}
}
