package main

func subtractArrays(array1, array2 []string) []string {
	var result []string

	// Create a map to efficiently check if an element is in array2
	exists := make(map[string]bool)
	for _, elem := range array2 {
		exists[elem] = true
	}

	// Iterate through array1 and add elements not in array2 to the result
	for _, elem := range array1 {
		if !exists[elem] {
			result = append(result, elem)
		}
	}

	return result
}
