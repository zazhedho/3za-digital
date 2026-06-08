package h2h

import "encoding/json"

func unmarshalWithWire[T any, W any](data []byte, target *T, wire *W, mapFn func(*T, *W)) error {
	if err := json.Unmarshal(data, wire); err != nil {
		return err
	}
	if mapFn != nil {
		mapFn(target, wire)
	}
	return nil
}
