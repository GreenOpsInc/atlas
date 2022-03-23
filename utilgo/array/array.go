package array

func LastIndexOf(arr []string, val string) int {
	for id, el := range arr {
		if el == val {
			return id
		}
	}
	return -1
}

func ContainElement(arr []string, val string) bool {
	for _, el := range arr {
		if el == val {
			return true
		}
	}
	return false
}
