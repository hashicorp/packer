package testing

import (
	"fmt"
	"math/rand"
	"time"
	"encoding/json"
)

func NewVMName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func RenderConfig(config map[string]interface{}) string {
	t := map[string][]map[string]interface{}{
		"builders": {
			map[string]interface{}{
				"type": "test",
			},
		},
	}
	for k, v := range config {
		t["builders"][0][k] = v
	}

	j, _ := json.Marshal(t)
	return string(j)
}
