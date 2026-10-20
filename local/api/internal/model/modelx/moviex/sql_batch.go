package moviex

const mysqlStatementPlaceholderLimit = 65535

func sqlValueBatches[T any](values []T, reservedPlaceholders int) [][]T {
	if len(values) == 0 {
		return nil
	}
	batchSize := mysqlStatementPlaceholderLimit - reservedPlaceholders
	if batchSize <= 0 {
		batchSize = 1
	}
	batches := make([][]T, 0, (len(values)+batchSize-1)/batchSize)
	for start := 0; start < len(values); start += batchSize {
		end := start + batchSize
		if end > len(values) {
			end = len(values)
		}
		batches = append(batches, values[start:end])
	}
	return batches
}
