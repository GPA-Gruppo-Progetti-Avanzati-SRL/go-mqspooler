package fallback

type Config struct {
	ExitOnFailure bool `yaml:"exit-on-failure" mapstructure:"exit-on-failure"`
}
