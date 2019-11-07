package hcl2template

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func translateBuilder(path string) (string, error) {

	type ConfigV1 map[string]json.RawMessage

	type ConfigV1V2 struct {
		Artifact map[string]map[string]json.RawMessage `json:"artifact"`
	}

	type Type struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	type PostProcessor struct {
		Type   string   `json:"type"`
		Except []string `json:"except"`
		Only   []string `json:"only"`
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	c1 := ConfigV1{}
	if err := json.Unmarshal(b, &c1); err != nil {
		return "", err
	}
	c12 := ConfigV1V2{}
	if err := json.Unmarshal(b, &c12); err != nil {
		return "", err
	}

	rawBuilder, found := c1["builders"]
	if !found {
		// no v1 builders
		return path, nil
	}

	var tn []Type
	if err := json.Unmarshal([]byte(rawBuilder), &tn); err != nil {
		return "", err
	}
	var rawbuilders []json.RawMessage
	if err := json.Unmarshal([]byte(rawBuilder), &rawbuilders); err != nil {
		return "", err
	}

	var typePPs []PostProcessor
	var rawPPs []json.RawMessage
	if rawPP := c1["post-processors"]; len(rawPP) != 0 {
		if err := json.Unmarshal([]byte(rawPP), &typePPs); err != nil {
			return "", err
		}
		if err := json.Unmarshal([]byte(rawPP), &rawPPs); err != nil {
			return "", err
		}
	}

	for n, tn := range tn {
		builderName := tn.Type
		if tn.Name != "" {
			builderName = tn.Name
		}

		if c12.Artifact[tn.Type] == nil {
			c12.Artifact[tn.Type] = map[string]json.RawMessage{}
		}

		name := tn.Name
		if name == "" {
			name = fmt.Sprintf("autotranslated-builder-%d", len(c12.Artifact[tn.Type]))
		}
		if _, exists := c12.Artifact[tn.Type][name]; exists {
			return "", fmt.Errorf("%s-%s is defined in old and new config", tn.Type, name)
		}
		rawbuilder := rawbuilders[n]
		rawbuilder = removeKey(rawbuilder, "name", "only", "type")
		c12.Artifact[tn.Type][name] = rawbuilder

		for n, pp := range typePPs {
			skip := false
			for _, except := range pp.Except {
				if except == builderName {
					skip = true
					break
				}
			}
			for _, only := range pp.Only {
				if only != builderName {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			if c12.Artifact[pp.Type] == nil {
				c12.Artifact[pp.Type] = map[string]json.RawMessage{}
			}
			name := fmt.Sprintf("autotranslated-post-processor-%d", len(c12.Artifact[pp.Type]))
			if _, exists := c12.Artifact[tn.Type][name]; exists {
				return "", fmt.Errorf("%s-%s is defined in old and new config", tn.Type, name)
			}
			rawpp := rawPPs[n]
			rawpp = rawpp[:len(rawpp)-1]
			rawpp = append(rawpp, json.RawMessage(`,"source":"$artifacts.`+tn.Type+`.`+builderName+`"}`)...)
			rawpp = removeKey(rawpp, "name", "only", "type")
			c12.Artifact[pp.Type][name] = rawpp

			log.Printf("%s", rawpp)
		}

	}

	path = strings.TrimSuffix(path, ".json")
	path = strings.TrimSuffix(path, ".pk")
	path = path + ".v2.pk.json"

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	return path, enc.Encode(c12)
}

func removeKey(in json.RawMessage, keys ...string) json.RawMessage {
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(in, &m); err != nil {
		panic(err)
	}

	for _, key := range keys {
		delete(m, key)
	}

	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}
