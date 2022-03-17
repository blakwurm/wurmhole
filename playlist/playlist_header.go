package playlist

import (
	"regexp"
	"strconv"
	"strings"
)

type HeaderValue struct {
	value string
}

func ParseHeader(value string) (string, HeaderValue) {
	split := strings.SplitN(value[1:], ":", 2)
	return split[0], ToHeaderValue(split[1])
}

func ToHeaderValue(value string) HeaderValue {
	return HeaderValue{value}
}

func (h HeaderValue) String() string {
	return h.value
}

func (h HeaderValue) Bool() bool {
	return strings.ToLower(h.value) == "yes" ||
		strings.ToLower(h.value) == "true"
}

func (h HeaderValue) Float(bits int) (float64, error) {
	return strconv.ParseFloat(h.value, bits)
}

func (h HeaderValue) Int(bits int) (int64, error) {
	return strconv.ParseInt(h.value, 10, bits)
}

func (h HeaderValue) Range() (int, int) {
	split := strings.SplitN(h.value, "@", 2)
	start, err := strconv.Atoi(split[0])
	if err != nil {
		return 0, 0
	}

	end, err := strconv.Atoi(split[1])
	if err != nil {
		return 0, 0
	}

	return start, end
}

func (h HeaderValue) Param() (string, HeaderValue) {
	split := strings.SplitN(h.value, "=", 2)
	return split[0], ToHeaderValue(split[1])
}

func (h HeaderValue) Section() map[string]HeaderValue {
	regex := regexp.MustCompile(` *, *`)
	deconstructed := regex.Split(h.value, -1)
	values := make(map[string]HeaderValue)

	for _, value := range deconstructed {
		split := strings.SplitN(value, "=", 2)
		values[split[0]] = ToHeaderValue(split[1])
	}

	return values
}
