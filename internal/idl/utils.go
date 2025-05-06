package idl

import (
	"encoding/json"
	"fmt"
)

// transcodeJSON converts the input to json, and then unmarshals it into the
// destination. The destination must be a pointer.
func transcodeJSON(input any, destinationPointer any) error {
	b, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error while marshaling input to json: %s", err)
	}

	err = json.Unmarshal(b, destinationPointer)
	if err != nil {
		return fmt.Errorf("error while unmarshaling json to destination: %s", err)
	}
	return nil
}
