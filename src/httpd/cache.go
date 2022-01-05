package httpd

const (
	cacheMultiplayer = HTTPCacheID(iota)
	cacheAdminSessions
	cacheThrottle
)

type HTTPCacheID int

type HTTPCache map[HTTPCacheID]interface{}
