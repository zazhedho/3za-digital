package qrisly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FlexibleString string

func (s *FlexibleString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*s = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = FlexibleString(text)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*s = FlexibleString(number.String())
		return nil
	}
	return fmt.Errorf("unsupported string value: %s", string(data))
}

func (s FlexibleString) String() string {
	return string(s)
}

type FlexibleInt64 int64

func (i *FlexibleInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*i = 0
		return nil
	}
	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*i = FlexibleInt64(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			*i = 0
			return nil
		}
		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return err
		}
		*i = FlexibleInt64(parsed)
		return nil
	}
	return fmt.Errorf("unsupported int value: %s", string(data))
}

func (i FlexibleInt64) Int64() int64 {
	return int64(i)
}
