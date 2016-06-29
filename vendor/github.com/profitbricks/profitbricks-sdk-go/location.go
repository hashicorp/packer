package profitbricks

// ListLocations returns location collection data
func ListLocations() Collection {
	return is_list(location_col_path())
}

// GetLocation returns location data
func GetLocation(locid string) Instance {
	return is_get(location_path(locid))
}
