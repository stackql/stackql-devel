package provider

func getGoogleMap() map[string]interface{} {
	googleMap := map[string]interface{}{
		"name": googleProviderName,
	}
	return googleMap
}

func getGoogleMapExtended() map[string]interface{} {
	return getGoogleMap()
}
