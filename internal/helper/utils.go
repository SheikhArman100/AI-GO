package helper


func ConvertToMapSlice(input []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(input))
	for _, v := range input {
		if m, ok := v.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}