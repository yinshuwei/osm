package osm

import "log"

type Logger interface {
	Log(msg string, data map[string]string)
}

type DefaultLogger struct {
}

func (*DefaultLogger) Log(msg string, data map[string]string) {
	log.Println(msg, data)
}
