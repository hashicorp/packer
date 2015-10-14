package common

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ParseVMX parses the keys and values from a VMX file and returns
// them as a Go map.
func ParseVMX(contents string) map[string]string {
	results := make(map[string]string)

	lineRe := regexp.MustCompile(`^(.+?)\s*=\s*"(.*?)"\s*$`)

	for _, line := range strings.Split(contents, "\n") {
		matches := lineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		key := strings.ToLower(matches[1])
		results[key] = matches[2]
	}

	return results
}

// EncodeVMX takes a map and turns it into valid VMX contents.
func EncodeVMX(contents map[string]string) string {
	var buf bytes.Buffer

	i := 0
	keys := make([]string, len(contents))
	for k, _ := range contents {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", k, contents[k]))
	}

	return buf.String()
}

// WriteVMX takes a path to a VMX file and contents in the form of a
// map and writes it out.
func WriteVMX(path string, data map[string]string) (err error) {
	log.Printf("Writing VMX to: %s", path)
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	buf.WriteString(EncodeVMX(data))
	if _, err = io.Copy(f, &buf); err != nil {
		return
	}

	return
}

// ReadVMX takes a path to a VMX file and reads it into a k/v mapping.
func ReadVMX(path string) (map[string]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseVMX(string(data)), nil
}
