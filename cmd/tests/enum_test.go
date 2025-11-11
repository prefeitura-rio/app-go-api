package tests

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestModalidadeCurso(t *testing.T) {
	// Teste de enums da Modalidade - legacy values
	assert.Equal(t, "PRESENCIAL", string(models.ModalidadePresencialLegacy), "Modalidade presencial legacy deve ser PRESENCIAL")
	assert.Equal(t, "ONLINE", string(models.ModalidadeOnline), "Modalidade online deve ser ONLINE")
	assert.Equal(t, "HIBRIDO", string(models.ModalidadeHibrido), "Modalidade híbrido deve ser HIBRIDO")

	// Teste de enums da Modalidade - new values
	assert.Equal(t, "Presencial", string(models.ModalidadePresencial), "Modalidade presencial deve ser Presencial")
	assert.Equal(t, "Semipresencial", string(models.ModalidadeSemipresencial), "Modalidade semipresencial deve ser Semipresencial")
	assert.Equal(t, "Remoto", string(models.ModalidadeRemoto), "Modalidade remoto deve ser Remoto")
}

func TestStatusCurso(t *testing.T) {
	// Teste de enums do Status do Curso - legacy values
	assert.Equal(t, "CRIADO", string(models.StatusCursoCriado), "Status criado deve ser CRIADO")
	assert.Equal(t, "ABERTO", string(models.StatusCursoAberto), "Status aberto deve ser ABERTO")
	assert.Equal(t, "ENCERRADO", string(models.StatusCursoEncerrado), "Status encerrado deve ser ENCERRADO")

	// Teste de enums do Status do Curso - new values
	assert.Equal(t, "draft", string(models.StatusCursoDraft), "Status draft deve ser draft")
	assert.Equal(t, "opened", string(models.StatusCursoOpened), "Status opened deve ser opened")
	assert.Equal(t, "closed", string(models.StatusCursoClosed), "Status closed deve ser closed")
	assert.Equal(t, "canceled", string(models.StatusCursoCanceled), "Status canceled deve ser canceled")
}

func TestFormatoAula(t *testing.T) {
	// Teste de enums do Formato de Aula
	assert.Equal(t, "GRAVADO", string(models.FormatoAulaGravado), "Formato gravado deve ser GRAVADO")
	assert.Equal(t, "AO_VIVO", string(models.FormatoAulaAoVivo), "Formato ao vivo deve ser AO_VIVO")
}

func TestTipoContratacao(t *testing.T) {
	// Teste de enums do Tipo de Contratação
	assert.Equal(t, "CLT", string(models.TipoContratacaoCLT), "Tipo CLT deve ser CLT")
	assert.Equal(t, "ESTAGIO", string(models.TipoContratacaoEstagio), "Tipo estágio deve ser ESTAGIO")
	assert.Equal(t, "JOVEM_APRENDIZ", string(models.TipoContratacaoJovemAprendiz), "Tipo jovem aprendiz deve ser JOVEM_APRENDIZ")
	assert.Equal(t, "MEI", string(models.TipoContratacaoMEI), "Tipo MEI deve ser MEI")
	assert.Equal(t, "PJ", string(models.TipoContratacaoPJ), "Tipo PJ deve ser PJ")
}

func TestStatusEmprego(t *testing.T) {
	// Teste de enums do Status do Emprego
	assert.Equal(t, "CRIADO", string(models.StatusEmpregoCriado), "Status criado deve ser CRIADO")
	assert.Equal(t, "ABERTO", string(models.StatusEmpregoAberto), "Status aberto deve ser ABERTO")
	assert.Equal(t, "EM_ANALISE", string(models.StatusEmpregoEmAnalise), "Status em análise deve ser EM_ANALISE")
	assert.Equal(t, "ENCERRADO", string(models.StatusEmpregoEncerrado), "Status encerrado deve ser ENCERRADO")
}

func TestJornadaTrabalho(t *testing.T) {
	// Teste de enums da Jornada de Trabalho
	assert.Equal(t, "INTEGRAL", string(models.JornadaIntegral), "Jornada integral deve ser INTEGRAL")
	assert.Equal(t, "MEIO_PERIODO", string(models.JornadaMeioPeriodo), "Jornada meio período deve ser MEIO_PERIODO")
	assert.Equal(t, "FLEXIVEL", string(models.JornadaFlexivel), "Jornada flexível deve ser FLEXIVEL")
	assert.Equal(t, "ESTAGIO", string(models.JornadaEstagio), "Jornada estágio deve ser ESTAGIO")
	assert.Equal(t, "HOME_OFFICE", string(models.JornadaHomeOffice), "Jornada home office deve ser HOME_OFFICE")
}

func TestTurno(t *testing.T) {
	// Teste de enums do Turno (compartilhado)
	assert.Equal(t, "MANHA", string(models.TurnoManha), "Turno manhã deve ser MANHA")
	assert.Equal(t, "TARDE", string(models.TurnoTarde), "Turno tarde deve ser TARDE")
	assert.Equal(t, "NOITE", string(models.TurnoNoite), "Turno noite deve ser NOITE")
	assert.Equal(t, "INTEGRAL", string(models.TurnoIntegral), "Turno integral deve ser INTEGRAL")
	assert.Equal(t, "LIVRE", string(models.TurnoLivre), "Turno livre deve ser LIVRE")
}

func TestStatusInscricao(t *testing.T) {
	// Teste de enums do Status da Inscrição
	assert.Equal(t, "pending", string(models.StatusInscricaoPending), "Status pending deve ser pending")
	assert.Equal(t, "approved", string(models.StatusInscricaoApproved), "Status approved deve ser approved")
	assert.Equal(t, "rejected", string(models.StatusInscricaoRejected), "Status rejected deve ser rejected")
	assert.Equal(t, "cancelled", string(models.StatusInscricaoCancelled), "Status cancelled deve ser cancelled")
}
