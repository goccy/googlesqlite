package value

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// GeographyValue follows the OGC Simple Features specification (SFS).
// Internally we carry the geometry as a tagged Go struct. The
// supported tags are POINT, LINESTRING, POLYGON, MULTIPOINT,
// MULTILINESTRING, MULTIPOLYGON, and GEOMETRYCOLLECTION.
//
// Pure-Go geometry implementations (without S2 / GEOS) cover
// round-trip storage, structural equality, and distance / intersection
// for Point–Point pairs. Full spherical computations on non-Point
// geometries return a clear "unsupported" error rather than wrong
// numbers.
type GeographyValue struct {
	g geographyType
}

// geographyType is implemented by every concrete geometry tag.
type geographyType interface {
	Kind() string
	ToWKT() (string, error)
	equal(other geographyType) bool
}

// ---- public constructors ----

func NewGeographyPoint(longitude, latitude float64) *GeographyValue {
	return &GeographyValue{g: &geographyPoint{
		longitude: geographyNormalizeLongitude(longitude),
		latitude:  latitude,
	}}
}

func NewGeographyLineString(points [][2]float64) *GeographyValue {
	normalised := make([][2]float64, len(points))
	for i, p := range points {
		normalised[i] = [2]float64{geographyNormalizeLongitude(p[0]), p[1]}
	}
	return &GeographyValue{g: &geographyLineString{points: normalised}}
}

func NewGeographyPolygon(rings [][][2]float64) *GeographyValue {
	out := make([][][2]float64, len(rings))
	for i, ring := range rings {
		ringOut := make([][2]float64, len(ring))
		for j, p := range ring {
			ringOut[j] = [2]float64{geographyNormalizeLongitude(p[0]), p[1]}
		}
		out[i] = ringOut
	}
	return &GeographyValue{g: &geographyPolygon{rings: out}}
}

func NewGeographyMultiPoint(points [][2]float64) *GeographyValue {
	out := make([][2]float64, len(points))
	for i, p := range points {
		out[i] = [2]float64{geographyNormalizeLongitude(p[0]), p[1]}
	}
	return &GeographyValue{g: &geographyMultiPoint{points: out}}
}

func NewGeographyMultiLineString(lines [][][2]float64) *GeographyValue {
	out := make([][][2]float64, len(lines))
	for i, ls := range lines {
		copyLs := make([][2]float64, len(ls))
		for j, p := range ls {
			copyLs[j] = [2]float64{geographyNormalizeLongitude(p[0]), p[1]}
		}
		out[i] = copyLs
	}
	return &GeographyValue{g: &geographyMultiLineString{lines: out}}
}

func NewGeographyMultiPolygon(polys [][][][2]float64) *GeographyValue {
	out := make([][][][2]float64, len(polys))
	for i, poly := range polys {
		ringsOut := make([][][2]float64, len(poly))
		for j, ring := range poly {
			ringOut := make([][2]float64, len(ring))
			for k, p := range ring {
				ringOut[k] = [2]float64{geographyNormalizeLongitude(p[0]), p[1]}
			}
			ringsOut[j] = ringOut
		}
		out[i] = ringsOut
	}
	return &GeographyValue{g: &geographyMultiPolygon{polys: out}}
}

func NewGeographyCollection(parts []*GeographyValue) *GeographyValue {
	return &GeographyValue{g: &geographyCollection{parts: parts}}
}

// GeographyFromWKT parses a WKT string into a GeographyValue.
// Supports POINT / LINESTRING / POLYGON / MULTIPOINT /
// MULTILINESTRING / MULTIPOLYGON / GEOMETRYCOLLECTION.
func GeographyFromWKT(wkt string) (*GeographyValue, error) {
	p := &wktParser{input: strings.TrimSpace(wkt)}
	g, err := p.parseGeometry()
	if err != nil {
		return nil, fmt.Errorf("geography WKT %q: %w", wkt, err)
	}
	p.skipSpaces()
	if p.pos != len(p.input) {
		return nil, fmt.Errorf("geography WKT %q: trailing content %q", wkt, p.input[p.pos:])
	}
	return g, nil
}

// ---- Value interface noise ----

func (g *GeographyValue) Add(_ Value) (Value, error) {
	return nil, fmt.Errorf("unsupported add operator for geography value")
}
func (g *GeographyValue) Sub(_ Value) (Value, error) {
	return nil, fmt.Errorf("unsupported sub operator for geography value")
}
func (g *GeographyValue) Mul(_ Value) (Value, error) {
	return nil, fmt.Errorf("unsupported mul operator for geography value")
}
func (g *GeographyValue) Div(_ Value) (Value, error) {
	return nil, fmt.Errorf("unsupported div operator for geography value")
}

// EQ implements structural equality for geography. Two geographies
// compare equal when they have the same kind and identical
// coordinates in declaration order. SFS spatial-equality (same
// point set, possibly different orderings) is more permissive but
// requires S2 / GEOS to compute robustly; structural equality is
// what every emulator-side workload here expects.
func (g *GeographyValue) EQ(other Value) (bool, error) {
	o, ok := other.(*GeographyValue)
	if !ok {
		return false, fmt.Errorf("geography EQ: expected GeographyValue, got %T", other)
	}
	if g == nil || o == nil {
		return g == o, nil
	}
	return g.g.equal(o.g), nil
}

func (g *GeographyValue) GT(_ Value) (bool, error) {
	return false, fmt.Errorf("unsupported gt operator for geography value")
}
func (g *GeographyValue) GTE(_ Value) (bool, error) {
	return false, fmt.Errorf("unsupported gte operator for geography value")
}
func (g *GeographyValue) LT(_ Value) (bool, error) {
	return false, fmt.Errorf("unsupported lt operator for geography value")
}
func (g *GeographyValue) LTE(_ Value) (bool, error) {
	return false, fmt.Errorf("unsupported lte operator for geography value")
}

func (g *GeographyValue) ToInt64() (int64, error) {
	return 0, fmt.Errorf("unsupported ToInt64 operator for geography value")
}

func (g *GeographyValue) ToWKT() (string, error) {
	if g == nil || g.g == nil {
		return "", nil
	}
	return g.g.ToWKT()
}

func (g *GeographyValue) ToString() (string, error) { return g.ToWKT() }
func (g *GeographyValue) String() (string, error)   { return g.ToWKT() }

func (g *GeographyValue) ToBytes() ([]byte, error) {
	v, err := g.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (g *GeographyValue) ToFloat64() (float64, error) {
	return 0, fmt.Errorf("unsupported ToFloat64 operator for geography value")
}
func (g *GeographyValue) ToBool() (bool, error) {
	return false, fmt.Errorf("unsupported ToBool operator for geography value")
}
func (g *GeographyValue) ToArray() (*ArrayValue, error) {
	return nil, fmt.Errorf("unsupported ToArray operator for geography value")
}
func (g *GeographyValue) ToStruct() (*StructValue, error) {
	return nil, fmt.Errorf("unsupported ToStruct operator for geography value")
}

func (g *GeographyValue) ToJSON() (string, error) {
	s, err := g.ToString()
	if err != nil {
		return "", err
	}
	return strconv.Quote(s), nil
}

func (g *GeographyValue) ToTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("unsupported ToTime operator for geography value")
}
func (g *GeographyValue) ToRat() (*big.Rat, error) {
	return nil, fmt.Errorf("unsupported ToRat operator for geography value")
}

func (g *GeographyValue) Format(verb rune) string {
	str, err := g.ToString()
	if err != nil {
		return "error"
	}
	switch verb {
	case 't':
		return str
	case 'T':
		return fmt.Sprintf(`GEOGRAPHY %q`, str)
	}
	return str
}

func (g *GeographyValue) Interface() any {
	s, err := g.ToString()
	if err != nil {
		return nil
	}
	return s
}

// PointCoordinates returns (longitude, latitude) when the geography
// is a Point. Returns ok=false otherwise (including POINT EMPTY,
// which carries no coordinates). Used by ST_X / ST_Y.
func (g *GeographyValue) PointCoordinates() (float64, float64, bool) {
	if g == nil {
		return 0, 0, false
	}
	p, ok := g.g.(*geographyPoint)
	if !ok {
		return 0, 0, false
	}
	if p.empty {
		return 0, 0, false
	}
	return p.longitude, p.latitude, true
}

// EqualPoint returns true when both geographies are points and have
// equal coordinates.
func (g *GeographyValue) EqualPoint(other *GeographyValue) bool {
	if g == nil || other == nil {
		return false
	}
	a, _ := g.g.(*geographyPoint)
	b, _ := other.g.(*geographyPoint)
	if a == nil || b == nil {
		return false
	}
	return a.longitude == b.longitude && a.latitude == b.latitude
}

// Kind returns the WKT-style geometry tag (POINT / LINESTRING / ...).
// Returns "" for a nil GeographyValue.
func (g *GeographyValue) Kind() string {
	if g == nil || g.g == nil {
		return ""
	}
	return g.g.Kind()
}

// LineStringPoints returns the ordered points of a LINESTRING.
// Returns ok=false for any other geometry kind.
func (g *GeographyValue) LineStringPoints() ([][2]float64, bool) {
	if g == nil {
		return nil, false
	}
	ls, ok := g.g.(*geographyLineString)
	if !ok {
		return nil, false
	}
	return ls.points, true
}

// PolygonRings returns the rings of a POLYGON (outer ring first,
// inner rings second). Returns ok=false for any other geometry
// kind.
func (g *GeographyValue) PolygonRings() ([][][2]float64, bool) {
	if g == nil {
		return nil, false
	}
	p, ok := g.g.(*geographyPolygon)
	if !ok {
		return nil, false
	}
	return p.rings, true
}

// MultiPointPoints returns the points of a MULTIPOINT geography.
// Returns ok=false otherwise.
func (g *GeographyValue) MultiPointPoints() ([][2]float64, bool) {
	if g == nil {
		return nil, false
	}
	mp, ok := g.g.(*geographyMultiPoint)
	if !ok {
		return nil, false
	}
	return mp.points, true
}

// MultiLineStringLines returns the individual line strings of a
// MULTILINESTRING. Returns ok=false otherwise.
func (g *GeographyValue) MultiLineStringLines() ([][][2]float64, bool) {
	if g == nil {
		return nil, false
	}
	mls, ok := g.g.(*geographyMultiLineString)
	if !ok {
		return nil, false
	}
	return mls.lines, true
}

// MultiPolygonPolys returns the polygons of a MULTIPOLYGON, each
// represented as its ring list (outer then inner).
func (g *GeographyValue) MultiPolygonPolys() ([][][][2]float64, bool) {
	if g == nil {
		return nil, false
	}
	mp, ok := g.g.(*geographyMultiPolygon)
	if !ok {
		return nil, false
	}
	return mp.polys, true
}

// CollectionParts returns the constituent geographies of a
// GEOMETRYCOLLECTION. Returns ok=false otherwise.
func (g *GeographyValue) CollectionParts() ([]*GeographyValue, bool) {
	if g == nil {
		return nil, false
	}
	c, ok := g.g.(*geographyCollection)
	if !ok {
		return nil, false
	}
	return c.parts, true
}

// IsEmpty reports whether the geography is one of the WKT EMPTY
// sentinels (POINT EMPTY, LINESTRING EMPTY, ...): no vertices to
// participate in distance / area / bounding-box computation.
// `fullglobe` is NOT empty; it has zero rings but represents the
// entire surface of Earth.
func (g *GeographyValue) IsEmpty() bool {
	if g == nil || g.g == nil {
		return true
	}
	switch x := g.g.(type) {
	case *geographyFullGlobe:
		_ = x
		return false
	case *geographyPoint:
		return x.empty
	case *geographyLineString:
		return len(x.points) == 0
	case *geographyPolygon:
		return len(x.rings) == 0
	case *geographyMultiPoint:
		return len(x.points) == 0
	case *geographyMultiLineString:
		return len(x.lines) == 0
	case *geographyMultiPolygon:
		return len(x.polys) == 0
	case *geographyCollection:
		if len(x.parts) == 0 {
			return true
		}
		for _, p := range x.parts {
			if !p.IsEmpty() {
				return false
			}
		}
		return true
	}
	return false
}

// ---- concrete geometries ----

// geographyFullGlobe represents the GoogleSQL `fullglobe` literal —
// the entire surface of Earth as a single geography. No vertices,
// no rings; downstream consumers detect it via Kind() == "FULLGLOBE".
type geographyFullGlobe struct{}

func (g *geographyFullGlobe) Kind() string { return "FULLGLOBE" }
func (g *geographyFullGlobe) ToWKT() (string, error) {
	return "fullglobe", nil
}
func (g *geographyFullGlobe) equal(o geographyType) bool {
	_, ok := o.(*geographyFullGlobe)
	return ok
}

type geographyPoint struct {
	longitude float64
	latitude  float64
	empty     bool
}

func (g *geographyPoint) Kind() string { return "POINT" }
func (g *geographyPoint) ToWKT() (string, error) {
	if g.empty {
		return "POINT EMPTY", nil
	}
	return fmt.Sprintf("POINT (%s %s)", formatCoord(g.longitude), formatCoord(g.latitude)), nil
}
func (g *geographyPoint) equal(o geographyType) bool {
	other, ok := o.(*geographyPoint)
	if !ok {
		return false
	}
	if g.empty != other.empty {
		return false
	}
	if g.empty {
		return true
	}
	return g.longitude == other.longitude && g.latitude == other.latitude
}

type geographyLineString struct {
	points [][2]float64
}

func (g *geographyLineString) Kind() string { return "LINESTRING" }
func (g *geographyLineString) ToWKT() (string, error) {
	return "LINESTRING " + formatCoordList(g.points), nil
}
func (g *geographyLineString) equal(o geographyType) bool {
	other, ok := o.(*geographyLineString)
	if !ok {
		return false
	}
	return coordSliceEqual(g.points, other.points)
}

type geographyPolygon struct {
	rings [][][2]float64
	// inverted is true when the polygon was parsed with oriented=TRUE
	// AND its outer ring is clockwise (signed spherical area negative),
	// which by the upstream convention means the geometry's interior is
	// the *exterior* of the ring — the small region cut out of the
	// rest of the globe. ST_BOUNDINGBOX uses this to widen the box to
	// the entire globe; other consumers fall back to the ring's
	// vertices unchanged.
	inverted bool
}

// Inverted reports whether the polygon was parsed in oriented mode
// with a clockwise outer ring (interior is the globe minus the
// ring's enclosed region).
func (g *GeographyValue) Inverted() bool {
	p, ok := g.g.(*geographyPolygon)
	return ok && p.inverted
}

// MarkInverted is called by the WKT parser (or any caller that knows
// the oriented=TRUE convention) to flag the polygon as representing
// the complement of its enclosed region.
func (g *GeographyValue) MarkInverted() {
	if p, ok := g.g.(*geographyPolygon); ok {
		p.inverted = true
	}
}

func (g *geographyPolygon) Kind() string { return "POLYGON" }
func (g *geographyPolygon) ToWKT() (string, error) {
	if len(g.rings) == 0 {
		return "POLYGON EMPTY", nil
	}
	parts := make([]string, len(g.rings))
	for i, r := range g.rings {
		parts[i] = formatCoordList(r)
	}
	return "POLYGON (" + strings.Join(parts, ", ") + ")", nil
}
func (g *geographyPolygon) equal(o geographyType) bool {
	other, ok := o.(*geographyPolygon)
	if !ok || len(g.rings) != len(other.rings) {
		return false
	}
	for i := range g.rings {
		if !coordSliceEqual(g.rings[i], other.rings[i]) {
			return false
		}
	}
	return true
}

type geographyMultiPoint struct {
	points [][2]float64
}

func (g *geographyMultiPoint) Kind() string { return "MULTIPOINT" }
func (g *geographyMultiPoint) ToWKT() (string, error) {
	return "MULTIPOINT " + formatCoordList(g.points), nil
}
func (g *geographyMultiPoint) equal(o geographyType) bool {
	other, ok := o.(*geographyMultiPoint)
	if !ok {
		return false
	}
	return coordSliceEqual(g.points, other.points)
}

type geographyMultiLineString struct {
	lines [][][2]float64
}

func (g *geographyMultiLineString) Kind() string { return "MULTILINESTRING" }
func (g *geographyMultiLineString) ToWKT() (string, error) {
	if len(g.lines) == 0 {
		// Per the OGC SF / WKT v1.2 spec, empty multi-geometries use
		// the literal "EMPTY" sentinel instead of an empty paren list.
		// The downstream WKT parser only accepts the spec form, so
		// emitting "MULTILINESTRING ()" round-trips as a parse error.
		return "MULTILINESTRING EMPTY", nil
	}
	parts := make([]string, len(g.lines))
	for i, ls := range g.lines {
		parts[i] = formatCoordList(ls)
	}
	return "MULTILINESTRING (" + strings.Join(parts, ", ") + ")", nil
}
func (g *geographyMultiLineString) equal(o geographyType) bool {
	other, ok := o.(*geographyMultiLineString)
	if !ok || len(g.lines) != len(other.lines) {
		return false
	}
	for i := range g.lines {
		if !coordSliceEqual(g.lines[i], other.lines[i]) {
			return false
		}
	}
	return true
}

type geographyMultiPolygon struct {
	polys [][][][2]float64
}

func (g *geographyMultiPolygon) Kind() string { return "MULTIPOLYGON" }
func (g *geographyMultiPolygon) ToWKT() (string, error) {
	if len(g.polys) == 0 {
		return "MULTIPOLYGON EMPTY", nil
	}
	parts := make([]string, len(g.polys))
	for i, poly := range g.polys {
		ringStrs := make([]string, len(poly))
		for j, ring := range poly {
			ringStrs[j] = formatCoordList(ring)
		}
		parts[i] = "(" + strings.Join(ringStrs, ", ") + ")"
	}
	return "MULTIPOLYGON (" + strings.Join(parts, ", ") + ")", nil
}
func (g *geographyMultiPolygon) equal(o geographyType) bool {
	other, ok := o.(*geographyMultiPolygon)
	if !ok || len(g.polys) != len(other.polys) {
		return false
	}
	for i := range g.polys {
		if len(g.polys[i]) != len(other.polys[i]) {
			return false
		}
		for j := range g.polys[i] {
			if !coordSliceEqual(g.polys[i][j], other.polys[i][j]) {
				return false
			}
		}
	}
	return true
}

type geographyCollection struct {
	parts []*GeographyValue
}

func (g *geographyCollection) Kind() string { return "GEOMETRYCOLLECTION" }
func (g *geographyCollection) ToWKT() (string, error) {
	parts := make([]string, len(g.parts))
	for i, p := range g.parts {
		s, err := p.ToWKT()
		if err != nil {
			return "", err
		}
		parts[i] = s
	}
	return "GEOMETRYCOLLECTION (" + strings.Join(parts, ", ") + ")", nil
}
func (g *geographyCollection) equal(o geographyType) bool {
	other, ok := o.(*geographyCollection)
	if !ok || len(g.parts) != len(other.parts) {
		return false
	}
	for i := range g.parts {
		if !g.parts[i].g.equal(other.parts[i].g) {
			return false
		}
	}
	return true
}

// ---- WKT formatting helpers ----

func formatCoord(f float64) string {
	// Use %.15g so a 1-ULP-below-1.0 value (0.9999999999999998) folds
	// to "1" the way BigQuery prints it. -1 precision gives Go's
	// "shortest unique" form which exposes that ULP and breaks
	// upstream-display compatibility.
	s := strconv.FormatFloat(f, 'g', 15, 64)
	// `g` may produce values like "1e+02" for 100; coerce back to a
	// plain fixed-point shape (`100`) to match the upstream WKT
	// renderer's appearance.
	if !strings.ContainsAny(s, "eE") {
		return s
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// formatCoordList returns "(lon lat, lon lat, ...)".
func formatCoordList(points [][2]float64) string {
	parts := make([]string, len(points))
	for i, p := range points {
		parts[i] = formatCoord(p[0]) + " " + formatCoord(p[1])
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func coordSliceEqual(a, b [][2]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// geographyNormalizeLongitude wraps longitudes outside [-180, 180].
func geographyNormalizeLongitude(longitude float64) float64 {
	if -180 <= longitude && longitude <= 180 {
		return longitude
	}
	longitude = math.Mod(longitude+180, 360)
	if longitude < 0 {
		longitude += 360
	}
	return longitude - 180
}

// ---- WKT parser ----

type wktParser struct {
	input string
	pos   int
}

func (p *wktParser) parseGeometry() (*GeographyValue, error) {
	p.skipSpaces()
	tag := p.readTag()
	if tag == "" {
		return nil, errors.New("empty geometry tag")
	}
	// GoogleSQL allows the literal `fullglobe` to denote the full
	// surface of the Earth as a geography. Model it as a polygon
	// whose ring set is empty; downstream geometry operations treat
	// it as "no boundary, all points are inside".
	if tag == "FULLGLOBE" {
		return &GeographyValue{g: &geographyFullGlobe{}}, nil
	}
	// `<TAG> EMPTY` is the WKT v1.2 empty-geometry sentinel. Match it
	// before reading the paren-list so empty multi-geometries
	// round-trip through ToWKT / GeographyFromWKT.
	if savePos := p.pos; consumeKeyword(p, "EMPTY") {
		switch tag {
		case "POINT":
			return &GeographyValue{g: &geographyPoint{empty: true}}, nil
		case "LINESTRING":
			return NewGeographyLineString(nil), nil
		case "POLYGON":
			return NewGeographyPolygon(nil), nil
		case "MULTIPOINT":
			return NewGeographyMultiPoint(nil), nil
		case "MULTILINESTRING":
			return NewGeographyMultiLineString(nil), nil
		case "MULTIPOLYGON":
			return NewGeographyMultiPolygon(nil), nil
		case "GEOMETRYCOLLECTION":
			return NewGeographyCollection(nil), nil
		}
		p.pos = savePos
	}
	switch tag {
	case "POINT":
		coords, err := p.readCoordList()
		if err != nil {
			return nil, err
		}
		if len(coords) != 1 {
			return nil, fmt.Errorf("POINT must contain exactly one coordinate, got %d", len(coords))
		}
		return NewGeographyPoint(coords[0][0], coords[0][1]), nil
	case "LINESTRING":
		coords, err := p.readCoordList()
		if err != nil {
			return nil, err
		}
		return NewGeographyLineString(coords), nil
	case "POLYGON":
		rings, err := p.readCoordRings()
		if err != nil {
			return nil, err
		}
		return NewGeographyPolygon(rings), nil
	case "MULTIPOINT":
		coords, err := p.readMultiPointCoords()
		if err != nil {
			return nil, err
		}
		return NewGeographyMultiPoint(coords), nil
	case "MULTILINESTRING":
		lines, err := p.readCoordRings()
		if err != nil {
			return nil, err
		}
		return NewGeographyMultiLineString(lines), nil
	case "MULTIPOLYGON":
		polys, err := p.readPolygonList()
		if err != nil {
			return nil, err
		}
		return NewGeographyMultiPolygon(polys), nil
	case "GEOMETRYCOLLECTION":
		parts, err := p.readCollection()
		if err != nil {
			return nil, err
		}
		return NewGeographyCollection(parts), nil
	}
	return nil, fmt.Errorf("unsupported geometry tag %q", tag)
}

// consumeKeyword tries to read `kw` (case-insensitive) at the current
// position, advancing past it on success and leaving p untouched on
// failure. Used for the WKT v1.2 EMPTY sentinel.
func consumeKeyword(p *wktParser, kw string) bool {
	p.skipSpaces()
	if p.pos+len(kw) > len(p.input) {
		return false
	}
	if !strings.EqualFold(p.input[p.pos:p.pos+len(kw)], kw) {
		return false
	}
	end := p.pos + len(kw)
	// Ensure the keyword stands alone (followed by whitespace, end,
	// or a paren / comma), so e.g. "EMPTYISH" is not mistaken.
	if end < len(p.input) {
		switch p.input[end] {
		case ' ', '\t', '\n', '\r', '(', ')', ',':
		default:
			return false
		}
	}
	p.pos = end
	return true
}

func (p *wktParser) skipSpaces() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *wktParser) readTag() string {
	p.skipSpaces()
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			p.pos++
			continue
		}
		break
	}
	return strings.ToUpper(p.input[start:p.pos])
}

// expect advances past a single literal byte after skipping spaces.
func (p *wktParser) expect(c byte) error {
	p.skipSpaces()
	if p.pos >= len(p.input) || p.input[p.pos] != c {
		return fmt.Errorf("expected %q at offset %d", c, p.pos)
	}
	p.pos++
	return nil
}

func (p *wktParser) peek() byte {
	p.skipSpaces()
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *wktParser) readNumber() (float64, error) {
	p.skipSpaces()
	start := p.pos
	if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
		p.pos++
	}
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if (c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
			p.pos++
			continue
		}
		break
	}
	s := p.input[start:p.pos]
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return f, nil
}

// readCoord reads one (lon lat) pair without surrounding parens.
func (p *wktParser) readCoord() ([2]float64, error) {
	lon, err := p.readNumber()
	if err != nil {
		return [2]float64{}, err
	}
	lat, err := p.readNumber()
	if err != nil {
		return [2]float64{}, err
	}
	return [2]float64{lon, lat}, nil
}

// readCoordList reads "(c, c, ...)" where each c is "lon lat".
func (p *wktParser) readCoordList() ([][2]float64, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var coords [][2]float64
	for {
		c, err := p.readCoord()
		if err != nil {
			return nil, err
		}
		coords = append(coords, c)
		p.skipSpaces()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return coords, nil
}

// readMultiPointCoords accepts both forms WKT permits:
//
//	MULTIPOINT (1 2, 3 4)
//	MULTIPOINT ((1 2), (3 4))
func (p *wktParser) readMultiPointCoords() ([][2]float64, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var coords [][2]float64
	for {
		p.skipSpaces()
		if p.peek() == '(' {
			inner, err := p.readCoordList()
			if err != nil {
				return nil, err
			}
			if len(inner) != 1 {
				return nil, fmt.Errorf("MULTIPOINT inner group must have exactly one point, got %d", len(inner))
			}
			coords = append(coords, inner[0])
		} else {
			c, err := p.readCoord()
			if err != nil {
				return nil, err
			}
			coords = append(coords, c)
		}
		p.skipSpaces()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return coords, nil
}

// readCoordRings reads "((c, c, ...), (c, c, ...), ...)" — used by
// POLYGON (rings) and MULTILINESTRING.
func (p *wktParser) readCoordRings() ([][][2]float64, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var rings [][][2]float64
	for {
		coords, err := p.readCoordList()
		if err != nil {
			return nil, err
		}
		rings = append(rings, coords)
		p.skipSpaces()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return rings, nil
}

// readPolygonList reads "(((..),..), ((..),..))" for MULTIPOLYGON.
func (p *wktParser) readPolygonList() ([][][][2]float64, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var polys [][][][2]float64
	for {
		rings, err := p.readCoordRings()
		if err != nil {
			return nil, err
		}
		polys = append(polys, rings)
		p.skipSpaces()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return polys, nil
}

// readCollection reads "(geom, geom, ...)" — recursive geometries.
func (p *wktParser) readCollection() ([]*GeographyValue, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var parts []*GeographyValue
	for {
		g, err := p.parseGeometry()
		if err != nil {
			return nil, err
		}
		parts = append(parts, g)
		p.skipSpaces()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return parts, nil
}

// ---- distance ----

// DistanceTo returns the great-circle (haversine) distance in meters
// between two GeographyValue points. Non-Point geometries return an
// error — full polygon / line distance needs spherical geometry
// (S2) which is out of scope for the pure-Go build.
func (g *GeographyValue) DistanceTo(other *GeographyValue) (float64, error) {
	if g == nil || other == nil {
		return 0, fmt.Errorf("nil geography")
	}
	p1, ok := g.g.(*geographyPoint)
	if !ok {
		return 0, fmt.Errorf("unsupported geography type %s for ST_DISTANCE (only POINT supported)", g.Kind())
	}
	p2, ok := other.g.(*geographyPoint)
	if !ok {
		return 0, fmt.Errorf("unsupported geography type %s for ST_DISTANCE (only POINT supported)", other.Kind())
	}
	return haversineDistance(p1.latitude, p1.longitude, p2.latitude, p2.longitude), nil
}
