package utils

func ReadFromData(maxRead int, data []byte) (int, []byte) {
	if data == nil {
		return -1, nil
	}
	if len(data) < maxRead {
		return len(data), data
	} else {
		return maxRead, data[:maxRead]
	}
}
