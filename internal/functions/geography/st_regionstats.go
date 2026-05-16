package geography

import (
	"encoding/binary"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

// BindStRegionStats reads a local GeoTIFF and reports
// count/min/max/mean/std/sum/area statistics for the pixels
// intersecting the input GEOGRAPHY's bounding box.
//
// Arguments:
//
//	args[0]: GEOGRAPHY  — region of interest
//	args[1]: STRING     — raster identifier
//	                       file:///path/to/raster.tif
//	                       /local/absolute/path.tif
//	                       https://host/raster.tif
//	                       (Earth-Engine asset paths are not
//	                       resolvable locally.)
//
// Returns a STRUCT(count INT64, min FLOAT64, max FLOAT64,
// sum FLOAT64, mean FLOAT64, std FLOAT64, area FLOAT64).
func BindStRegionStats(args ...value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 5 {
		return nil, sqError("ST_REGIONSTATS", "invalid number of arguments: got %d, want between 2 and 5", len(args))
	}
	g := geographyArg(args[0])
	if g == nil || args[1] == nil {
		return nil, nil
	}
	rasterID, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	data, err := loadRasterBytes(rasterID)
	if err != nil {
		return nil, sqError("ST_REGIONSTATS", "%v", err)
	}
	stats, err := computeRegionStats(g, data)
	if err != nil {
		return nil, sqError("ST_REGIONSTATS", "%v", err)
	}
	return stats, nil
}

func loadRasterBytes(raster string) ([]byte, error) {
	if strings.HasPrefix(raster, "https://") || strings.HasPrefix(raster, "http://") {
		resp, err := http.Get(raster)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("raster fetch returned status %d", resp.StatusCode)
		}
		buf := make([]byte, 0, 64<<10)
		tmp := make([]byte, 32<<10)
		for {
			n, rerr := resp.Body.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if rerr != nil {
				break
			}
		}
		return buf, nil
	}
	path := strings.TrimPrefix(raster, "file://")
	return os.ReadFile(path)
}

// computeRegionStats parses a GeoTIFF byte stream and accumulates
// the requested statistics over pixels whose centre falls inside
// the geography's bounding box.
func computeRegionStats(g *value.GeographyValue, raw []byte) (*value.StructValue, error) {
	tiff, err := parseGeoTIFF(raw)
	if err != nil {
		return nil, err
	}
	minLng, minLat, maxLng, maxLat, ok := geographyBBox(g)
	if !ok {
		return nil, fmt.Errorf("empty geography")
	}
	var (
		count   int64
		sumVal  float64
		sumSq   float64
		minVal  = math.Inf(1)
		maxVal  = math.Inf(-1)
		areaSum float64
	)
	for y := 0; y < tiff.height; y++ {
		for x := 0; x < tiff.width; x++ {
			lng, lat := tiff.pixelToWorld(x, y)
			if lng < minLng || lng > maxLng || lat < minLat || lat > maxLat {
				continue
			}
			v, ok := tiff.pixelValue(x, y)
			if !ok {
				continue
			}
			count++
			sumVal += v
			sumSq += v * v
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
			// Approximate per-pixel area in m² assuming pixels are
			// small enough that local equirectangular projection
			// holds. dx ≈ pixelScaleX deg * cos(lat) * 111195 m,
			// dy ≈ pixelScaleY deg * 111195 m.
			dx := math.Abs(tiff.pixelScaleX) * math.Cos(lat*math.Pi/180) * 111195
			dy := math.Abs(tiff.pixelScaleY) * 111195
			areaSum += dx * dy
		}
	}
	if count == 0 {
		minVal = 0
		maxVal = 0
	}
	mean := 0.0
	std := 0.0
	if count > 0 {
		mean = sumVal / float64(count)
		variance := sumSq/float64(count) - mean*mean
		if variance > 0 {
			std = math.Sqrt(variance)
		}
	}
	keys := []string{"count", "min", "max", "sum", "mean", "std", "area"}
	vals := []value.Value{
		value.IntValue(count),
		value.FloatValue(minVal),
		value.FloatValue(maxVal),
		value.FloatValue(sumVal),
		value.FloatValue(mean),
		value.FloatValue(std),
		value.FloatValue(areaSum),
	}
	m := map[string]value.Value{}
	for i, k := range keys {
		m[k] = vals[i]
	}
	return &value.StructValue{Keys: keys, Values: vals, M: m}, nil
}

func geographyBBox(g *value.GeographyValue) (minLng, minLat, maxLng, maxLat float64, ok bool) {
	minLng, minLat = 180, 90
	maxLng, maxLat = -180, -90
	any := false
	push := func(lng, lat float64) {
		any = true
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}
	for _, p := range pointsOf(g) {
		push(p[0], p[1])
	}
	return minLng, minLat, maxLng, maxLat, any
}

// ----- minimal GeoTIFF parser -----

type geoTIFF struct {
	width, height int
	bitsPerSample int
	sampleFormat  int // 1=uint, 2=int, 3=float
	tiePoint      [2]float64
	pixelScaleX   float64
	pixelScaleY   float64
	pixelOriginX  float64
	pixelOriginY  float64
	data          []float64
	byteOrder     binary.ByteOrder
}

func (t *geoTIFF) pixelToWorld(x, y int) (lng, lat float64) {
	lng = t.pixelOriginX + (float64(x)+0.5)*t.pixelScaleX
	lat = t.pixelOriginY - (float64(y)+0.5)*t.pixelScaleY
	return
}

func (t *geoTIFF) pixelValue(x, y int) (float64, bool) {
	if x < 0 || y < 0 || x >= t.width || y >= t.height {
		return 0, false
	}
	v := t.data[y*t.width+x]
	if math.IsNaN(v) {
		return 0, false
	}
	return v, true
}

func parseGeoTIFF(raw []byte) (*geoTIFF, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("TIFF too short")
	}
	var bo binary.ByteOrder
	switch {
	case raw[0] == 'I' && raw[1] == 'I':
		bo = binary.LittleEndian
	case raw[0] == 'M' && raw[1] == 'M':
		bo = binary.BigEndian
	default:
		return nil, fmt.Errorf("not a TIFF header")
	}
	if bo.Uint16(raw[2:]) != 42 {
		return nil, fmt.Errorf("TIFF version != 42")
	}
	ifdOff := bo.Uint32(raw[4:])
	if int(ifdOff)+2 > len(raw) {
		return nil, fmt.Errorf("IFD offset out of bounds")
	}
	tiff := &geoTIFF{byteOrder: bo, bitsPerSample: 8, sampleFormat: 1}
	nEntries := int(bo.Uint16(raw[ifdOff:]))
	entryStart := int(ifdOff) + 2
	if entryStart+nEntries*12 > len(raw) {
		return nil, fmt.Errorf("IFD entries out of bounds")
	}
	var stripOffsets, stripByteCounts []uint32
	for i := 0; i < nEntries; i++ {
		off := entryStart + i*12
		tag := bo.Uint16(raw[off:])
		typ := bo.Uint16(raw[off+2:])
		cnt := bo.Uint32(raw[off+4:])
		valOff := bo.Uint32(raw[off+8:])
		switch tag {
		case 256: // ImageWidth
			tiff.width = int(valOff)
		case 257: // ImageLength
			tiff.height = int(valOff)
		case 258: // BitsPerSample
			tiff.bitsPerSample = int(valOff & 0xFFFF)
		case 339: // SampleFormat
			tiff.sampleFormat = int(valOff & 0xFFFF)
		case 273: // StripOffsets
			stripOffsets = readUint32Values(raw, bo, typ, cnt, valOff)
		case 279: // StripByteCounts
			stripByteCounts = readUint32Values(raw, bo, typ, cnt, valOff)
		case 33550: // ModelPixelScaleTag
			vals := readFloat64Values(raw, bo, typ, cnt, valOff)
			if len(vals) >= 2 {
				tiff.pixelScaleX = vals[0]
				tiff.pixelScaleY = vals[1]
			}
		case 33922: // ModelTiepointTag
			vals := readFloat64Values(raw, bo, typ, cnt, valOff)
			if len(vals) >= 6 {
				// (I, J, K, X, Y, Z): pixel (I,J,K) maps to (X,Y,Z) world.
				tiff.tiePoint = [2]float64{vals[3], vals[4]}
				tiff.pixelOriginX = vals[3] - vals[0]*tiff.pixelScaleX
				tiff.pixelOriginY = vals[4] + vals[1]*tiff.pixelScaleY
			}
		}
	}
	if tiff.width == 0 || tiff.height == 0 {
		return nil, fmt.Errorf("missing dimensions")
	}
	tiff.data = make([]float64, tiff.width*tiff.height)
	pixelOffset := 0
	for s := 0; s < len(stripOffsets); s++ {
		off := int(stripOffsets[s])
		count := int(stripByteCounts[s])
		if off+count > len(raw) {
			return nil, fmt.Errorf("strip out of bounds")
		}
		strip := raw[off : off+count]
		bytesPerSample := tiff.bitsPerSample / 8
		if bytesPerSample == 0 {
			bytesPerSample = 1
		}
		nPixels := count / bytesPerSample
		for i := 0; i < nPixels && pixelOffset+i < len(tiff.data); i++ {
			tiff.data[pixelOffset+i] = decodeSample(strip[i*bytesPerSample:], bytesPerSample, tiff.sampleFormat, bo)
		}
		pixelOffset += nPixels
	}
	return tiff, nil
}

func readUint32Values(raw []byte, bo binary.ByteOrder, typ uint16, cnt, valOff uint32) []uint32 {
	if cnt == 1 {
		return []uint32{valOff}
	}
	out := make([]uint32, cnt)
	size := uint32(2)
	if typ == 4 {
		size = 4
	}
	if int(valOff+cnt*size) > len(raw) {
		return out
	}
	for i := uint32(0); i < cnt; i++ {
		if typ == 3 {
			out[i] = uint32(bo.Uint16(raw[valOff+i*2:]))
		} else {
			out[i] = bo.Uint32(raw[valOff+i*4:])
		}
	}
	return out
}

func readFloat64Values(raw []byte, bo binary.ByteOrder, typ uint16, cnt, valOff uint32) []float64 {
	size := uint32(8)
	if int(valOff+cnt*size) > len(raw) {
		return nil
	}
	out := make([]float64, cnt)
	for i := uint32(0); i < cnt; i++ {
		out[i] = math.Float64frombits(bo.Uint64(raw[valOff+i*size:]))
	}
	return out
}

func decodeSample(b []byte, bytesPerSample, format int, bo binary.ByteOrder) float64 {
	switch format {
	case 3: // float
		switch bytesPerSample {
		case 4:
			return float64(math.Float32frombits(bo.Uint32(b)))
		case 8:
			return math.Float64frombits(bo.Uint64(b))
		}
	case 2: // int
		switch bytesPerSample {
		case 1:
			return float64(int8(b[0]))
		case 2:
			return float64(int16(bo.Uint16(b)))
		case 4:
			return float64(int32(bo.Uint32(b)))
		}
	default: // unsigned int
		switch bytesPerSample {
		case 1:
			return float64(b[0])
		case 2:
			return float64(bo.Uint16(b))
		case 4:
			return float64(bo.Uint32(b))
		}
	}
	return 0
}
