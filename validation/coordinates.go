package validation

import (
	"fmt"
	"math"
)

// Latitude validates that value is a valid latitude (-90 to 90).
func Latitude(field string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -90 || value > 90 {
		return &ValidationError{Field: field, Message: "latitude must be between -90 and 90"}
	}
	return nil
}

// Longitude validates that value is a valid longitude (-180 to 180).
func Longitude(field string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -180 || value > 180 {
		return &ValidationError{Field: field, Message: "longitude must be between -180 and 180"}
	}
	return nil
}

// Coordinates validates a latitude and longitude pair.
func Coordinates(latField string, lat float64, lonField string, lon float64) error {
	if err := Latitude(latField, lat); err != nil {
		return err
	}
	return Longitude(lonField, lon)
}

// CoordinatesInRadius validates that a point (lat, lon) is within the given
// radius (in kilometres) of the centre point (centLat, centLon) using the
// Haversine formula.
func CoordinatesInRadius(field string, lat, lon, centLat, centLon, radiusKm float64) error {
	if err := Coordinates(field+"_lat", lat, field+"_lon", lon); err != nil {
		return err
	}
	if err := Coordinates(field+"_center_lat", centLat, field+"_center_lon", centLon); err != nil {
		return err
	}
	dist := haversine(lat, lon, centLat, centLon)
	if dist > radiusKm {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("coordinates must be within %.2f km of centre (distance: %.2f km)", radiusKm, dist),
		}
	}
	return nil
}

const earthRadiusKm = 6371.0

// haversine calculates the great-circle distance between two points on
// Earth in kilometres.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
