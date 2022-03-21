package array

func LastIndexOf(arr []string, val string) int {
	for id, el := range arr {
		if el == val {
			return id
		}
	}
	return -1
}
