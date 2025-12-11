package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GitHub   GitHubConfig   `mapstructure:"github"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Review   ReviewConfig   `mapstructure:"review"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

func (c ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	Path   string `mapstructure:"path"`
}

type GitHubConfig struct {
	AppID          int64  `mapstructure:"app_id"`
	InstallationID int64  `mapstructure:"installation_id"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	Token          string `mapstructure:"token"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	BaseURL        string `mapstructure:"base_url"`
}

type LLMConfig struct {
	Provider  string `mapstructure:"provider"`
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	BaseURL   string `mapstructure:"base_url"`
	Timeout   int    `mapstructure:"timeout"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type ReviewConfig struct {
	Languages      []string `mapstructure:"languages"`
	MaxDiffLines   int      `mapstructure:"max_diff_lines"`
	IgnorePatterns []string `mapstructure:"ignore_patterns"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("CODE_SENTINEL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.path", "./data/sentinel.db")

	viper.SetDefault("github.base_url", "https://api.github.com")

	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm.model", "gpt-4")
	viper.SetDefault("llm.timeout", 60)
	viper.SetDefault("llm.max_tokens", 4096)

	viper.SetDefault("review.languages", []string{"go", "java", "python"})
	viper.SetDefault("review.max_diff_lines", 500)
	viper.SetDefault("review.ignore_patterns", []string{"*.md", "*.json", "go.mod", "go.sum"})

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}
