package repository_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Os testes deste arquivo exercitam, contra um Postgres real, a migration de
// dados da DEV-120: as inscrições dos cursos 418 e 420 de staging que apontam
// para uma turma apagada voltam a apontar para a turma atual idêntica, e
// nenhuma outra inscrição muda. Gated by RUN_REPOSITORY_INTEGRATION.

const migrationReligaOrfasDEV120 = "20261119120000_religa_inscricoes_orfas_dev120.sql"

// Turmas remotas atuais dos cursos 418 e 420 em staging: são elas que travam o
// escopo da migration (não existem em produção).
const (
	remoteClassCurso418 = "0f335b86-ab22-4c4a-b9cf-81bbc5be7713"
	remoteClassCurso420 = "df1f0580-2112-414e-8245-98a1563168d2"
)

// sqlDaMigrationDEV120 extrai o UPDATE da migration, para testar o SQL exato
// que vai rodar no goose up do pod.
func sqlDaMigrationDEV120(t *testing.T) string {
	t.Helper()
	conteudo, err := os.ReadFile(filepath.Join("..", "db", "migrations", migrationReligaOrfasDEV120))
	require.NoError(t, err)

	up := strings.Split(string(conteudo), "-- +goose Down")[0]
	blocos := strings.Split(up, "-- +goose StatementBegin")
	require.Len(t, blocos, 2, "a migration tem um único bloco StatementBegin no Up")
	sql := strings.Split(blocos[1], "-- +goose StatementEnd")[0]
	require.Contains(t, sql, "UPDATE inscricoes")
	return sql
}

type fixtureOrfasDEV120 struct {
	tx *gorm.DB
	n  int
}

func (f *fixtureOrfasDEV120) curso(t *testing.T, id int) {
	t.Helper()
	require.NoError(t, f.tx.Exec(
		`INSERT INTO cursos (id, titulo, modalidade, status) VALUES (?, 'Curso DEV-120', 'ONLINE', 'finished')`, id).Error)
}

func (f *fixtureOrfasDEV120) turmaRemota(t *testing.T, id string, cursoID int) {
	t.Helper()
	require.NoError(t, f.tx.Exec(`INSERT INTO remote_classes (id, curso_id) VALUES (?, ?)`, id, cursoID).Error)
}

// horario cria um horário da turma remota e devolve o id.
func (f *fixtureOrfasDEV120) horario(t *testing.T, remoteClassID string, vagas int, horario, dias, inicio, fim string) string {
	t.Helper()
	var id string
	require.NoError(t, f.tx.Raw(`INSERT INTO remote_schedules
		(remote_class_id, vacancies, class_time, class_days, class_start_date, class_end_date)
		VALUES (?, ?, ?, ?, ?::timestamptz, ?::timestamptz) RETURNING id`,
		remoteClassID, vagas, horario, dias, inicio, fim).Scan(&id).Error)
	return id
}

// inscricao grava uma inscrição sintética e devolve o id.
func (f *fixtureOrfasDEV120) inscricao(t *testing.T, cursoID int, scheduleID *string, enrolledUnit *string) string {
	t.Helper()
	f.n++
	cpf := fmt.Sprintf("9999990%04d", f.n)
	var id string
	require.NoError(t, f.tx.Raw(`INSERT INTO inscricoes (curso_id, cpf, name, status, schedule_id, enrolled_unit)
		VALUES (?, ?, 'Pessoa Sintética', 'pending', ?::uuid, ?::jsonb) RETURNING id`,
		cursoID, cpf, scheduleID, enrolledUnit).Scan(&id).Error)
	return id
}

func (f *fixtureOrfasDEV120) scheduleDe(t *testing.T, inscricaoID string) *string {
	t.Helper()
	var s *string
	require.NoError(t, f.tx.Raw(`SELECT schedule_id::text FROM inscricoes WHERE id = ?`, inscricaoID).Row().Scan(&s))
	return s
}

func strPtrDEV120(s string) *string { return &s }

// snapshotDEV120 monta o enrolled_unit gravado na inscrição, com um único horário.
func snapshotDEV120(remoteClassID, scheduleID string, vagas int, horario, dias, inicio, fim string) *string {
	return strPtrDEV120(fmt.Sprintf(
		`{"id":%q,"schedules":[{"id":%q,"location_id":%q,"vacancies":%d,"class_time":%q,"class_days":%q,"class_start_date":%q,"class_end_date":%q}]}`,
		remoteClassID, scheduleID, remoteClassID, vagas, horario, dias, inicio, fim))
}

func TestMigrationDEV120_Integration_ReligaSoAsOrfasInequivocas(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()

	var existentes int64
	require.NoError(t, tx.Raw(`SELECT count(*) FROM cursos WHERE id IN (418, 420, 999420)`).Scan(&existentes).Error)
	if existentes > 0 {
		t.Skip("o banco de teste já tem os cursos 418, 420 ou 999420")
	}

	f := &fixtureOrfasDEV120{tx: tx}
	const (
		apagada    = "a07f436b-0000-4000-8000-000000000001" // turma que o PUT apagou
		apagada418 = "a07f436b-0000-4000-8000-000000000418"
		inicio     = "2025-11-25T22:29:26.339-03:00"
		fim        = "2025-11-30T22:29:26-03:00"
	)

	// Curso 420: a turma atual idêntica à apagada e outra turma diferente.
	f.curso(t, 420)
	f.turmaRemota(t, remoteClassCurso420, 420)
	atual420 := f.horario(t, remoteClassCurso420, 35, "19:00 - 22:00", "Segunda", inicio, fim)
	outra420 := f.horario(t, remoteClassCurso420, 120, "14:00 - 15:00", "Segunda", "2025-11-26T17:36:00-03:00", "2025-11-26T17:36:00-03:00")

	// Snapshot do import (datas YYYY-MM-DD, no fuso UTC do processo) e do
	// serviço (RFC3339 no fuso de São Paulo): os dois religam.
	orfaImport := f.inscricao(t, 420, strPtrDEV120(apagada), snapshotDEV120(remoteClassCurso420, apagada, 35, "19:00 - 22:00", "Segunda", "2025-11-26", "2025-12-01"))
	orfaServico := f.inscricao(t, 420, strPtrDEV120(apagada), snapshotDEV120(remoteClassCurso420, apagada, 35, "19:00 - 22:00", "Segunda", inicio, fim))
	// Snapshot que não bate com nenhuma turma atual (vagas diferentes): fica.
	orfaSemPar := f.inscricao(t, 420, strPtrDEV120(apagada), snapshotDEV120(remoteClassCurso420, apagada, 40, "19:00 - 22:00", "Segunda", inicio, fim))
	// Sem snapshot: não há como provar o vínculo, fica.
	orfaSemSnapshot := f.inscricao(t, 420, strPtrDEV120(apagada), nil)
	// Inscrição já válida e inscrição sem turma: ficam.
	valida := f.inscricao(t, 420, strPtrDEV120(outra420), snapshotDEV120(remoteClassCurso420, outra420, 120, "14:00 - 15:00", "Segunda", "2025-11-26", "2025-11-26"))
	semTurma := f.inscricao(t, 420, nil, nil)

	// Curso 418: duas turmas atuais idênticas ao snapshot (par ambíguo): fica.
	f.curso(t, 418)
	f.turmaRemota(t, remoteClassCurso418, 418)
	f.horario(t, remoteClassCurso418, 10, "14:00 - 15:00", "Segunda", inicio, fim)
	f.horario(t, remoteClassCurso418, 10, "14:00 - 15:00", "Segunda", inicio, fim)
	atual418 := f.horario(t, remoteClassCurso418, 15, "15:00 - 17:00", "Terça", inicio, fim)
	orfaAmbigua := f.inscricao(t, 418, strPtrDEV120(apagada418), snapshotDEV120(remoteClassCurso418, apagada418, 10, "14:00 - 15:00", "Segunda", inicio, fim))
	orfa418 := f.inscricao(t, 418, strPtrDEV120(apagada418), snapshotDEV120(remoteClassCurso418, apagada418, 15, "15:00 - 17:00", "Terça", inicio, fim))

	// Curso fora do escopo (outra turma remota), com o mesmo formato: fica.
	const outraTurmaRemota = "11111111-2222-4333-8444-555555555555"
	f.curso(t, 999420)
	f.turmaRemota(t, outraTurmaRemota, 999420)
	f.horario(t, outraTurmaRemota, 35, "19:00 - 22:00", "Segunda", inicio, fim)
	foraDoEscopo := f.inscricao(t, 999420, strPtrDEV120(apagada), snapshotDEV120(outraTurmaRemota, apagada, 35, "19:00 - 22:00", "Segunda", inicio, fim))

	sql := sqlDaMigrationDEV120(t)
	res := tx.Exec(sql)
	require.NoError(t, res.Error)
	assert.Equal(t, int64(3), res.RowsAffected)

	assert.Equal(t, strPtrDEV120(atual420), f.scheduleDe(t, orfaImport))
	assert.Equal(t, strPtrDEV120(atual420), f.scheduleDe(t, orfaServico))
	assert.Equal(t, strPtrDEV120(atual418), f.scheduleDe(t, orfa418))
	assert.Equal(t, strPtrDEV120(apagada), f.scheduleDe(t, orfaSemPar))
	assert.Equal(t, strPtrDEV120(apagada), f.scheduleDe(t, orfaSemSnapshot))
	assert.Equal(t, strPtrDEV120(outra420), f.scheduleDe(t, valida))
	assert.Nil(t, f.scheduleDe(t, semTurma))
	assert.Equal(t, strPtrDEV120(apagada418), f.scheduleDe(t, orfaAmbigua))
	assert.Equal(t, strPtrDEV120(apagada), f.scheduleDe(t, foraDoEscopo))

	// Rodar de novo não muda nada (idempotente).
	res = tx.Exec(sql)
	require.NoError(t, res.Error)
	assert.Equal(t, int64(0), res.RowsAffected)
}

// Em produção os cursos têm outros ids e as turmas remotas de staging não
// existem: a migration não toca em nada.
func TestMigrationDEV120_Integration_SemAsTurmasDeStagingNaoMudaNada(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()

	var existentes int64
	require.NoError(t, tx.Raw(`SELECT count(*) FROM cursos WHERE id = 420`).Scan(&existentes).Error)
	if existentes > 0 {
		t.Skip("o banco de teste já tem o curso 420")
	}

	f := &fixtureOrfasDEV120{tx: tx}
	const apagada = "a07f436b-0000-4000-8000-000000000002"
	const turmaDeProducao = "22222222-3333-4444-8555-666666666666"
	// Mesmo id de curso, turma remota com outro UUID.
	f.curso(t, 420)
	f.turmaRemota(t, turmaDeProducao, 420)
	f.horario(t, turmaDeProducao, 35, "19:00 - 22:00", "Segunda", "2025-11-25T22:29:26-03:00", "2025-11-30T22:29:26-03:00")
	orfa := f.inscricao(t, 420, strPtrDEV120(apagada), snapshotDEV120(turmaDeProducao, apagada, 35, "19:00 - 22:00", "Segunda", "2025-11-25T22:29:26-03:00", "2025-11-30T22:29:26-03:00"))

	res := tx.Exec(sqlDaMigrationDEV120(t))
	require.NoError(t, res.Error)
	assert.Equal(t, int64(0), res.RowsAffected)
	assert.Equal(t, strPtrDEV120(apagada), f.scheduleDe(t, orfa))
}
