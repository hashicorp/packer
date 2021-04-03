package proxmox

import (
	"regexp"
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
	implicitFirstKey string,
) QemuDevice {
	var confMap = QemuDevice{}
	confList := strings.Split(kvString, confSeparator)

	if implicitFirstKey != "" {
		if !strings.Contains(confList[0], "=") {
			confMap[implicitFirstKey] = confList[0]
			confList = confList[1:]
		}
	}

	for _, item := range confList {
		key, value := ParseSubConf(item, subConfSeparator)
		confMap[key] = value
	}
	return confMap
}

func ParsePMConf(
	kvString string,
	implicitFirstKey string,
) QemuDevice {
	return ParseConf(kvString, ",", "=", implicitFirstKey)
}

// Convert a disk-size string to a GB float
func DiskSizeGB(dcSize interface{}) float64 {
	var diskSize float64
	switch dcSize.(type) {
	case string:
		diskString := strings.ToUpper(dcSize.(string))
		re := regexp.MustCompile("([0-9]+)([A-Z]*)")
		diskArray := re.FindStringSubmatch(diskString)

		diskSize, _ = strconv.ParseFloat(diskArray[1], 64)

		if len(diskArray) >= 3 {
			switch diskArray[2] {
			case "T", "TB":
				diskSize *= 1024
			case "G", "GB":
				//Nothing to do
			case "M", "MB":
				diskSize /= 1024
			case "K", "KB":
				diskSize /= 1048576
			}
		}
	case float64:
		diskSize = dcSize.(float64)
	}
	return diskSize
}
