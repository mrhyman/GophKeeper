package models

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URI      string `json:"uri,omitempty"`
}

type TextData struct {
	Text string `json:"text"`
}

type BinaryData struct {
	FileName string `json:"file_name"`
	Data     []byte `json:"data"`
}

type BankCard struct {
	Number   string `json:"number"`
	Holder   string `json:"holder"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	CVV      string `json:"cvv"`
}
