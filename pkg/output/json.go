package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func MustPrintJSON(v interface{}) {
	if err := PrintJSON(v); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding json: %v\n", err)
	}
}

func ToJSONString(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
