package fat

import (
	"bytes"
	"fmt"
	"strings"
)

// checksumShortName returns the checksum for the shortname that is used
// for the long name entries.
func checksumShortName(name string) uint8 {
	var sum uint8 = name[0]
	for i := uint8(1); i < 11; i++ {
		sum = name[i] + (((sum & 1) << 7) + ((sum & 0xFE) >> 1))
	}

	return sum
}

// generateShortName takes a list of existing short names and a long
// name and generates the next valid short name. This process is done
// according to the MS specification.
func generateShortName(longName string, used []string) (string, error) {
	longName = strings.ToUpper(longName)

	// Split the string at the final "."
	dotIdx := strings.LastIndex(longName, ".")

	var ext string
	if dotIdx == -1 {
		dotIdx = len(longName)
	} else {
		ext = longName[dotIdx+1:]
	}

	ext = cleanShortString(ext)
	ext = ext[0:]
	rawName := longName[0:dotIdx]
	name := cleanShortString(rawName)
	simpleName := fmt.Sprintf("%s.%s", name, ext)
	if ext == "" {
		simpleName = simpleName[0 : len(simpleName)-1]
	}

	doSuffix := name != rawName || len(name) > 8 || len(ext) > 3
	if !doSuffix {
		for _, usedSingle := range used {
			if strings.ToUpper(usedSingle) == simpleName {
				doSuffix = true
				break
			}
		}
	}

	if doSuffix {
		if len(ext) > 3 {
			ext = ext[:3]
		}
		found := false
		for i := 1; i < 99999; i++ {
			serial := fmt.Sprintf("~%d", i)

			nameOffset := 8 - len(serial)
			if len(name) < nameOffset {
				nameOffset = len(name)
			}

			serialName := fmt.Sprintf("%s%s", name[0:nameOffset], serial)
			simpleName = fmt.Sprintf("%s.%s", serialName, ext)

			exists := false
			for _, usedSingle := range used {
				if strings.ToUpper(usedSingle) == simpleName {
					exists = true
					break
				}
			}

			if !exists {
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("could not generate short name for %s", longName)
		}
	}

	return simpleName, nil
}

// shortNameEntryValue returns the proper formatted short name value
// for the directory cluster entry.
func shortNameEntryValue(name string) string {
	var shortParts []string
	if name == "." || name == ".." {
		shortParts = []string{name, ""}
	} else {
		shortParts = strings.Split(name, ".")
	}

	if len(shortParts) == 1 {
		shortParts = append(shortParts, "")
	}

	if len(shortParts[0]) < 8 {
		var temp bytes.Buffer
		temp.WriteString(shortParts[0])
		for i := 0; i < 8-len(shortParts[0]); i++ {
			temp.WriteRune(' ')
		}

		shortParts[0] = temp.String()
	}

	if len(shortParts[1]) < 3 {
		var temp bytes.Buffer
		temp.WriteString(shortParts[1])
		for i := 0; i < 3-len(shortParts[1]); i++ {
			temp.WriteRune(' ')
		}

		shortParts[1] = temp.String()
	}

	return fmt.Sprintf("%s%s", shortParts[0], shortParts[1])
}

func cleanShortString(v string) string {
	var result bytes.Buffer
	for _, char := range v {
		// We skip these chars
		if char == '.' || char == ' ' {
			continue
		}

		if !validShortChar(char) {
			char = '_'
		}

		result.WriteRune(char)
	}

	return result.String()
}

func validShortChar(char rune) bool {
	if char >= 'A' && char <= 'Z' {
		return true
	}

	if char >= '0' && char <= '9' {
		return true
	}

	validShortSymbols := []rune{
		'_', '^', '$', '~', '!', '#', '%', '&', '-', '{', '}', '(',
		')', '@', '\'', '`',
	}

	for _, valid := range validShortSymbols {
		if char == valid {
			return true
		}
	}

	return false
}
