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
