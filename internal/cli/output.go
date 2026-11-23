package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func WriteJSON(w io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}

	if _, err := fmt.Fprintln(w, string(data)); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	return nil
}

func FormatUSD(amount float64) string {
	return "$" + strconv.FormatFloat(amount, 'f', -1, 64)
}
