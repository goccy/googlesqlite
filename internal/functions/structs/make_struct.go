package structs

import (
	"fmt"

	"github.com/goccy/googlesqlite/internal/value"
)

func MAKE_STRUCT(args ...value.Value) (value.Value, error) {
	keys := make([]string, len(args)/2)
	values := make([]value.Value, len(args)/2)
	fieldMap := map[string]value.Value{}
	for i := 0; i < len(args)/2; i++ {
		key := args[i*2]
		val := args[i*2+1]
		k, err := key.ToString()
		if err != nil {
			return nil, err
		}
		keys[i] = k
		values[i] = val
		fieldMap[k] = val
	}
	return &value.StructValue{
		Keys:   keys,
		Values: values,
		M:      fieldMap,
	}, nil
}

func BindMakeStruct(args ...value.Value) (value.Value, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("MAKE_STRUCT: invalid number of arguments: got %d, want an even number of name/value pairs", len(args))
	}
	return MAKE_STRUCT(args...)
}
