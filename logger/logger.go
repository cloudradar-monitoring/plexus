package logger

type Logger interface {
	Errorf(f string, args ...interface{})
	Infof(f string, args ...interface{})
	Debugf(f string, args ...interface{})
}
