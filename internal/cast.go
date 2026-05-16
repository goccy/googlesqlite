package internal

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/goccy/googlesqlite/internal/value"
)

func CAST(expr value.Value, fromType, toType *Type, isSafeCast bool) (value.Value, error) {
	from, err := fromType.ToGoogleSQLType()
	if err != nil {
		return nil, fmt.Errorf("failed to get googlesql type from cast base type: %w", err)
	}
	to, err := toType.ToGoogleSQLType()
	if err != nil {
		return nil, fmt.Errorf("failed to get googlesql type from cast target type: %w", err)
	}
	fromValue, err := CastValue(from, expr)
	if err != nil {
		if isSafeCast {
			return nil, nil
		}
		return nil, err
	}
	casted, err := CastValue(to, fromValue)
	if err != nil {
		if isSafeCast {
			return nil, nil
		}
		return nil, err
	}
	return casted, nil
}

func bindCast(args ...value.Value) (value.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("CAST: invalid number of arguments: got %d, want 4", len(args))
	}
	jsonEncodedFromType, err := args[1].ToString()
	if err != nil {
		return nil, err
	}
	jsonEncodedToType, err := args[2].ToString()
	if err != nil {
		return nil, err
	}
	var fromType Type
	if err := json.Unmarshal([]byte(jsonEncodedFromType), &fromType); err != nil {
		return nil, err
	}
	var toType Type
	if err := json.Unmarshal([]byte(jsonEncodedToType), &toType); err != nil {
		return nil, err
	}
	isSafeCast, err := args[3].ToBool()
	if err != nil {
		return nil, err
	}
	return CAST(args[0], &fromType, &toType, isSafeCast)
}
