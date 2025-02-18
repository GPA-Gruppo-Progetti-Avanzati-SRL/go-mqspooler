package mqspooler

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-mqspooler/fallback"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-mqspooler/mq"
)

type Config struct {
	Fallback fallback.Config `yaml:"fallback" mapstructure:"fallback"`
	Mq       mq.Config       `yaml:"mq" mapstructure:"mq"`
}
