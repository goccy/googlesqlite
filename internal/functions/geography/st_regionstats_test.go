package geography

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

// buildTestGeoTIFF returns a minimal little-endian GeoTIFF byte
// stream encoding a width-by-height grid of float32 pixels with
// the requested per-pixel value at (x, y). Origin (lng/lat) and
// pixel scale (degrees per pixel) come from the caller so the
// test can position the raster anywhere on the globe.
func buildTestGeoTIFF(width, height int, pixel func(x, y int) float32, originLng, originLat, pxX, pxY float64) []byte {
	bo := binary.LittleEndian
	// Compute pixel-data length and offsets.
	imgSize := width * height * 4
	// IFD entry count: ImageWidth, ImageLength, BitsPerSample,
	// SampleFormat, StripOffsets, StripByteCounts, ModelPixelScale,
	// ModelTiepoint = 8 entries.
	entryCount := 8
	headerLen := 8
	ifdLen := 2 + entryCount*12 + 4 // 2 byte count + entries + next-IFD pointer
	// Out-of-line value blocks: ModelPixelScale (3 doubles = 24B),
	// ModelTiepoint (6 doubles = 48B). Pack them after the IFD.
	pixelScaleOff := uint32(headerLen + ifdLen)
	tiePointOff := pixelScaleOff + 24
	imgOff := tiePointOff + 48
	buf := make([]byte, int(imgOff)+imgSize)
	// Header.
	buf[0], buf[1] = 'I', 'I'
	bo.PutUint16(buf[2:], 42)
	bo.PutUint32(buf[4:], uint32(headerLen))
	// IFD.
	idx := headerLen
	bo.PutUint16(buf[idx:], uint16(entryCount))
	idx += 2
	writeEntry := func(tag uint16, typ uint16, count uint32, valOrOff uint32) {
		bo.PutUint16(buf[idx:], tag)
		bo.PutUint16(buf[idx+2:], typ)
		bo.PutUint32(buf[idx+4:], count)
		bo.PutUint32(buf[idx+8:], valOrOff)
		idx += 12
	}
	writeEntry(256, 3, 1, uint32(width))    // ImageWidth (SHORT)
	writeEntry(257, 3, 1, uint32(height))   // ImageLength (SHORT)
	writeEntry(258, 3, 1, 32)               // BitsPerSample = 32
	writeEntry(339, 3, 1, 3)                // SampleFormat = float
	writeEntry(273, 4, 1, imgOff)           // StripOffsets (LONG)
	writeEntry(279, 4, 1, uint32(imgSize))  // StripByteCounts (LONG)
	writeEntry(33550, 12, 3, pixelScaleOff) // ModelPixelScale (DOUBLE x3)
	writeEntry(33922, 12, 6, tiePointOff)   // ModelTiepoint (DOUBLE x6)
	bo.PutUint32(buf[idx:], 0)              // next IFD = 0
	// Pixel-scale values (X, Y, Z).
	bo.PutUint64(buf[pixelScaleOff:], math.Float64bits(pxX))
	bo.PutUint64(buf[pixelScaleOff+8:], math.Float64bits(pxY))
	bo.PutUint64(buf[pixelScaleOff+16:], math.Float64bits(0))
	// Tiepoint: (I=0, J=0, K=0) maps to (X=originLng, Y=originLat, Z=0).
	bo.PutUint64(buf[tiePointOff:], math.Float64bits(0))
	bo.PutUint64(buf[tiePointOff+8:], math.Float64bits(0))
	bo.PutUint64(buf[tiePointOff+16:], math.Float64bits(0))
	bo.PutUint64(buf[tiePointOff+24:], math.Float64bits(originLng))
	bo.PutUint64(buf[tiePointOff+32:], math.Float64bits(originLat))
	bo.PutUint64(buf[tiePointOff+40:], math.Float64bits(0))
	// Pixel data.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			bo.PutUint32(buf[int(imgOff)+(y*width+x)*4:], math.Float32bits(pixel(x, y)))
		}
	}
	return buf
}

func TestStRegionStats(t *testing.T) {
	// 4x4 raster where every pixel has the same value 10.0.
	data := buildTestGeoTIFF(4, 4, func(x, y int) float32 { return 10.0 }, 0.0, 4.0, 1.0, 1.0)
	tmp := filepath.Join(t.TempDir(), "raster.tif")
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		t.Fatal(err)
	}
	g := value.NewGeographyPolygon([][][2]float64{
		{
			{0, 0},
			{4, 0},
			{4, 4},
			{0, 4},
			{0, 0},
		},
	})
	out, err := BindStRegionStats(g, value.StringValue(tmp))
	if err != nil {
		t.Fatal(err)
	}
	sv, ok := out.(*value.StructValue)
	if !ok {
		t.Fatalf("expected StructValue, got %T", out)
	}
	want := map[string]float64{
		"count": 16,
		"min":   10,
		"max":   10,
		"sum":   160,
		"mean":  10,
		"std":   0,
	}
	for k, w := range want {
		v, ok := sv.M[k]
		if !ok {
			t.Fatalf("missing field %q", k)
		}
		f, err := v.ToFloat64()
		if err != nil {
			t.Fatalf("%s: %v", k, err)
		}
		if math.Abs(f-w) > 1e-6 {
			t.Errorf("%s: got %v, want %v", k, f, w)
		}
	}
}
