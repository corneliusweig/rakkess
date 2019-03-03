package util

import (
	"fmt"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
)

func ValidateOutputFormat(format string) error {
	for _, o := range constants.ValidOutputFormats {
		if o == format {
			return nil
		}
	}
	return fmt.Errorf("unexpected output format: %s", format)
}
