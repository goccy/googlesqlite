package geography

import (
	"math"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func Test_FunctionGeography_ST_GEOGPOINT(t *testing.T) {
	t.Parallel()

	t.Run("ST_GEOGPOINT OK", func(t *testing.T) {
		t.Parallel()

		v, err := ST_GEOGPOINT(10, 20)
		if err != nil {
			t.Fatal(err)
		}

		res, err := v.ToString()
		if err != nil {
			t.Fatal(err)
		}

		if res != "POINT (10 20)" {
			t.Fatalf("unexpected result: %s", res)
		}
	})

	t.Run("ST_GEOGPOINT latitude boundaries", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			lat float64
		}{
			{-90},
			{90},
		}

		for _, tt := range tests {
			_, err := ST_GEOGPOINT(0, tt.lat)
			if err != nil {
				t.Fatalf("unexpected error for latitude %f: %v", tt.lat, err)
			}
		}
	})

	t.Run("ST_GEOGPOINT invalid latitude", func(t *testing.T) {
		t.Parallel()

		tests := []float64{
			-90.0001,
			90.0001,
			100,
			-100,
		}

		for _, lat := range tests {
			_, err := ST_GEOGPOINT(0, lat)
			if err == nil {
				t.Fatalf("expected error for latitude %f", lat)
			}
		}
	})

	t.Run("ST_GEOGPOINT normalizes longitude", func(t *testing.T) {
		t.Parallel()

		v, err := ST_GEOGPOINT(190, 0)
		if err != nil {
			t.Fatal(err)
		}

		res, err := v.ToString()
		if err != nil {
			t.Fatal(err)
		}

		// 190 -> -170
		if res != "POINT (-170 0)" {
			t.Fatalf("unexpected result: %s", res)
		}
	})
}

func Test_FunctionGeography_ST_GEOGFROMTEXT(t *testing.T) {
	t.Parallel()

	t.Run("ST_GEOGFROMTEXT OK", func(t *testing.T) {
		t.Parallel()

		val, err := ST_GEOGFROMTEXT("POINT(10 24.3)")
		if err != nil {
			t.Fatal(err)
		}

		res, err := val.ToString()
		if err != nil {
			t.Fatal(err)
		}

		if res != "POINT (10 24.3)" {
			t.Fatalf("unexpected result: %s", res)
		}
	})

	t.Run("ST_GEOGFROMTEXT LINESTRING round-trip", func(t *testing.T) {
		t.Parallel()
		v, err := ST_GEOGFROMTEXT("LINESTRING (0 0, 1 1, 2 2)")
		if err != nil {
			t.Fatal(err)
		}
		geo, ok := v.(*value.GeographyValue)
		if !ok {
			t.Fatalf("unexpected type %T", v)
		}
		got, err := geo.ToWKT()
		if err != nil {
			t.Fatal(err)
		}
		if got != "LINESTRING (0 0, 1 1, 2 2)" {
			t.Fatalf("unexpected WKT: %s", got)
		}
	})

	t.Run("ST_GEOGFROMTEXT invalid WKT (unknown tag)", func(t *testing.T) {
		t.Parallel()
		_, err := ST_GEOGFROMTEXT("WAFFLE (0 0)")
		if err == nil {
			t.Fatal("expected error for unknown geometry tag")
		}
	})
}

func Test_FunctionGeography_ST_DISTANCE(t *testing.T) {
	t.Parallel()

	t.Run("ST_DISTANCE OK, > 0", func(t *testing.T) {
		t.Parallel()

		p1 := value.NewGeographyPoint(100, 90)
		p2 := value.NewGeographyPoint(100.03, 89.999)

		val, err := ST_DISTANCE(p1, p2)
		if err != nil {
			t.Fatal(err)
		}

		dist, err := val.ToFloat64()
		if err != nil {
			t.Fatal(err)
		}

		testGeographyAssertDistanceEqual(t, dist, 111.19510117719409)
	})

	t.Run("ST_DISTANCE OK, 0 distance", func(t *testing.T) {
		t.Parallel()

		p1 := value.NewGeographyPoint(100.03, 89.999)
		p2 := value.NewGeographyPoint(100.03, 89.999)

		val, err := ST_DISTANCE(p1, p2)
		if err != nil {
			t.Fatal(err)
		}

		dist, err := val.ToFloat64()
		if err != nil {
			t.Fatal(err)
		}

		testGeographyAssertDistanceEqual(t, dist, 0)
	})

	t.Run("ST_DISTANCE bad arguments", func(t *testing.T) {
		t.Parallel()

		_, err := ST_DISTANCE(nil, nil)
		if err == nil {
			t.Fatal(err)
		}
	})
}

func testGeographyAssertDistanceEqual(t *testing.T, dist, expected float64) {
	const threshold = 5.0

	if math.Abs(dist-expected) >= threshold {
		t.Fatalf("expected distance close to %f meters, got %f", expected, dist)
	}
}
