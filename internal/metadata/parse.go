package metadata

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
)

// ParseMetadataValue converts CLI string to typed value
func ParseMetadataValue(field config.MetadataField, rawValue string) (interface{}, error) {
	switch field.Type {
	case "", "string":
		return rawValue, nil

	case "int", "integer":
		val, err := strconv.Atoi(strings.TrimSpace(rawValue))
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", rawValue)
		}
		return val, nil

	case "list", "array":
		return parseList(rawValue, field)

	case "map", "hashmap", "object":
		return parseMap(rawValue)

	default:
		return nil, fmt.Errorf("unsupported type: %s", field.Type)
	}
}

func parseList(rawValue string, field config.MetadataField) ([]interface{}, error) {
	trimmed := strings.TrimSpace(rawValue)

	// Try JSON array first
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return nil, fmt.Errorf("invalid JSON array: %v", err)
		}
		return convertListItems(arr, field)
	}

	// Fall back to comma-separated
	if trimmed == "" {
		return []interface{}{}, nil
	}

	parts := strings.Split(trimmed, ",")
	items := make([]interface{}, 0, len(parts))
	for _, part := range parts {
		items = append(items, strings.TrimSpace(part))
	}

	return convertListItems(items, field)
}

func convertListItems(items []interface{}, field config.MetadataField) ([]interface{}, error) {
	if field.ItemType == "" || field.ItemType == "string" {
		result := make([]interface{}, len(items))
		for i, item := range items {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result, nil
	}

	if field.ItemType == "int" || field.ItemType == "integer" {
		result := make([]interface{}, len(items))
		for i, item := range items {
			switch v := item.(type) {
			case float64:
				result[i] = int(v)
			case int:
				result[i] = v
			case string:
				parsed, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					return nil, fmt.Errorf("invalid integer: %s", v)
				}
				result[i] = parsed
			default:
				return nil, fmt.Errorf("cannot convert to int: %v", item)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported itemType: %s", field.ItemType)
}

func parseMap(rawValue string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(rawValue)

	if !strings.HasPrefix(trimmed, "{") {
		return nil, fmt.Errorf("map must use JSON syntax: {...}")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	return result, nil
}
