package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Money struct {
	amount   int64
	currency string
}

func NewMoneyFromString(s string, currency string) (Money, error) {
	if s == "" {
		return Money{}, ErrEmptyMoneyString
	}
	if strings.ContainsAny(s, "eE+") {
		return Money{}, ErrInvalidMoneyFormat
	}
	if strings.HasPrefix(s, "-") {
		return Money{}, ErrNegativeExternal
	}
	normalized := strings.ReplaceAll(s, ",", ".")
	parts := strings.Split(normalized, ".")
	if len(parts) > 2 {
		return Money{}, ErrMultipleSeparators
	}
	intPartStr := parts[0]
	decPartStr := ""
	if len(parts) == 2 {
		decPartStr = parts[1]
	}
	if len(decPartStr) != 2 {
		return Money{}, &ErrInvalidMoneyDecimalPlaces{Expected: 2, Got: len(decPartStr)}
	}
	fullStr := intPartStr + decPartStr
	var amount int64
	_, err := fmt.Sscanf(fullStr, "%d", &amount)
	if err != nil {
		return Money{}, &ErrMoneyOverflow{Err: err}
	}
	if err := validateCurrency(currency); err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: currency}, nil
}

func Zero(currency string) (Money, error) {
	if err := validateCurrency(currency); err != nil {
		return Money{}, err
	}
	return Money{amount: 0, currency: currency}, nil
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, &ErrMoneyCurrencyMismatch{
			Operation: "add",
			CurrencyA: m.currency,
			CurrencyB: other.currency,
		}
	}
	if (other.amount > 0 && m.amount > math.MaxInt64-other.amount) ||
		(other.amount < 0 && m.amount < math.MinInt64-other.amount) {
		return Money{}, ErrOverflowAddition
	}
	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, &ErrMoneyCurrencyMismatch{
			Operation: "subtract",
			CurrencyA: m.currency,
			CurrencyB: other.currency,
		}
	}
	if (other.amount < 0 && m.amount > math.MaxInt64+other.amount) ||
		(other.amount > 0 && m.amount < math.MinInt64+other.amount) {
		return Money{}, ErrOverflowSubtraction
	}
	return Money{
		amount:   m.amount - other.amount,
		currency: m.currency,
	}, nil
}

func (m Money) Negate() (Money, error) {
	if m.amount == math.MinInt64 {
		return Money{}, ErrOverflowNegation
	}
	return Money{
		amount:   -m.amount,
		currency: m.currency,
	}, nil
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

func (m Money) MarshalJSON() ([]byte, error) {
	dollars := m.amount / 100
	cents := m.amount % 100
	if cents < 0 {
		cents = -cents
	}
	formatted := fmt.Sprintf(`{"amount":"%d.%02d","currency":"%s"}`, dollars, cents, m.currency)
	return []byte(formatted), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var aux struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	parsed, err := NewMoneyFromString(aux.Amount, aux.Currency)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

func validateCurrency(currency string) error {
	if len(currency) != 3 {
		return &ErrInvalidCurrency{
			Currency: currency,
			Reason:   "must be a 3-letter ISO 4217 code",
		}
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return &ErrInvalidCurrency{
				Currency: currency,
				Reason:   "must contain only uppercase ASCII letters",
			}
		}
	}
	return nil
}