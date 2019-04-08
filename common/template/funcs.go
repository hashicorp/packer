package template

import (
	"log"
	"sync"
)

// DeprecatedTemplateFunc wraps a template func to warn users that it's
// deprecated. The deprecation warning is called only once.
func DeprecatedTemplateFunc(funcName, useInstead string, deprecated func(string) string) func(string) string {
	once := sync.Once{}
	return func(in string) string {
		once.Do(func() {
			log.Printf("[WARN]: the `%s` template func is deprecated, please use %s instead",
				funcName, useInstead)
		})
		return deprecated(in)
	}
}
