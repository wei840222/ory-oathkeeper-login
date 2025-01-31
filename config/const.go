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

	ConfigKeyArgoCDServerURL = "argo-cd.server.url"
	ConfigKeyArgoCDUsername  = "argo-cd.username"
	ConfigKeyArgoCDPassword  = "argo-cd.password"

	ConfigKeyGhostServerURL = "ghost.server.url"
	ConfigKeyGhostUsername  = "ghost.username"
	ConfigKeyGhostPassword  = "ghost.password"

	ConfigKeyN8NServerURL = "n8n.server.url"
	ConfigKeyN8NUsername  = "n8n.username"
	ConfigKeyN8NPassword  = "n8n.password"
)
