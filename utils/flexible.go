package utils

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

	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*s = FlexibleString(strconv.FormatBool(boolean))
		return nil
	}

	return fmt.Errorf("unsupported string value: %s", string(data))
}

func (s FlexibleString) String() string {
	return string(s)
}

type FlexibleNumber string

func (n *FlexibleNumber) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = ""
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*n = FlexibleNumber(strings.TrimSpace(text))
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*n = FlexibleNumber(number.String())
		return nil
	}

	return fmt.Errorf("unsupported number value: %s", string(data))
}

func (n FlexibleNumber) String() string {
	return string(n)
}

type FlexibleInt int

func (i *FlexibleInt) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*i = 0
		return nil
	}

	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*i = FlexibleInt(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			*i = 0
			return nil
		}
		parsed, err := strconv.Atoi(text)
		if err != nil {
			return err
		}
		*i = FlexibleInt(parsed)
		return nil
	}

	return fmt.Errorf("unsupported int value: %s", string(data))
}

func (i FlexibleInt) Int() int {
	return int(i)
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
