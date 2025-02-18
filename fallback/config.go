package fallback

type Config struct {
	Type          string `yaml:"type" mapstructure:"type"`
	ExitOnFailure bool   `yaml:"exit-on-failure" mapstructure:"exit-on-failure"`
}
