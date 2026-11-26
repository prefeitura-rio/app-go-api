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

func ptrBool(v bool) *bool { return &v }

// jovem_aprendiz entra no UPDATE como booleano quando a vaga o traz.
func TestVagaRepository_Update_JovemAprendiz(t *testing.T) {
	for _, valor := range []bool{true, false} {
		for _, metodo := range []string{"Update", "UpdateWithAssociations"} {
			t.Run(metodo, func(t *testing.T) {
				db, mock, cleanup := repository.SetupMockDB(t)
				defer cleanup()

				vaga := &empregabilidade.Vaga{
					ID:                  uuid.New(),
					Titulo:              "Auxiliar administrativo",
					IDContratante:       "12345678000190",
					IDRegimeContratacao: uuid.New(),
					IDModeloTrabalho:    uuid.New(),
					Status:              empregabilidade.StatusVagaEmEdicao,
					JovemAprendiz:       ptrBool(valor),
				}

				// Ordem alfabética das colunas do map: jovem_aprendiz é a 18ª,
				// depois de idade_minima; o id do WHERE é o 27º argumento.
				args := make([]driver.Value, 0, 27)
				for i := 0; i < 17; i++ {
					args = append(args, sqlmock.AnyArg())
				}
				args = append(args, valor)
				for i := 0; i < 8; i++ { // opcao_pcd ... valor_vaga
					args = append(args, sqlmock.AnyArg())
				}
				args = append(args, vaga.ID)

				mock.ExpectBegin()
				mock.ExpectExec(`(?s)UPDATE "emp_vagas" SET .*"idade_minima"=\$17,"jovem_aprendiz"=\$18,"opcao_pcd"=\$19,.*WHERE id = \$27`).
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

// Sem o campo na entidade, o UPDATE não toca a coluna: o valor gravado fica.
func TestVagaRepository_Update_JovemAprendizAusenteNaoAltera(t *testing.T) {
	for _, metodo := range []string{"Update", "UpdateWithAssociations"} {
		t.Run(metodo, func(t *testing.T) {
			db, mock, cleanup := repository.SetupMockDB(t)
			defer cleanup()

			vaga := &empregabilidade.Vaga{
				ID:                  uuid.New(),
				Titulo:              "Auxiliar administrativo",
				IDContratante:       "12345678000190",
				IDRegimeContratacao: uuid.New(),
				IDModeloTrabalho:    uuid.New(),
				Status:              empregabilidade.StatusVagaEmEdicao,
			}

			args := make([]driver.Value, 0, 26)
			for i := 0; i < 25; i++ {
				args = append(args, sqlmock.AnyArg())
			}
			args = append(args, vaga.ID)

			mock.ExpectBegin()
			mock.ExpectExec(`(?s)UPDATE "emp_vagas" SET .*"idade_minima"=\$17,"opcao_pcd"=\$18,.*WHERE id = \$26`).
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

// O INSERT grava jovem_aprendiz entre experiencia_carteira e
// criterios_bloqueio (ordem do struct), desligado quando a vaga não o traz.
func TestVagaRepository_Create_JovemAprendiz(t *testing.T) {
	tests := []struct {
		name  string
		valor *bool
		arg   driver.Value
	}{
		{name: "marcado", valor: ptrBool(true), arg: true},
		{name: "desmarcado", valor: ptrBool(false), arg: false},
		{name: "ausente nasce desligado", valor: nil, arg: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := repository.SetupMockDB(t)
			defer cleanup()

			vaga := &empregabilidade.Vaga{
				Titulo:              "Auxiliar administrativo",
				IDContratante:       "12345678000190",
				IDRegimeContratacao: uuid.New(),
				IDModeloTrabalho:    uuid.New(),
				Status:              empregabilidade.StatusVagaEmEdicao,
				JovemAprendiz:       tt.valor,
			}

			// 28 colunas: jovem_aprendiz é a 27ª, antes de criterios_bloqueio.
			args := make([]driver.Value, 0, 28)
			for i := 0; i < 26; i++ {
				args = append(args, sqlmock.AnyArg())
			}
			args = append(args, tt.arg, sqlmock.AnyArg())

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO "emp_vagas" \(.*"experiencia_carteira","jovem_aprendiz","criterios_bloqueio"\) VALUES .*\$28\)`).
				WithArgs(args...).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			mock.ExpectCommit()

			_, err := NewVagaRepository(db).Create(context.Background(), vaga)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
