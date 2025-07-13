package config

const (
	AppName  = "ory-oathkeeper-login"
	FileName = "config"

	KeyLogLevel  = "log.level"
	KeyLogFormat = "log.format"
	KeyLogColor  = "log.color"

	KeyO11yHost = "o11y.host"
	KeyO11yPort = "o11y.port"

	KeyGinMode = "gin.mode"

	KeyHTTPPort = "http.port"
	KeyHTTPHost = "http.host"

	KeyCacheRedisHost     = "cache.redis.host"
	KeyCacheRedisPort     = "cache.redis.port"
	KeyCacheRedisDB       = "cache.redis.db"
	KeyCacheRedisPassword = "cache.redis.password"

	KeyCacheTTL = "cache.ttl"

	KeyArgoCDServerURL = "argo_cd.server_url"
	KeyArgoCDUsername  = "argo_cd.username"
	KeyArgoCDPassword  = "argo_cd.password"

	KeyGhostServerURL = "ghost.server_url"
	KeyGhostOriginURL = "ghost.origin_url"
	KeyGhostUsername  = "ghost.username"
	KeyGhostPassword  = "ghost.password"

	KeyN8NServerURL = "n8n.server_url"
	KeyN8NUsername  = "n8n.username"
	KeyN8NPassword  = "n8n.password"

	KeyNocoDBServerURL = "nocodb.server_url"
	KeyNocoDBUsername  = "nocodb.username"
	KeyNocoDBPassword  = "nocodb.password"
)
