package cli

import (
	"fmt"

	"quibit/internal/input"
	"quibit/internal/model"
)

func RunWizard() (model.ProjectInput, error) {
	in, err := input.CollectProjectInput()
	if err != nil {
		return model.ProjectInput{}, fmt.Errorf("wizard: %w", err)
	}
	return in, nil
}
