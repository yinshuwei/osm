package osm

import "log"

type Logger interface {
	Error(msg string, data map[string]string)
	Info(msg string, data map[string]string)
	Warn(msg string, data map[string]string)
}

type DefaultLogger struct {
}

func (*DefaultLogger) Error(msg string, data map[string]string) {
	log.Println("error", msg, data)
}

func (*DefaultLogger) Info(msg string, data map[string]string) {
	log.Println("info", msg, data)
}

func (*DefaultLogger) Warn(msg string, data map[string]string) {
	log.Println("warn", msg, data)
}
