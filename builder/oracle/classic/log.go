package classic

import "log"

type Logger struct {
	Enabled bool
}

func (l *Logger) Log(input ...interface{}) {
	if !l.Enabled {
		return
	}
	log.Println(input...)
}
