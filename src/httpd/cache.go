package httpd

const (
	Multiplayer = HTTPCacheID(iota)
	LocalMultiplayer
	AdminSessions
	Throttle
)

type HTTPCacheID int

type HTTPCache map[HTTPCacheID]interface{}
