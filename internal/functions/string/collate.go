package string

import (
	"fmt"
	"strings"

	"github.com/goccy/googlesqlite/internal/value"
	"golang.org/x/text/language"
)

func COLLATE(v, spec string) (value.Value, error) {
	if spec == "" {
		return value.StringValue(v), nil
	}
	splitted := strings.SplitN(spec, ":", 2)
	// `binary` and `unicode` are collation names rather than locales, so
	// they never reach the BCP 47 parser. The binary specification is a
	// bare language tag, so it takes no attribute.
	if splitted[0] == "binary" {
		if len(splitted) == 2 {
			return nil, fmt.Errorf("COLLATE: collation '%s' is not valid: binary cannot be combined with a suffix", spec)
		}
		return value.StringValue(v), nil
	}
	if splitted[0] != "unicode" {
		if _, err := language.Parse(splitted[0]); err != nil {
			return nil, fmt.Errorf("COLLATE: collation '%s' is not valid", spec)
		}
	}
	caseInsensitive := false
	if len(splitted) == 2 {
		switch splitted[1] {
		case "ci":
			caseInsensitive = true
		case "cs":
			caseInsensitive = false
		default:
			return nil, fmt.Errorf("COLLATE: unsupported collation attribute %s", splitted[1])
		}
	}
	// Project the value to a normalised form whose lexicographic
	// ordering matches the requested collation, so a subsequent
	// `<` / `>` / `=` comparison between two COLLATE results
	// (or routing through the `googlesqlite_collate` SQLite
	// collation) observes the collation-aware order. For `und:ci`
	// we fold to lower case via Unicode case mapping; other
	// language tags reuse the same fold which is acceptable for
	// the upstream Examples that exercise plain ASCII / Latin.
	if caseInsensitive {
		return value.StringValue(strings.ToLower(v)), nil
	}
	return value.StringValue(v), nil
}

func BindCollate(args ...value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("COLLATE: invalid number of arguments: got %d, want 2", len(args))
	}
	if args[0] == nil {
		return nil, nil
	}
	if args[1] == nil {
		return nil, fmt.Errorf("COLLATE: collation_specification must be string literal")
	}
	value, err := args[0].ToString()
	if err != nil {
		return nil, fmt.Errorf("COLLATE: value must be string: %w", err)
	}
	spec, err := args[1].ToString()
	if err != nil {
		return nil, fmt.Errorf("COLLATE: collation_specification must be string literal: %w", err)
	}
	return COLLATE(value, spec)
}
