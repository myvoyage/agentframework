// Package drivers provides specific hardware driver implementations.
package drivers

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GetUint32Param extracts a uint32 parameter from a map.
func GetUint32Param(params map[string]interface{}, key string) (uint32, error) {
	value, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing parameter: %s", key)
	}

	switch v := value.(type) {
	case float64:
		return uint32(v), nil
	case int:
		return uint32(v), nil
	case int64:
		return uint32(v), nil
	case uint32:
		return v, nil
	default:
		return 0, fmt.Errorf("invalid parameter type for %s", key)
	}
}

// GetByteArrayParam extracts a byte array parameter from a map.
func GetByteArrayParam(params map[string]interface{}, key string) ([]byte, error) {
	value, ok := params[key]
	if !ok {
		return []byte{}, nil // Empty data is valid
	}

	switch v := value.(type) {
	case []interface{}:
		result := make([]byte, len(v))
		for i, item := range v {
			switch val := item.(type) {
			case float64:
				result[i] = byte(val)
			case int:
				result[i] = byte(val)
			default:
				return nil, fmt.Errorf("invalid element type at index %d", i)
			}
		}
		return result, nil
	case string:
		// Try to parse as JSON array
		var result []byte
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			// Try as hex string
			return []byte(v), nil
		}
		return result, nil
	default:
		return nil, fmt.Errorf("invalid parameter type for %s", key)
	}
}

// GetBoolParam extracts a bool parameter from a map.
func GetBoolParam(params map[string]interface{}, key string) (bool, error) {
	value, ok := params[key]
	if !ok {
		return false, nil // Default to false
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("invalid parameter type for %s", key)
	}
}

// GetIntParam extracts an int parameter from a map.
func GetIntParam(params map[string]interface{}, key string) (int, error) {
	value, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing parameter: %s", key)
	}

	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid parameter type for %s", key)
	}
}

// GetStringParam extracts a string parameter from a map.
func GetStringParam(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing parameter: %s", key)
	}

	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid parameter type for %s", key)
	}

	return s, nil
}

// GetBoolArrayParam extracts a bool array parameter from a map.
func GetBoolArrayParam(params map[string]interface{}, key string) ([]bool, error) {
	value, ok := params[key]
	if !ok {
		return nil, errors.New("missing parameter: " + key)
	}

	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("invalid parameter type for " + key)
	}

	bools := make([]bool, len(arr))
	for i, v := range arr {
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("invalid element type at index %d", i)
		}
		bools[i] = b
	}

	return bools, nil
}
