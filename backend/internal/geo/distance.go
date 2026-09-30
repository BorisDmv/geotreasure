package geo

import "math"

const earthRadiusMeters = 6371000.0

// HaversineMeters returns the great-circle distance in meters between two
// latitude/longitude points. This is the correct formula for distances on a
// sphere and avoids the longitude-scaling error of naive planar math.
func HaversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	dPhi := (lat2 - lat1) * math.Pi / 180
	dLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// BoundingBox returns the min/max latitude and longitude that enclose a circle
// of the given radius (meters) around a point. It is used to pre-filter rows
// with an index-friendly BETWEEN clause before computing the exact distance.
func BoundingBox(lat, lng, radiusMeters float64) (minLat, maxLat, minLng, maxLng float64) {
	latDelta := radiusMeters / 111320.0 // meters per degree of latitude (roughly constant)

	// Meters per degree of longitude shrinks with latitude.
	cosLat := math.Cos(lat * math.Pi / 180)
	if cosLat < 1e-6 {
		cosLat = 1e-6
	}
	lngDelta := radiusMeters / (111320.0 * cosLat)

	return lat - latDelta, lat + latDelta, lng - lngDelta, lng + lngDelta
}
