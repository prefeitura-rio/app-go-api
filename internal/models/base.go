package models

// Tipo Turno compartilhado entre Curso e Emprego
type Turno string

const (
	TurnoManha    Turno = "MANHA"
	TurnoTarde    Turno = "TARDE"
	TurnoNoite    Turno = "NOITE"
	TurnoIntegral Turno = "INTEGRAL"
	TurnoLivre    Turno = "LIVRE"
)

// IsValid validates if the Turno value is valid
func (t Turno) IsValid() bool {
	validValues := []string{"MANHA", "TARDE", "NOITE", "INTEGRAL", "LIVRE"}
	for _, v := range validValues {
		if string(t) == v {
			return true
		}
	}
	return false
} 