package value

import (
	"fmt"

	"github.com/rs/xid"
)

type ExampleID struct{ xid.ID }

func ParseExampleID(s string) (ExampleID, error) {
	id, err := xid.FromString(s)
	if err != nil {
		return ExampleID{}, fmt.Errorf("xid.FromString(%s): %w", s, err)
	}

	return ExampleID{ID: id}, nil
}
