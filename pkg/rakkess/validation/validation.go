package validation

import (
	"fmt"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Options(opts *options.RakkessOptions) error {
	if err := verbs(opts.Verbs); err != nil {
		return err
	}
	if err := outputFormat(opts.Output); err != nil {
		return err
	}
	return nil
}

func outputFormat(format string) error {
	for _, o := range constants.ValidOutputFormats {
		if o == format {
			return nil
		}
	}
	return fmt.Errorf("unexpected output format: %s", format)
}

func verbs(verbs []string) error {
	valid := sets.NewString(constants.ValidVerbs...)
	given := sets.NewString(verbs...)
	difference := given.Difference(valid)

	if difference.Len() > 0 {
		return fmt.Errorf("unexpected verbs: %s", difference.List())
	}

	return nil
}
