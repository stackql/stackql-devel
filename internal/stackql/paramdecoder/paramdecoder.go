// Package paramdecoder decodes parameter values from their wire format
// (text or binary) into string representations suitable for SQL substitution.
package paramdecoder

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/lib/pq/oid"
)

// Decoder decodes raw parameter bytes according to their format codes
// and OIDs, returning string representations for each.
type Decoder interface {
	DecodeParams(paramOIDs []uint32, paramFormats []int16, paramValues [][]byte) ([]string, error)
}

// NewDecoder creates a new parameter decoder.
func NewDecoder() Decoder {
	return &standardDecoder{}
}

type standardDecoder struct{}

func (d *standardDecoder) DecodeParams(
	paramOIDs []uint32, paramFormats []int16, paramValues [][]byte,
) ([]string, error) {
	result := make([]string, len(paramValues))
	for i, val := range paramValues {
		if val == nil {
			result[i] = "NULL"
			continue
		}
		format := resolveFormat(paramFormats, i)
		paramOID := oid.Oid(0)
		if i < len(paramOIDs) {
			paramOID = oid.Oid(paramOIDs[i])
		}
		decoded, err := decodeParam(paramOID, format, val)
		if err != nil {
			return nil, fmt.Errorf("parameter $%d: %w", i+1, err)
		}
		result[i] = decoded
	}
	return result, nil
}

// resolveFormat returns the format code for parameter at index i.
// Per postgres protocol: empty = all text, length 1 = applies to all,
// otherwise per-parameter.
func resolveFormat(formats []int16, i int) int16 {
	if len(formats) == 0 {
		return 0 // text
	}
	if len(formats) == 1 {
		return formats[0]
	}
	if i < len(formats) {
		return formats[i]
	}
	return 0 // text
}

// decodeParam decodes a single parameter value.
// Format 0 = text (bytes are UTF-8), format 1 = binary (OID-specific encoding).
func decodeParam(paramOID oid.Oid, format int16, val []byte) (string, error) {
	if format == 0 {
		// Text format: raw bytes are the UTF-8 string representation.
		return string(val), nil
	}
	// Binary format: decode based on OID.
	return decodeBinary(paramOID, val)
}

// decodeBinary decodes a binary-encoded parameter value to its string representation.
//
//nolint:cyclop // switch over OIDs is inherently branchy
func decodeBinary(paramOID oid.Oid, val []byte) (string, error) {
	switch paramOID {
	case oid.T_bool:
		if len(val) != 1 {
			return "", fmt.Errorf("bool: expected 1 byte, got %d", len(val))
		}
		if val[0] != 0 {
			return "true", nil
		}
		return "false", nil
	case oid.T_int2:
		if len(val) != 2 {
			return "", fmt.Errorf("int2: expected 2 bytes, got %d", len(val))
		}
		return strconv.FormatInt(int64(int16(binary.BigEndian.Uint16(val))), 10), nil //nolint:gosec // deliberate int16 conversion
	case oid.T_int4:
		if len(val) != 4 {
			return "", fmt.Errorf("int4: expected 4 bytes, got %d", len(val))
		}
		return strconv.FormatInt(int64(int32(binary.BigEndian.Uint32(val))), 10), nil //nolint:gosec // deliberate int32 conversion
	case oid.T_int8:
		if len(val) != 8 {
			return "", fmt.Errorf("int8: expected 8 bytes, got %d", len(val))
		}
		return strconv.FormatInt(int64(binary.BigEndian.Uint64(val)), 10), nil
	case oid.T_float4:
		if len(val) != 4 {
			return "", fmt.Errorf("float4: expected 4 bytes, got %d", len(val))
		}
		bits := binary.BigEndian.Uint32(val)
		return strconv.FormatFloat(float64(math.Float32frombits(bits)), 'f', -1, 32), nil
	case oid.T_float8:
		if len(val) != 8 {
			return "", fmt.Errorf("float8: expected 8 bytes, got %d", len(val))
		}
		bits := binary.BigEndian.Uint64(val)
		return strconv.FormatFloat(math.Float64frombits(bits), 'f', -1, 64), nil
	case oid.T_timestamp, oid.T_timestamptz:
		if len(val) != 8 {
			return "", fmt.Errorf("timestamp: expected 8 bytes, got %d", len(val))
		}
		// Postgres binary timestamp: microseconds since 2000-01-01 00:00:00 UTC.
		microseconds := int64(binary.BigEndian.Uint64(val))
		pgEpoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		ts := pgEpoch.Add(time.Duration(microseconds) * time.Microsecond)
		return ts.Format("2006-01-02 15:04:05.999999"), nil
	case oid.T_text, oid.T_varchar, oid.T_name:
		return string(val), nil
	default:
		// Unknown OID: treat as text (safe fallback).
		return string(val), nil
	}
}
