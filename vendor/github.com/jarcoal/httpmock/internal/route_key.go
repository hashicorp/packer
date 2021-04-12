package internal

type RouteKey struct {
	Method string
	URL    string
}

var NoResponder RouteKey

func (r RouteKey) String() string {
	if r == NoResponder {
		return "NO_RESPONDER"
	}
	return r.Method + " " + r.URL
}
