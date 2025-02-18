package mq

import (
	"time"
)

type Config struct {
	Hostname              string        `mapstructure:"hostname"`
	Port                  int           `mapstructure:"port"`
	Mode                  string        `mapstructure:"mode"`
	QueueManager          string        `mapstructure:"queueManager"`
	Channel               string        `mapstructure:"channel"`
	QueueName             string        `mapstructure:"queuename"`
	DecodeBody            bool          `mapstructure:"decodeBody"`
	UserName              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	SleepNoMessage        time.Duration `mapstructure:"sleepNoMessage"`
	TLSCipherSpec         string        `mapstructure:"tlsCipherSpec"`
	TLSClientAuth         string        `mapstructure:"tlsClientAuth"`
	KeyRepository         string        `mapstructure:"keyRepository"`
	CertificateLabel      string        `mapstructure:"certificateLabel"`
	MaxTransactionMessage int           `mapstructure:"maxTransactionMessage"`
	Length                int           `mapstructure:"messagelength"`
}
