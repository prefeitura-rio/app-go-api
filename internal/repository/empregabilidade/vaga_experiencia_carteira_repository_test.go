package empregabilidade

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
)

func ptrExperienciaCarteira(e empregabilidade.ExperienciaCarteira) *empregabilidade.ExperienciaCarteira {
	return &e
}

// experiencia_carteira entra no UPDATE como texto, ou NULL sem exigência.
func TestVagaRepository_Update_ExperienciaCarteira(t *testing.T) {
	tests := []struct {
		name  string
		valor *empregabilidade.ExperienciaCarteira
		arg   driver.Value
	}{
		{name: "mínimo de 6 meses", valor: ptrExperienciaCarteira(empregabilidade.ExperienciaCarteiraMinimo6Meses), arg: "minimo_6_meses"},
		{name: "qualquer tempo", valor: ptrExperienciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo), arg: "qualquer_tempo"},
		{name: "sem exigência grava NULL", valor: nil, arg: nil},
	}

	for _, tt := range tests {
		for _, metodo := range []string{"Update", "UpdateWithAssociations"} {
			t.Run(metodo+"/"+tt.name, func(t *testing.T) {
				db, mock, cleanup := repository.SetupMockDB(t)
				defer cleanup()

				vaga := &empregabilidade.Vaga{
					ID:                  uuid.New(),
					Titulo:              "Repositor",
					IDContratante:       "12345678000190",
					IDRegimeContratacao: uuid.New(),
					IDModeloTrabalho:    uuid.New(),
					Status:              empregabilidade.StatusVagaEmEdicao,
					ExperienciaCarteira: tt.valor,
				}

				// Ordem alfabética das colunas do map: experiencia_carteira é a
				// 10ª, depois de diferenciais; o id do WHERE é o 26º argumento.
				args := make([]driver.Value, 0, 26)
				for i := 0; i < 9; i++ {
					args = append(args, sqlmock.AnyArg())
				}
				if tt.arg == nil {
					args = append(args, nullArg{})
				} else {
					args = append(args, tt.arg)
				}
				for i := 0; i < 15; i++ { // id_contratante ... valor_vaga
					args = append(args, sqlmock.AnyArg())
				}
				args = append(args, vaga.ID)

				mock.ExpectBegin()
				mock.ExpectExec(`(?s)UPDATE "emp_vagas" SET .*"diferenciais"=\$9,"experiencia_carteira"=\$10,"id_contratante"=\$11,.*WHERE id = \$26`).
					WithArgs(args...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				repo := NewVagaRepository(db)
				var err error
				if metodo == "Update" {
					err = repo.Update(context.Background(), vaga)
				} else {
					err = repo.UpdateWithAssociations(context.Background(), vaga)
				}
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}

// O INSERT grava experiencia_carteira logo antes de criterios_bloqueio, as
// duas últimas das 27 colunas.
func TestVagaRepository_Create_ExperienciaCarteira(t *testing.T) {
	tests := []struct {
		name  string
		valor *empregabilidade.ExperienciaCarteira
		arg   sqlmock.Argument
	}{
		{name: "com exigência", valor: ptrExperienciaCarteira(empregabilidade.ExperienciaCarteiraMinimo6Meses), arg: textoArg{want: "minimo_6_meses"}},
		{name: "sem exigência grava NULL", valor: nil, arg: nullArg{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := repository.SetupMockDB(t)
			defer cleanup()

			vaga := &empregabilidade.Vaga{
				Titulo:              "Repositor",
				IDContratante:       "12345678000190",
				IDRegimeContratacao: uuid.New(),
				IDModeloTrabalho:    uuid.New(),
				Status:              empregabilidade.StatusVagaEmEdicao,
				ExperienciaCarteira: tt.valor,
			}

			args := make([]driver.Value, 0, 27)
			for i := 0; i < 25; i++ {
				args = append(args, sqlmock.AnyArg())
			}
			args = append(args, tt.arg, sqlmock.AnyArg())

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO "emp_vagas" \(.*"experiencia_carteira","criterios_bloqueio"\) VALUES .*\$27\)`).
				WithArgs(args...).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			mock.ExpectCommit()

			_, err := NewVagaRepository(db).Create(context.Background(), vaga)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// textoArg aceita o valor como string, mesmo vindo de um tipo nomeado
// (o driver recebe o valor já convertido).
type textoArg struct{ want string }

func (m textoArg) Match(v driver.Value) bool {
	switch t := v.(type) {
	case string:
		return t == m.want
	case []byte:
		return string(t) == m.want
	}
	return false
}
