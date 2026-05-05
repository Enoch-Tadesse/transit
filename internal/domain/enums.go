package domain

// string backed enums so they are self describing in json serialization
// and the contract examples use the exact same string the code does

type Role string

const (
	RolePassenger Role = "passenger"
	RoleDriver    Role = "driver"
	RoleAdmin     Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RolePassenger, RoleDriver, RoleAdmin:
		return true
	default:
		return false
	}
}

type TrackingStatus string

const (
	StatusActive      TrackingStatus = "active"
	StatusInactive    TrackingStatus = "inactive"
	StatusStale       TrackingStatus = "stale"
	StatusMaintenance TrackingStatus = "maintenance"
)

func (s TrackingStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusStale, StatusMaintenance:
		return true
	default:
		return false
	}
}

type Direction string

const (
	DirForward Direction = "forward"
	DirReverse Direction = "reverse"
)

func (d Direction) IsValid() bool {
	switch d {
	case DirForward, DirReverse:
		return true
	default:
		return false
	}
}

type TripStatus string

const (
	TripStatusPending      TripStatus = "pending"
	TripStatusPlanned      TripStatus = "planned"
	TripStatusNoRouteFound TripStatus = "no_route_found"
	TripStatusExpired      TripStatus = "expired"
)

func (s TripStatus) IsValid() bool {
	switch s {
	case TripStatusPending, TripStatusPlanned, TripStatusNoRouteFound, TripStatusExpired:
		return true
	default:
		return false
	}
}
