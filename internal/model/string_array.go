package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// StringArray maps a PostgreSQL text[] column to []string.
//
// A named []string already encodes to and decodes from JSON as a normal array,
// so the type only has to teach database/sql how to read and write the
// PostgreSQL array literal format (e.g. `{foo,"bar,baz"}`).
type StringArray []string

// Value implements driver.Valuer.
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, s := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		// Every element is quoted: that is valid for any text value and
		// avoids having to decide which characters force quoting.
		b.WriteByte('"')
		for j := 0; j < len(s); j++ {
			if s[j] == '"' || s[j] == '\\' {
				b.WriteByte('\\')
			}
			b.WriteByte(s[j])
		}
		b.WriteByte('"')
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Scan implements sql.Scanner. pgx (the driver behind gorm.io/driver/postgres)
// hands text[] columns to database/sql as their text representation, so both
// string and []byte are accepted here.
func (a *StringArray) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*a = nil
		return nil
	case string:
		return a.scanText(src)
	case []byte:
		return a.scanText(string(src))
	default:
		return fmt.Errorf("model: cannot convert %T to StringArray", src)
	}
}

func (a *StringArray) scanText(src string) error {
	elems, err := parseTextArray(src)
	if err != nil {
		return err
	}
	*a = elems
	return nil
}

// parseTextArray parses the one-dimensional text array format emitted by
// PostgreSQL. Elements are comma separated, may be double quoted, and use a
// backslash (or a doubled quote) to escape characters inside quotes.
func parseTextArray(src string) (StringArray, error) {
	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		return nil, fmt.Errorf("model: cannot parse %q as PostgreSQL text array", src)
	}

	body := src[1 : len(src)-1]
	if body == "" {
		return StringArray{}, nil
	}

	elems := StringArray{}
	for i := 0; ; {
		if i >= len(body) {
			return nil, fmt.Errorf("model: trailing comma in PostgreSQL text array %q", src)
		}

		if body[i] == '"' {
			var b strings.Builder
			i++
			closed := false
			for i < len(body) {
				c := body[i]
				switch {
				case c == '\\':
					if i+1 >= len(body) {
						return nil, fmt.Errorf("model: dangling escape in PostgreSQL text array %q", src)
					}
					b.WriteByte(body[i+1])
					i += 2
				case c == '"':
					if i+1 < len(body) && body[i+1] == '"' {
						// Doubled quote inside a quoted element is a literal quote.
						b.WriteByte('"')
						i += 2
						continue
					}
					i++
					closed = true
				default:
					b.WriteByte(c)
					i++
				}
				if closed {
					break
				}
			}
			if !closed {
				return nil, fmt.Errorf("model: unterminated quoted element in PostgreSQL text array %q", src)
			}
			elems = append(elems, b.String())
		} else {
			start := i
			for i < len(body) && body[i] != ',' {
				if body[i] == '{' || body[i] == '}' {
					return nil, fmt.Errorf("model: multi-dimensional arrays are not supported: %q", src)
				}
				i++
			}
			elem := body[start:i]
			if elem == "" {
				return nil, fmt.Errorf("model: empty unquoted element in PostgreSQL text array %q", src)
			}
			if elem == "NULL" {
				return nil, fmt.Errorf("model: NULL array elements cannot be represented as strings: %q", src)
			}
			elems = append(elems, elem)
		}

		if i == len(body) {
			return elems, nil
		}
		if body[i] != ',' {
			return nil, fmt.Errorf("model: unexpected %q at offset %d in PostgreSQL text array %q", body[i], i+1, src)
		}
		i++
	}
}
