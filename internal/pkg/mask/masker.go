package mask

import (
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config"
)

type Masker struct {
	masks []mask
}

func New(cfg *config.Masking) (*Masker, error) {
	if cfg == nil {
		return nil, nil
	}

	fields, err := parseFields(cfg.ProcessFields, cfg.IgnoreFields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fields: %w", err)
	}

	masks, err := compileMasks(cfg.Masks, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to compile masks: %w", err)
	}

	return &Masker{
		masks: masks,
	}, nil
}

func (m *Masker) Mask(event map[string]string) {
	if len(event) == 0 {
		return
	}

	for _, mask := range m.masks {
		mask.apply(event)
	}
}
