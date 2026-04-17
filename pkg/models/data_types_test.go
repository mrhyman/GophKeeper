package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginPassword_JSON(t *testing.T) {
	lp := LoginPassword{
		Login:    "user@example.com",
		Password: "secret123",
		URI:      "https://example.com",
	}

	data, err := json.Marshal(lp)
	require.NoError(t, err)

	var decoded LoginPassword
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, lp.Login, decoded.Login)
	assert.Equal(t, lp.Password, decoded.Password)
	assert.Equal(t, lp.URI, decoded.URI)
}

func TestLoginPassword_JSON_OmitEmptyURI(t *testing.T) {
	lp := LoginPassword{
		Login:    "user",
		Password: "pass",
	}

	data, err := json.Marshal(lp)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "uri")
}

func TestTextData_JSON(t *testing.T) {
	td := TextData{
		Text: "some secret text",
	}

	data, err := json.Marshal(td)
	require.NoError(t, err)

	var decoded TextData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, td.Text, decoded.Text)
}

func TestBinaryData_JSON(t *testing.T) {
	bd := BinaryData{
		FileName: "secret.pdf",
		Data:     []byte("binary content"),
	}

	data, err := json.Marshal(bd)
	require.NoError(t, err)

	var decoded BinaryData
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, bd.FileName, decoded.FileName)
	assert.Equal(t, bd.Data, decoded.Data)
}

func TestBankCard_JSON(t *testing.T) {
	bc := BankCard{
		Number:   "4111111111111111",
		Holder:   "JOHN DOE",
		ExpMonth: 12,
		ExpYear:  2025,
		CVV:      "123",
	}

	data, err := json.Marshal(bc)
	require.NoError(t, err)

	var decoded BankCard
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, bc.Number, decoded.Number)
	assert.Equal(t, bc.Holder, decoded.Holder)
	assert.Equal(t, bc.ExpMonth, decoded.ExpMonth)
	assert.Equal(t, bc.ExpYear, decoded.ExpYear)
	assert.Equal(t, bc.CVV, decoded.CVV)
}

func TestBankCard_JSONFieldNames(t *testing.T) {
	bc := BankCard{
		Number:   "4111111111111111",
		Holder:   "JOHN DOE",
		ExpMonth: 12,
		ExpYear:  2025,
		CVV:      "123",
	}

	data, err := json.Marshal(bc)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"number"`)
	assert.Contains(t, jsonStr, `"holder"`)
	assert.Contains(t, jsonStr, `"exp_month"`)
	assert.Contains(t, jsonStr, `"exp_year"`)
	assert.Contains(t, jsonStr, `"cvv"`)
}
