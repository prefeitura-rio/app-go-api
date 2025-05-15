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