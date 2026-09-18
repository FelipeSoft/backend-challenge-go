package domain

import (
	"encoding/json"
	"testing"
)

func TestNewMoneyFromString_Success(t *testing.T) {
	m, err := NewMoneyFromString("25.00", "BRL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Amount() != 2500 {
		t.Errorf("expected 2500 cents, got %d", m.Amount())
	}
	if m.Currency() != "BRL" {
		t.Errorf("expected BRL, got %s", m.Currency())
	}
}

func TestNewMoneyFromString_Rejections(t *testing.T) {
	invalidInputs := []string{
		"",
		"abc",
		"25.1",
		"25.005",
		"-25.00",
		"1e2",
		"25.00.00",
	}
	for _, input := range invalidInputs {
		_, err := NewMoneyFromString(input, "BRL")
		if err == nil {
			t.Errorf("expected error for input %q, got nil", input)
		}
	}
}

func TestMoney_CurrencyMismatch(t *testing.T) {
	m1, _ := NewMoneyFromString("10.00", "BRL")
	m2, _ := NewMoneyFromString("10.00", "USD")
	_, errAdd := m1.Add(m2)
	if errAdd == nil {
		t.Error("expected error when adding different currencies, got nil")
	}
	_, errSub := m1.Sub(m2)
	if errSub == nil {
		t.Error("expected error when subtracting different currencies, got nil")
	}
}

func TestMoney_JSONSerialization(t *testing.T) {
	m, _ := NewMoneyFromString("123.45", "BRL")
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	expectedJSON := `{"amount":"123.45","currency":"BRL"}`
	if string(data) != expectedJSON {
		t.Errorf("expected %s, got %s", expectedJSON, string(data))
	}
	var unmarshalled Money
	err = json.Unmarshal(data, &unmarshalled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !m.Equals(unmarshalled) {
		t.Errorf("unmarshalled object does not match original")
	}
}