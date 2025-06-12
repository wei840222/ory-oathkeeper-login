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
	KeyGinPort = "gin.port"
	KeyGinHost = "gin.host"

	KeyCacheRedisHost     = "cache.redis.host"
	KeyCacheRedisPort     = "cache.redis.port"
	KeyCacheRedisDB       = "cache.redis.db"
	KeyCacheRedisPassword = "cache.redis.password"

	KeyCacheTTL = "cache.ttl"

	KeyArgoCDServerURL = "argo-cd.server-url"
	KeyArgoCDUsername  = "argo-cd.username"
	KeyArgoCDPassword  = "argo-cd.password"

	KeyGhostServerURL = "ghost.server-url"
	KeyGhostOriginURL = "ghost.origin-url"
	KeyGhostUsername  = "ghost.username"
	KeyGhostPassword  = "ghost.password"

	KeyN8NServerURL = "n8n.server-url"
	KeyN8NUsername  = "n8n.username"
	KeyN8NPassword  = "n8n.password"
)
