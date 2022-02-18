package osm

type Logger interface {
	Printf(format string, v ...interface{})
}
