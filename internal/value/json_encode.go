package value

import (
	"fmt"
	"strconv"
	"strings"
)

// EncodeJSON renders v with BigQuery's JSON encodings, for TO_JSON and
// TO_JSON_STRING. It differs from ToJSON only where ToJSON's text is also
// what the driver scans: DATE, DATETIME, TIME and TIMESTAMP are JSON strings
// ("2017-03-06"), at any depth inside an ARRAY or STRUCT.
// https://cloud.google.com/bigquery/docs/reference/standard-sql/json_functions#json_encodings
func EncodeJSON(v Value) (string, error) {
	switch vv := v.(type) {
	case nil:
		return "null", nil
	case DateValue, DatetimeValue, TimeValue, TimestampValue:
		s, err := vv.ToString()
		if err != nil {
			return "", err
		}
		return strconv.Quote(s), nil
	case *ArrayValue:
		elems := make([]string, 0, len(vv.Values))
		for _, e := range vv.Values {
			s, err := EncodeJSON(e)
			if err != nil {
				return "", err
			}
			elems = append(elems, s)
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ",")), nil
	case *StructValue:
		fields := make([]string, 0, len(vv.Keys))
		for i, key := range vv.Keys {
			s, err := EncodeJSON(vv.Values[i])
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("%s:%s", strconv.Quote(key), s))
		}
		return fmt.Sprintf("{%s}", strings.Join(fields, ",")), nil
	}
	return v.ToJSON()
}
