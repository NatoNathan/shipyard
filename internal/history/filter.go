package history

// FilterByPackage filters history entries by package name
// Returns all entries if packageName is empty
func FilterByPackage(entries []Entry, packageName string) []Entry {
	if packageName == "" {
		return entries
	}

	var filtered []Entry
	for _, entry := range entries {
		if entry.Package == packageName {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterByVersion filters history entries by version
// Returns all entries if version is empty
func FilterByVersion(entries []Entry, version string) []Entry {
	if version == "" {
		return entries
	}

	var filtered []Entry
	for _, entry := range entries {
		if entry.Version == version {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterConsignmentsByMetadata filters consignments within entries by metadata
// Returns entries with only matching consignments; entries may have empty consignments arrays
// metadataKey: e.g., "environment", "team" (must be type="string" or type="enum")
// metadataValue: e.g., "production", "backend"
func FilterConsignmentsByMetadata(entries []Entry, metadataKey, metadataValue string) []Entry {
	filtered := make([]Entry, len(entries))
	for i, entry := range entries {
		// Copy entry metadata
		filtered[i] = Entry{
			Version:      entry.Version,
			Package:      entry.Package,
			Tag:          entry.Tag,
			Timestamp:    entry.Timestamp,
			Consignments: []Consignment{},
		}

		// Filter consignments by metadata
		for _, c := range entry.Consignments {
			if matchesMetadata(c.Metadata, metadataKey, metadataValue) {
				filtered[i].Consignments = append(filtered[i].Consignments, c)
			}
		}
	}
	return filtered
}

// SortByTimestamp sorts entries by timestamp
func SortByTimestamp(entries []Entry, descending bool) []Entry {
	// Create a copy to avoid modifying original
	sorted := make([]Entry, len(entries))
	copy(sorted, entries)

	// Sort by timestamp
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			var shouldSwap bool
			if descending {
				shouldSwap = sorted[i].Timestamp.Before(sorted[j].Timestamp)
			} else {
				shouldSwap = sorted[i].Timestamp.After(sorted[j].Timestamp)
			}
			if shouldSwap {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// matchesMetadata checks if a consignment's metadata matches the given key-value pair
func matchesMetadata(metadata map[string]interface{}, key, value string) bool {
	if metadata == nil {
		return false
	}

	metaValue, exists := metadata[key]
	if !exists {
		return false
	}

	// Convert metadata value to string for comparison
	var metaStr string
	switch v := metaValue.(type) {
	case string:
		metaStr = v
	default:
		// For non-string types, we don't match
		return false
	}

	return metaStr == value
}
