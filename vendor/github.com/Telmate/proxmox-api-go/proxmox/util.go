package proxmox

import (
	"strconv"
	"strings"
)

func inArray(arr []string, str string) bool {
	for _, elem := range arr {
		if elem == str {
			return true
		}
	}

	return false
}

func Itob(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

// ParseSubConf - Parse standard sub-conf strings `key=value`.
func ParseSubConf(
	element string,
	separator string,
) (key string, value interface{}) {
	if strings.Contains(element, separator) {
		conf := strings.Split(element, separator)
		key, value := conf[0], conf[1]
		var interValue interface{}

		// Make sure to add value in right type,
		// because all subconfig are returned as strings from Proxmox API.
		if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			interValue = int(iValue)
		} else if bValue, err := strconv.ParseBool(value); err == nil {
			interValue = bValue
		} else {
			interValue = value
		}
		return key, interValue
	}
	return
}

// ParseConf - Parse standard device conf string `key1=val1,key2=val2`.
func ParseConf(
	kvString string,
	confSeparator string,
	subConfSeparator string,
) QemuDevice {
	var confMap = QemuDevice{}
	confList := strings.Split(kvString, confSeparator)
	for _, item := range confList {
		key, value := ParseSubConf(item, subConfSeparator)
		confMap[key] = value
	}
	return confMap
}
