package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arian-parser/internal/domain"
)

// ExtractFields applies each regex to the email body and returns the single capture group for each key
func ExtractFields(emailBody string, patterns map[string]*regexp.Regexp) (map[string]string, error) {
	out := make(map[string]string, len(patterns))
	for key, re := range patterns {
		m := re.FindStringSubmatch(emailBody)
		if len(m) < 2 {
			return nil, fmt.Errorf("field %q not found in text", key)
		}
		out[key] = m[1]
	}

	return out, nil
}

// ParseEmailDate normalizes a string into time.Time
func ParseEmailDate(raw string) (time.Time, error) {
	return time.Parse("January 2, 2006", raw)
}

// BuildTransaction assembles a domain.Transaction
func BuildTransaction(
	m EmailMeta,
	fields map[string]string,
	bank string,
	currency string,
	dir domain.Direction,
	desc string,
) (*domain.Transaction, error) {

	recv, err := time.Parse(time.RFC3339, m.Date)
	if err != nil {
		return nil, err
	}

	bodyDate, err := ParseEmailDate(fields["txdate"])
	if err != nil {
		return nil, err
	}

	// decide which timestamp to keep
	var final time.Time
	if recv.Year() == bodyDate.Year() && recv.YearDay() == bodyDate.YearDay() {
		final = recv
	} else {
		final = bodyDate
	}

	rawAmt := strings.ReplaceAll(fields["amount"], ",", "")
	amt, err := strconv.ParseFloat(rawAmt, 64)
	if err != nil {
		return nil, err
	}

	return &domain.Transaction{
		EmailID:         m.ID,
		TxDate:          final,
		TxBank:          bank,
		TxAccount:       fields["account"],
		TxAmount:        amt,
		TxCurrency:      currency,
		TxDirection:     dir,
		TxDesc:          desc,
		Category:        "",
		UserNotes:       "",
		ForeignAmount:   nil,
		ForeignCurrency: nil,
		ExchangeRate:    nil,
	}, nil
}
