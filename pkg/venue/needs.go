package venue

import (
	"fmt"
	"io"
	"os"

	"github.com/auroq/cnh/pkg/model"
	"gopkg.in/yaml.v3"
)

// Need is the model type; this package owns the needs file's shape.
type Need = model.Need

type needYAML struct {
	ID      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Phrases []string `yaml:"phrases"`
	Kinds   []string `yaml:"kinds"`
	Note    string   `yaml:"note"`
}

func (n needYAML) model() model.Need {
	return model.Need{ID: n.ID, Name: n.Name, Phrases: n.Phrases, Kinds: n.Kinds, Note: n.Note}
}

type needsFile struct {
	Needs []needYAML `yaml:"needs"`
}

// LoadNeeds reads needs.yaml and validates that every need has an id, a name, at least one phrase and at least one kind.
func LoadNeeds(r io.Reader) ([]Need, error) {
	var f needsFile
	if err := yaml.NewDecoder(r).Decode(&f); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for i, n := range f.Needs {
		switch {
		case n.ID == "":
			return nil, fmt.Errorf("need %d: missing id", i)
		case seen[n.ID]:
			return nil, fmt.Errorf("need %s: duplicate id", n.ID)
		case n.Name == "" || len(n.Phrases) == 0 || len(n.Kinds) == 0:
			return nil, fmt.Errorf("need %s: name, phrases and kinds are required", n.ID)
		}
		seen[n.ID] = true
	}
	needs := make([]Need, 0, len(f.Needs))
	for _, n := range f.Needs {
		needs = append(needs, n.model())
	}
	return needs, nil
}

// NeedsFile is a needs.yaml on disk, loaded on demand.
type NeedsFile string

// Needs loads and validates the file.
func (f NeedsFile) Needs() ([]Need, error) {
	r, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return LoadNeeds(r)
}
