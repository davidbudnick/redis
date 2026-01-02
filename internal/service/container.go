package service

// Container holds all service dependencies for the application.
// This enables dependency injection and makes the code more testable.
type Container struct {
	Config ConfigService
	Redis  RedisService
}

// NewContainer creates a new service container with the provided services.
func NewContainer(config ConfigService, redis RedisService) *Container {
	return &Container{
		Config: config,
		Redis:  redis,
	}
}

// Close closes all services in the container.
func (c *Container) Close() error {
	var lastErr error

	if c.Config != nil {
		if err := c.Config.Close(); err != nil {
			lastErr = err
		}
	}

	if c.Redis != nil {
		if err := c.Redis.Disconnect(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
