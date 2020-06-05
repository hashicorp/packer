package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// FormatResult formats the given result value for human-readable output.
//
// The value must currently be a string, list, map, and any nested values
// with those same types.
func FormatResult(value interface{}) string {
	return formatResult(value, false)
}

func formatResult(value interface{}, nested bool) string {
	if value == nil {
		return "null"
	}
	switch output := value.(type) {
	case string:
		if nested {
			return fmt.Sprintf("%q", output)
		}
		return output
	case int:
		return strconv.Itoa(output)
	case float64:
		return fmt.Sprintf("%g", output)
	case bool:
		switch {
		case output == true:
			return "true"
		default:
			return "false"
		}
	case []interface{}:
		return formatListResult(output)
	case map[string]interface{}:
		return formatMapResult(output)
	default:
		return "<unknown-type>"
	}
}

func formatListResult(value []interface{}) string {
	var outputBuf bytes.Buffer
	outputBuf.WriteString("[")
	if len(value) > 0 {
		outputBuf.WriteString("\n")
	}

	for _, v := range value {
		raw := formatResult(v, true)
		outputBuf.WriteString(indent(raw))
		outputBuf.WriteString(",\n")
	}

	outputBuf.WriteString("]")
	return outputBuf.String()
}

func formatMapResult(value map[string]interface{}) string {
	ks := make([]string, 0, len(value))
	for k := range value {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	var outputBuf bytes.Buffer
	outputBuf.WriteString("{")
	if len(value) > 0 {
		outputBuf.WriteString("\n")
	}

	for _, k := range ks {
		v := value[k]
		rawK := formatResult(k, true)
		rawV := formatResult(v, true)

		outputBuf.WriteString(indent(fmt.Sprintf("%s = %s", rawK, rawV)))
		outputBuf.WriteString("\n")
	}

	outputBuf.WriteString("}")
	return outputBuf.String()
}

func indent(value string) string {
	var outputBuf bytes.Buffer
	s := bufio.NewScanner(strings.NewReader(value))
	newline := false
	for s.Scan() {
		if newline {
			outputBuf.WriteByte('\n')
		}
		outputBuf.WriteString("  " + s.Text())
		newline = true
	}

	return outputBuf.String()
}
