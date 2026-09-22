package utils

// MaskCPFForLog masks a CPF for logging (shows only first 3 and last 2 digits).
func MaskCPFForLog(cpf string) string {
	if len(cpf) < 5 {
		return "***"
	}
	return cpf[:3] + "******" + cpf[len(cpf)-2:]
}
