package config

type Database struct {
	Name           string `mapstructure:"CORE_DB_NAME"`
	User           string `mapstructure:"CORE_DB_USER"`
	Password       string `mapstructure:"CORE_DB_PASSWORD"`
	Port           string `mapstructure:"CORE_DB_PORT"`
	Host           string `mapstructure:"CORE_DB_HOST"`
	MaxConnections int64  `mapstructure:"CORE_DB_MAX_CONN"`
}

type HTTPServer struct {
	ServerHost               string `mapstructure:"CORE_SERVER_HOST"`
	ServerPort               string `mapstructure:"CORE_SERVER_PORT"`
	ServerAllowedOriginLocal string `mapstructure:"CORE_ALLOWED_ORIGIN_LOCAL"`
	ServerAllowedOriginProd  string `mapstructure:"CORE_ALLOWED_ORIGIN_PROD"`
}

type ClerkConfig struct {
	APIKey string `mapstructure:"CORE_CLERK_KEY"`
}

type GoogleConfig struct {
	ClientID     string `mapstructure:"CORE_GOOGLE_CLIENT_ID"`
	ClientSecret string `mapstructure:"CORE_GOOGLE_CLIENT_SECRET"`
}

type Cloudinary struct {
	APIKey string `mapstructure:"CORE_CLOUDINARY_KEY"`
}

type Sentry struct {
	DSN        string  `mapstructure:"CORE_SENTRY_DSN"`
	SampleRate float64 `mapstructure:"CORE_SENTRY_SAMPLE_RATE"`
}

type Config struct {
	Env        string     `mapstructure:"CORE_ENV"`
	DB         Database   `mapstructure:",squash"`
	HTTPServer HTTPServer `mapstructure:",squash"`

	// Clerk (auth)
	Clerk ClerkConfig `mapstructure:",squash"`

	// Google
	Google GoogleConfig `mapstructure:",squash"`

	// Cloudinary
	Cloudinary Cloudinary `mapstructure:",squash"`

	// Sentry monitoring
	Sentry Sentry `mapstructure:",squash"`
}
