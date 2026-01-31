package metadata

import (
	"fmt"
	"regexp"

	"github.com/NatoNathan/shipyard/internal/config"
)

// ValidateMetadata validates metadata against the configuration
func ValidateMetadata(cfg *config.Config, metadata map[string]string) error {
	if cfg.Metadata.Fields == nil || len(cfg.Metadata.Fields) == 0 {
		return nil
	}

	fieldDefs := make(map[string]config.MetadataField)
	for _, field := range cfg.Metadata.Fields {
		fieldDefs[field.Name] = field
	}

	parsedMetadata := make(map[string]interface{})

	for key, rawValue := range metadata {
		field, exists := fieldDefs[key]
		if !exists {
			parsedMetadata[key] = rawValue
			continue
		}

		// Parse value
		parsedValue, err := ParseMetadataValue(field, rawValue)
		if err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}

		// Validate
		if err := validateField(field, parsedValue); err != nil {
			return fmt.Errorf("validation failed for %s: %w", key, err)
		}

		parsedMetadata[key] = parsedValue
	}

	// Check required fields
	for _, field := range cfg.Metadata.Fields {
		if field.Required {
			if _, ok := parsedMetadata[field.Name]; !ok {
				if field.Default != "" {
					parsed, err := ParseMetadataValue(field, field.Default)
					if err != nil {
						return fmt.Errorf("invalid default for %s: %w", field.Name, err)
					}
					parsedMetadata[field.Name] = parsed
				} else {
					return fmt.Errorf("required field missing: %s", field.Name)
				}
			}
		}
	}

	return nil
}

func validateField(field config.MetadataField, value interface{}) error {
	if field.Required && isEmpty(value) {
		return fmt.Errorf("cannot be empty")
	}

	switch field.Type {
	case "", "string":
		return validateString(field, value)
	case "int", "integer":
		return validateInt(field, value)
	case "list", "array":
		return validateList(field, value)
	case "map", "hashmap", "object":
		return validateMap(field, value)
	}
	return nil
}

func validateString(field config.MetadataField, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	// Pattern validation
	if field.Pattern != "" {
		matched, err := regexp.MatchString(field.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("%q does not match pattern %s", str, field.Pattern)
		}
	}

	// Length validation
	if field.MinLength != nil && len(str) < *field.MinLength {
		return fmt.Errorf("too short (min: %d)", *field.MinLength)
	}
	if field.MaxLength != nil && len(str) > *field.MaxLength {
		return fmt.Errorf("too long (max: %d)", *field.MaxLength)
	}

	// Enum validation
	if len(field.AllowedValues) > 0 {
		for _, allowed := range field.AllowedValues {
			if str == allowed {
				return nil
			}
		}
		return fmt.Errorf("%q not in allowed values: %v", str, field.AllowedValues)
	}

	return nil
}

func validateInt(field config.MetadataField, value interface{}) error {
	var intVal int
	switch v := value.(type) {
	case int:
		intVal = v
	case float64:
		intVal = int(v)
	default:
		return fmt.Errorf("expected int, got %T", value)
	}

	if field.Min != nil && intVal < *field.Min {
		return fmt.Errorf("%d below minimum %d", intVal, *field.Min)
	}
	if field.Max != nil && intVal > *field.Max {
		return fmt.Errorf("%d above maximum %d", intVal, *field.Max)
	}

	return nil
}

func validateList(field config.MetadataField, value interface{}) error {
	list, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected list, got %T", value)
	}

	if field.MinItems != nil && len(list) < *field.MinItems {
		return fmt.Errorf("too few items (min: %d)", *field.MinItems)
	}
	if field.MaxItems != nil && len(list) > *field.MaxItems {
		return fmt.Errorf("too many items (max: %d)", *field.MaxItems)
	}

	return nil
}

func validateMap(field config.MetadataField, value interface{}) error {
	_, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map, got %T", value)
	}
	return nil
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}
