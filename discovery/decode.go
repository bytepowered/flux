package discovery

import (
	"errors"
	"fmt"
	"strings"
)

func checkjson(bytes []byte) error {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") {
		return fmt.Errorf("CHECK/JSON/SIZE: %d", size)
	}
	prefix := strings.TrimSpace(string(bytes[:5]))
	if prefix[0] != '[' && prefix[0] != '{' {
		return errors.New("CHECK/JSON/MALFORMED")
	}
	return nil
}
