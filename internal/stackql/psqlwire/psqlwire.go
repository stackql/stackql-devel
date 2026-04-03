package psqlwire

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"

	postgreswire "github.com/stackql/psql-wire"

	"github.com/jackc/pgtype"
	"github.com/stackql/psql-wire/pkg/sqldata"
)

func processRowElement(rowElement interface{}) interface{} {
	switch re := rowElement.(type) {
	case driver.Valuer:
		v, _ := re.Value()
		return v
	default:
		return re
	}
}

// TODO: remove this hack once correct type system comes in
func shimNumericElement(rowElement interface{}) interface{} {
	switch re := rowElement.(type) { //nolint:gocritic // acceptable
	case []byte:
		f, err := strconv.ParseFloat(string(re), 64)
		if err == nil {
			return f
		}
		if string(re) == "null" {
			return nil
		}
	}
	return rowElement
}

// TODO: remove this hack once correct type system comes in
// Acknowledgement: This is from the MIT-licensed:
//
//	https://github.com/jackc/pgx/blob/9ae852eb583d2dced83b1d2ffe1c8803dda2c92e/pgtype/numeric.go#L256
//
//nolint:gocritic // this is a hack
func shimNumericTextBytes(n *pgtype.Numeric) []byte {
	intStr := n.Int.String()

	if intStr == "<nil>" {
		return []byte("null")
	}

	buf := &bytes.Buffer{}

	if len(intStr) > 0 && intStr[:1] == "-" {
		intStr = intStr[1:]
		buf.WriteByte('-')
	}

	exp := int(n.Exp)
	if exp > 0 {
		buf.WriteString(intStr)
		for i := 0; i < exp; i++ {
			buf.WriteByte('0')
		}
	} else if exp < 0 {
		if len(intStr) <= -exp {
			buf.WriteString("0.")
			leadingZeros := -exp - len(intStr)
			for i := 0; i < leadingZeros; i++ {
				buf.WriteByte('0')
			}
			buf.WriteString(intStr)
		} else if len(intStr) > -exp {
			dpPos := len(intStr) + exp
			buf.WriteString(intStr[:dpPos])
			buf.WriteByte('.')
			buf.WriteString(intStr[dpPos:])
		}
	} else {
		buf.WriteString(intStr)
	}

	return buf.Bytes()
}

// end hack

func ExtractRowElement(column sqldata.ISQLColumn, src interface{}, ci *pgtype.ConnInfo) ([]byte, error) {
	typed, has := ci.DataTypeForOID(column.GetObjectID())
	if !has {
		return nil, fmt.Errorf("unknown data type: %T", column)
	}

	processedElement := processRowElement(src)
	// TODO: retire this hack once correct type system comes in
	if typed.Name == "numeric" {
		processedElement = shimNumericElement(src)
	}
	// end hack
	// NOTE: coerceForOID is available for binary format encoding (Phase 4).
	// For text format (current default), string values pass through to
	// pgtype's text encoder which handles all types correctly.
	// Coercing to native Go types here would change the text encoding
	// format (e.g. bool: "t"→"true", int: quoted→unquoted).

	err := typed.Value.Set(processedElement)
	if err != nil {
		return nil, err
	}

	fc, err := getFormatCode(column.GetFormat())
	if err != nil {
		return nil, err
	}
	// TODO: retire this hack once correct type system comes in
	switch t := typed.Value.(type) { //nolint:gocritic // acceptable
	case *pgtype.Numeric:
		b := shimNumericTextBytes(t)
		return b, nil
	}
	// end hack
	encoder := fc.Encoder(typed)
	bb, err := encoder(ci, nil)
	if err != nil {
		return nil, err
	}
	return bb, nil
}

// coerceForOID converts a value (often a string or []byte from the RDBMS)
// into the Go type that pgtype expects for the given type name.
// This bridges the gap between sqlite's string-heavy output and pgtype's
// type-specific Set() requirements.
//
//nolint:gocyclo,cyclop // switch over type names is inherently branchy
func coerceForOID(typeName string, val interface{}) interface{} {
	if val == nil {
		return nil
	}
	s, isStr := valToString(val)
	if !isStr {
		return val // already a native Go type; pass through
	}
	if s == "null" || s == "NULL" || s == "<nil>" {
		return nil
	}
	switch typeName {
	case "numeric":
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
		return val
	case "int2":
		i, err := strconv.ParseInt(s, 10, 16)
		if err == nil {
			return int16(i)
		}
		return val
	case "int4":
		i, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			return int32(i)
		}
		return val
	case "int8":
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return i
		}
		return val
	case "float4":
		f, err := strconv.ParseFloat(s, 32)
		if err == nil {
			return float32(f)
		}
		return val
	case "float8":
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f
		}
		return val
	case "bool":
		switch s {
		case "true", "TRUE", "t", "1", "yes":
			return true
		case "false", "FALSE", "f", "0", "no":
			return false
		}
		return val
	case "json", "jsonb":
		return s // pass as string; pgtype text encoder handles it
	default:
		return val // text, varchar, timestamp, etc. — string is fine
	}
}

// valToString extracts a string from string or []byte values.
func valToString(val interface{}) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return "", false
	}
}

func getFormatCode(fc string) (postgreswire.FormatCode, error) {
	switch fc {
	case "TextFormat":
		return postgreswire.TextFormat, nil
	case "":
		return postgreswire.BinaryFormat, nil
	default:
		return -1, fmt.Errorf("cannot find format code for '%s'", fc)
	}
}
