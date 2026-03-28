// Package config handles loading bot configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	// Discord bot token.
	Token string

	// Guild ID for registering commands. Empty string registers globally.
	GuildID string
}

func Load() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	return &Config{
		Token:   token,
		GuildID: os.Getenv("DISCORD_GUILD_ID"),
	}, nil
}
