package utils

import "strings"

// MaskCPFForLog masks a CPF for logging (shows only first 3 and last 2 digits).
func MaskCPFForLog(cpf string) string {
	if len(cpf) < 5 {
		return "***"
	}
	return cpf[:3] + "******" + cpf[len(cpf)-2:]
}

// MaskCNPJForLog masks a CNPJ for logging. Formatting is dropped and only the
// first 2 and last 2 digits are kept, so the root that identifies the company
// (and, for a MEI, its owner) never reaches the logs.
func MaskCNPJForLog(cnpj string) string {
	digits := []rune(ExtractDigits(cnpj))
	if len(digits) < 5 {
		return "***"
	}
	return string(digits[:2]) + strings.Repeat("*", len(digits)-4) + string(digits[len(digits)-2:])
}

// MaskEmailForLog masks an e-mail address for logging: the local part, which
// identifies the citizen, is dropped and only the lower-cased domain is kept
// ("***@example.com"). Values without a domain become "***".
func MaskEmailForLog(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return "***"
	}
	return "***@" + strings.ToLower(email[at+1:])
}
