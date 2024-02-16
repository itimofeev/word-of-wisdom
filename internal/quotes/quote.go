package quotes

import (
	"bufio"
	"bytes"
	_ "embed"
	"math/rand"
	"strings"
)

//go:embed quotes.txt
var quotesSource []byte

type Quotes struct {
	quotesList []string
}

func New() (*Quotes, error) {
	reader := bytes.NewReader(quotesSource)
	s := bufio.NewScanner(reader)

	quotes := make([]string, 0)

	for s.Scan() {
		if quote := strings.TrimSpace(s.Text()); quote != "" {
			quotes = append(quotes, quote)
		}
	}

	return &Quotes{quotesList: quotes}, nil
}

func (q *Quotes) GetQuote() (string, error) {
	return q.quotesList[rand.Intn(len(q.quotesList))], nil
}
