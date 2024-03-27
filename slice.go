package httprouter

func insert[T any](slice []T, index int, value T) []T {
	if len(slice) < 1 {
		return []T{value}
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}

func prepend[T any](slice []T, value T) []T {
	return append([]T{value}, slice...)
}