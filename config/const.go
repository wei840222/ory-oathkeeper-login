package config

const (
	AppName        = "login-server"
	ConfigFileName = "config"

	ConfigKeyLogLevel  = "log.level"
	ConfigKeyLogFormat = "log.format"
	ConfigKeyLogColor  = "log.color"

	ConfigKeyO11yHost = "o11y.host"
	ConfigKeyO11yPort = "o11y.port"

	ConfigKeyGinMode = "gin.mode"
	ConfigKeyGinPort = "gin.port"
	ConfigKeyGinHost = "gin.host"

	ConfigKeyCacheRedisHost     = "cache.redis.host"
	ConfigKeyCacheRedisPort     = "cache.redis.port"
	ConfigKeyCacheRedisDB       = "cache.redis.db"
	ConfigKeyCacheRedisPassword = "cache.redis.password"

	ConfigKeyCacheTTL = "cache.ttl"

	ConfigKeyArgoCDServerURL = "argo-cd.server-url"
	ConfigKeyArgoCDUsername  = "argo-cd.username"
	ConfigKeyArgoCDPassword  = "argo-cd.password"

	ConfigKeyGhostServerURL = "ghost.server-url"
	ConfigKeyGhostOrigin    = "ghost.origin-url"
	ConfigKeyGhostUsername  = "ghost.username"
	ConfigKeyGhostPassword  = "ghost.password"

	ConfigKeyN8NServerURL = "n8n.server-url"
	ConfigKeyN8NUsername  = "n8n.username"
	ConfigKeyN8NPassword  = "n8n.password"
)
