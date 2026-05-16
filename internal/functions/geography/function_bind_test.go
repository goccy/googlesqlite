package geography

import (
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func Test_FunctionBind_bindStGeogPoint(t *testing.T) {
	t.Parallel()

	t.Run("BindStGeogPoint OK", func(t *testing.T) {
		t.Parallel()

		point, err := BindStGeogPoint(value.FloatValue(1), value.FloatValue(2))
		if err != nil {
			t.Fatal(err)
		}

		res, err := point.ToString()
		if err != nil {
			t.Fatal(err)
		}

		if res != "POINT (1 2)" {
			t.Fatalf("unexpected result: %s", res)
		}
	})

	t.Run("BindStGeogPoint bad arguments", func(t *testing.T) {
		t.Parallel()

		_, err := BindStGeogPoint(value.FloatValue(1))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func Test_FunctionBind_bindStGeogFromText(t *testing.T) {
	t.Parallel()

	t.Run("BindStGeogFromText OK", func(t *testing.T) {
		t.Parallel()

		point, err := BindStGeogFromText(value.StringValue("POINT (0 -49.23)"))
		if err != nil {
			t.Fatal(err)
		}

		res, err := point.ToString()
		if err != nil {
			t.Fatal(err)
		}

		if res != "POINT (0 -49.23)" {
			t.Fatalf("unexpected result: %s", res)
		}
	})

	t.Run("BindStGeogFromText bad arguments", func(t *testing.T) {
		t.Parallel()

		_, err := BindStGeogFromText(value.FloatValue(1))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func Test_FunctionBind_bindStDistance(t *testing.T) {
	t.Parallel()

	t.Run("StDistance OK", func(t *testing.T) {
		t.Parallel()

		dist, err := BindStDistance(value.NewGeographyPoint(1, 2), value.NewGeographyPoint(2, 3))
		if err != nil {
			t.Fatal(err)
		}

		res, err := dist.ToFloat64()
		if err != nil {
			t.Fatal(err)
		}

		if res <= 0 {
			t.Fatalf("unexpected result: %f", res)
		}
	})

	t.Run("BindStDistance bad arguments", func(t *testing.T) {
		t.Parallel()

		_, err := BindStDistance(value.FloatValue(1))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
