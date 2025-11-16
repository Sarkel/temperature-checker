package mqtt

import "strings"

const newLine = "\n"

func decode(payload []MessagePayload, separator string) string {
	res := make([]string, len(payload))

	for i, p := range payload {
		res[i] = strings.Join(p, separator)
	}

	return strings.Join(res, newLine)
}

func encode(msg string, separator string) []MessagePayload {
	parts := strings.Split(msg, newLine)

	res := make([]MessagePayload, len(parts))

	for i, p := range parts {
		res[i] = strings.Split(p, separator)
	}

	return res
}
