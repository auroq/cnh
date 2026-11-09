package venue

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// BlockedWordsFile is the path of the word list a report note may not contain.
type BlockedWordsFile string

// Words reads one word or phrase per line, lowercased and trimmed; blank lines and lines starting with # are skipped.
func (f BlockedWordsFile) Words() ([]string, error) {
	r, err := os.Open(string(f))
	if err != nil {
		return nil, fmt.Errorf("open blocked words %s: %w", string(f), err)
	}
	defer r.Close()
	return LoadBlockedWords(r)
}

func LoadBlockedWords(r io.Reader) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.ToLower(strings.TrimSpace(sc.Text()))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read blocked words: %w", err)
	}
	return out, nil
}
