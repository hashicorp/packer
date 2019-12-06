package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// PrintMap prints a map
func PrintMap(m interface{}) string {
	var outputText string

	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		outputText = fmt.Sprintf("error while printing map\nreceived a %T\n  %s\n", m, m)
	} else {

		keys := v.MapKeys()
		var xs []string

		for _, k := range keys {
			xs = append(xs, fmt.Sprintf("[%s]=%s", k, v.MapIndex(k)))
		}

		sort.Strings(xs)
		if len(xs) == 0 {
			outputText = "len(map) == 0"
		} else {
			outputText = strings.Join(xs, "\n")
		}
	}

	return outputText
}

// PrintArray prints an array
func PrintArray(m interface{}) string {
	var outputText string

	switch reflect.TypeOf(m).Kind() {
	case reflect.Slice:
		var xs []string

		slice := reflect.ValueOf(m)
		for i := 0; i < slice.Len(); i++ {
			xs = append(xs, fmt.Sprintf("[%d]=%s", i, slice.Index(i)))
		}

		if len(xs) == 0 {
			outputText = "len(array) == 0"
		} else {
			outputText = strings.Join(xs, "\n")
		}
	default:
		outputText = fmt.Sprintf("error while printing array\nreceived a %T\n  %s\n", m, m)
	}

	return outputText
}

// MapToString maps a collection to a string collection
func MapToString(collection interface{}, transform func(x interface{}) string) []string {
	switch reflect.TypeOf(collection).Kind() {
	case reflect.Slice:
		var xs []string

		slice := reflect.ValueOf(collection)
		for i := 0; i < slice.Len(); i++ {
			xs = append(xs, transform(slice.Index(i).Interface()))
		}

		return xs
	default:
		panic(fmt.Sprintf("error while mapping array to string\nreceived a %T\n  %s\n", collection, collection))
	}
}
