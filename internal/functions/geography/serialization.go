package geography

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
)

var jsonUnmarshal = json.Unmarshal

// BindStAsGeoJSON renders the geography as a GeoJSON document.
func BindStAsGeoJSON(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_ASGEOJSON", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	s, err := geographyToGeoJSON(g)
	if err != nil {
		return nil, err
	}
	return value.StringValue(s), nil
}

// BindStGeogFromGeoJSON parses a GeoJSON document into a Geography.
func BindStGeogFromGeoJSON(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, sqError("ST_GEOGFROMGEOJSON", "invalid number of arguments: got %d, want between 1 and 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	s, err := args[0].ToString()
	if err != nil {
		return nil, err
	}
	return geographyFromGeoJSON(s)
}

// BindStAsBinary returns the WKB encoding of the geography.
func BindStAsBinary(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, sqError("ST_ASBINARY", "invalid number of arguments: got %d, want 1", len(args))
	}
	g := geographyArg(args[0])
	if g == nil {
		return nil, nil
	}
	b, err := geographyToWKB(g)
	if err != nil {
		return nil, err
	}
	return value.BytesValue(b), nil
}

// BindStGeogFromWKB parses a WKB byte string into a Geography.
//
// Upstream signatures:
//
//	ST_GEOGFROMWKB(wkb_string [, oriented])
//	ST_GEOGFROMWKB(wkb_string [, planar => boolean]
//	                          [, make_valid => boolean]
//	                          [, oriented => boolean])
//
// The analyzer always materialises named arguments as positional
// after the leading `wkb` argument, with unspecified ones defaulted
// to NULL/FALSE — so the runtime can see up to 4 arguments here.
// `planar` / `make_valid` / `oriented` are advisory hints to the
// upstream parser; our parser accepts both planar and geodesic input
// equivalently, so we treat the flags as informational and parse the
// WKB straight through.
func BindStGeogFromWKB(args ...value.Value) (value.Value, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, sqError("ST_GEOGFROMWKB", "invalid number of arguments: got %d, want between 1 and 4", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	// Upstream accepts either raw BYTES or a hex-encoded STRING. When
	// the caller passes a STRING that fully decodes as hex (and is an
	// even length), interpret it as WKB hex; otherwise read its UTF-8
	// bytes as-is. For BYTES, pass them straight through.
	var b []byte
	if sv, ok := args[0].(value.StringValue); ok {
		s := string(sv)
		if decoded, err := hex.DecodeString(s); err == nil {
			b = decoded
		} else {
			b = []byte(s)
		}
	} else {
		raw, err := args[0].ToBytes()
		if err != nil {
			return nil, err
		}
		b = raw
	}
	return geographyFromWKB(b)
}

// ----- GeoJSON serialization -----

func geographyToGeoJSON(g *value.GeographyValue) (string, error) {
	switch g.Kind() {
	case "POINT":
		lng, lat, _ := g.PointCoordinates()
		return fmt.Sprintf(`{"type":"Point","coordinates":[%s,%s]}`, gjFloat(lng), gjFloat(lat)), nil
	case "LINESTRING":
		pts, _ := g.LineStringPoints()
		return fmt.Sprintf(`{"type":"LineString","coordinates":%s}`, gjPoints(pts)), nil
	case "POLYGON":
		rings, _ := g.PolygonRings()
		return fmt.Sprintf(`{"type":"Polygon","coordinates":%s}`, gjRings(rings)), nil
	case "MULTIPOINT":
		pts, _ := g.MultiPointPoints()
		return fmt.Sprintf(`{"type":"MultiPoint","coordinates":%s}`, gjPoints(pts)), nil
	case "MULTILINESTRING":
		lines, _ := g.MultiLineStringLines()
		var parts []string
		for _, ls := range lines {
			parts = append(parts, gjPoints(ls))
		}
		return fmt.Sprintf(`{"type":"MultiLineString","coordinates":[%s]}`, strings.Join(parts, ",")), nil
	case "MULTIPOLYGON":
		polys, _ := g.MultiPolygonPolys()
		var parts []string
		for _, rings := range polys {
			parts = append(parts, gjRings(rings))
		}
		return fmt.Sprintf(`{"type":"MultiPolygon","coordinates":[%s]}`, strings.Join(parts, ",")), nil
	}
	return "", sqError("ST_ASGEOJSON", "unsupported geometry kind %q", g.Kind())
}

func gjFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func gjPoints(pts [][2]float64) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, p := range pts {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "[%s,%s]", gjFloat(p[0]), gjFloat(p[1]))
	}
	b.WriteByte(']')
	return b.String()
}

func gjRings(rings [][][2]float64) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, r := range rings {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(gjPoints(r))
	}
	b.WriteByte(']')
	return b.String()
}

func geographyFromGeoJSON(s string) (value.Value, error) {
	var doc map[string]any
	if err := jsonUnmarshal([]byte(s), &doc); err != nil {
		return nil, sqError("ST_GEOGFROMGEOJSON", "invalid JSON: %v", err)
	}
	return parseGeoJSON(doc)
}

func parseGeoJSON(doc map[string]any) (value.Value, error) {
	typ, _ := doc["type"].(string)
	coords, ok := doc["coordinates"]
	if !ok {
		return nil, sqError("ST_GEOGFROMGEOJSON", "missing coordinates")
	}
	switch typ {
	case "Point":
		c, err := readCoord(coords)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyPoint(c[0], c[1]), nil
	case "LineString":
		pts, err := readCoords(coords)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyLineString(pts), nil
	case "Polygon":
		rings, err := readRings(coords)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyPolygon(rings), nil
	case "MultiPoint":
		pts, err := readCoords(coords)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyMultiPoint(pts), nil
	case "MultiLineString":
		arr, ok := coords.([]any)
		if !ok {
			return nil, sqError("ST_GEOGFROMGEOJSON", "MultiLineString coords malformed")
		}
		lines := make([][][2]float64, 0, len(arr))
		for _, x := range arr {
			pts, err := readCoords(x)
			if err != nil {
				return nil, err
			}
			lines = append(lines, pts)
		}
		return value.NewGeographyMultiLineString(lines), nil
	case "MultiPolygon":
		arr, ok := coords.([]any)
		if !ok {
			return nil, sqError("ST_GEOGFROMGEOJSON", "MultiPolygon coords malformed")
		}
		polys := make([][][][2]float64, 0, len(arr))
		for _, x := range arr {
			rings, err := readRings(x)
			if err != nil {
				return nil, err
			}
			polys = append(polys, rings)
		}
		return value.NewGeographyMultiPolygon(polys), nil
	}
	return nil, sqError("ST_GEOGFROMGEOJSON", "unsupported type %q", typ)
}

func readCoord(v any) ([2]float64, error) {
	arr, ok := v.([]any)
	if !ok || len(arr) < 2 {
		return [2]float64{}, sqError("geojson", "expected [lng,lat]")
	}
	lng, ok1 := toFloat(arr[0])
	lat, ok2 := toFloat(arr[1])
	if !ok1 || !ok2 {
		return [2]float64{}, sqError("geojson", "coordinates must be numeric")
	}
	return [2]float64{lng, lat}, nil
}

func readCoords(v any) ([][2]float64, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, sqError("geojson", "expected array of coordinates")
	}
	out := make([][2]float64, 0, len(arr))
	for _, x := range arr {
		c, err := readCoord(x)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func readRings(v any) ([][][2]float64, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, sqError("geojson", "expected ring array")
	}
	out := make([][][2]float64, 0, len(arr))
	for _, r := range arr {
		ring, err := readCoords(r)
		if err != nil {
			return nil, err
		}
		out = append(out, ring)
	}
	return out, nil
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

// ----- WKB serialization -----

// Constants per OGC SFA-1 / ISO 19125.
const (
	wkbPoint              = 1
	wkbLineString         = 2
	wkbPolygon            = 3
	wkbMultiPoint         = 4
	wkbMultiLineString    = 5
	wkbMultiPolygon       = 6
	wkbGeometryCollection = 7
)

func geographyToWKB(g *value.GeographyValue) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeWKBGeometry(&buf, g); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeWKBGeometry(buf *bytes.Buffer, g *value.GeographyValue) error {
	endian := binary.LittleEndian
	buf.WriteByte(1) // little-endian byte order
	writeUint32 := func(v uint32) {
		var b [4]byte
		endian.PutUint32(b[:], v)
		buf.Write(b[:])
	}
	writeFloat64 := func(v float64) {
		var b [8]byte
		endian.PutUint64(b[:], math.Float64bits(v))
		buf.Write(b[:])
	}
	writePts := func(pts [][2]float64) {
		writeUint32(uint32(len(pts)))
		for _, p := range pts {
			writeFloat64(p[0])
			writeFloat64(p[1])
		}
	}
	switch g.Kind() {
	case "POINT":
		writeUint32(wkbPoint)
		lng, lat, _ := g.PointCoordinates()
		writeFloat64(lng)
		writeFloat64(lat)
	case "LINESTRING":
		writeUint32(wkbLineString)
		pts, _ := g.LineStringPoints()
		writePts(pts)
	case "POLYGON":
		writeUint32(wkbPolygon)
		rings, _ := g.PolygonRings()
		writeUint32(uint32(len(rings)))
		for _, r := range rings {
			writePts(r)
		}
	case "MULTIPOINT":
		writeUint32(wkbMultiPoint)
		pts, _ := g.MultiPointPoints()
		writeUint32(uint32(len(pts)))
		for _, p := range pts {
			if err := writeWKBGeometry(buf, value.NewGeographyPoint(p[0], p[1])); err != nil {
				return err
			}
		}
	case "MULTILINESTRING":
		writeUint32(wkbMultiLineString)
		lines, _ := g.MultiLineStringLines()
		writeUint32(uint32(len(lines)))
		for _, ls := range lines {
			if err := writeWKBGeometry(buf, value.NewGeographyLineString(ls)); err != nil {
				return err
			}
		}
	case "MULTIPOLYGON":
		writeUint32(wkbMultiPolygon)
		polys, _ := g.MultiPolygonPolys()
		writeUint32(uint32(len(polys)))
		for _, rings := range polys {
			if err := writeWKBGeometry(buf, value.NewGeographyPolygon(rings)); err != nil {
				return err
			}
		}
	default:
		return sqError("ST_ASBINARY", "unsupported geometry kind %q", g.Kind())
	}
	return nil
}

func geographyFromWKB(b []byte) (value.Value, error) {
	r := &wkbReader{data: b}
	return r.readGeometry()
}

type wkbReader struct {
	data []byte
	off  int
}

func (r *wkbReader) readByte() (byte, error) {
	if r.off >= len(r.data) {
		return 0, sqError("ST_GEOGFROMWKB", "unexpected EOF")
	}
	b := r.data[r.off]
	r.off++
	return b, nil
}

func (r *wkbReader) readUint32(le bool) (uint32, error) {
	if r.off+4 > len(r.data) {
		return 0, sqError("ST_GEOGFROMWKB", "short read")
	}
	var v uint32
	if le {
		v = binary.LittleEndian.Uint32(r.data[r.off:])
	} else {
		v = binary.BigEndian.Uint32(r.data[r.off:])
	}
	r.off += 4
	return v, nil
}

func (r *wkbReader) readFloat64(le bool) (float64, error) {
	if r.off+8 > len(r.data) {
		return 0, sqError("ST_GEOGFROMWKB", "short read")
	}
	var v uint64
	if le {
		v = binary.LittleEndian.Uint64(r.data[r.off:])
	} else {
		v = binary.BigEndian.Uint64(r.data[r.off:])
	}
	r.off += 8
	return math.Float64frombits(v), nil
}

func (r *wkbReader) readGeometry() (value.Value, error) {
	order, err := r.readByte()
	if err != nil {
		return nil, err
	}
	le := order == 1
	typ, err := r.readUint32(le)
	if err != nil {
		return nil, err
	}
	switch typ {
	case wkbPoint:
		lng, err := r.readFloat64(le)
		if err != nil {
			return nil, err
		}
		lat, err := r.readFloat64(le)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyPoint(lng, lat), nil
	case wkbLineString:
		pts, err := r.readPoints(le)
		if err != nil {
			return nil, err
		}
		return value.NewGeographyLineString(pts), nil
	case wkbPolygon:
		nRings, err := r.readUint32(le)
		if err != nil {
			return nil, err
		}
		rings := make([][][2]float64, nRings)
		for i := uint32(0); i < nRings; i++ {
			pts, err := r.readPoints(le)
			if err != nil {
				return nil, err
			}
			rings[i] = pts
		}
		return value.NewGeographyPolygon(rings), nil
	case wkbMultiPoint:
		n, err := r.readUint32(le)
		if err != nil {
			return nil, err
		}
		pts := make([][2]float64, n)
		for i := uint32(0); i < n; i++ {
			sub, err := r.readGeometry()
			if err != nil {
				return nil, err
			}
			gp, ok := sub.(*value.GeographyValue)
			if !ok {
				return nil, sqError("ST_GEOGFROMWKB", "invalid sub-geometry")
			}
			lng, lat, _ := gp.PointCoordinates()
			pts[i] = [2]float64{lng, lat}
		}
		return value.NewGeographyMultiPoint(pts), nil
	case wkbMultiLineString:
		n, err := r.readUint32(le)
		if err != nil {
			return nil, err
		}
		lines := make([][][2]float64, n)
		for i := uint32(0); i < n; i++ {
			sub, err := r.readGeometry()
			if err != nil {
				return nil, err
			}
			gv := sub.(*value.GeographyValue)
			ls, _ := gv.LineStringPoints()
			lines[i] = ls
		}
		return value.NewGeographyMultiLineString(lines), nil
	case wkbMultiPolygon:
		n, err := r.readUint32(le)
		if err != nil {
			return nil, err
		}
		polys := make([][][][2]float64, n)
		for i := uint32(0); i < n; i++ {
			sub, err := r.readGeometry()
			if err != nil {
				return nil, err
			}
			gv := sub.(*value.GeographyValue)
			pr, _ := gv.PolygonRings()
			polys[i] = pr
		}
		return value.NewGeographyMultiPolygon(polys), nil
	}
	return nil, sqError("ST_GEOGFROMWKB", "unsupported WKB type %d", typ)
}

func (r *wkbReader) readPoints(le bool) ([][2]float64, error) {
	n, err := r.readUint32(le)
	if err != nil {
		return nil, err
	}
	pts := make([][2]float64, n)
	for i := uint32(0); i < n; i++ {
		lng, err := r.readFloat64(le)
		if err != nil {
			return nil, err
		}
		lat, err := r.readFloat64(le)
		if err != nil {
			return nil, err
		}
		pts[i] = [2]float64{lng, lat}
	}
	return pts, nil
}
