package core

func includes(fields []string, field string) bool {
	for _, v := range fields {
		if v == field {
			return true
		}
	}

	return false
}