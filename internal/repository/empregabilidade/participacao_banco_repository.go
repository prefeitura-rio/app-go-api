package empregabilidade

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// janelaCandidaturaNoCadastro: currículo criado até esse tempo antes da primeira
// candidatura própria foi montado para ela (veio pela vaga); antes disso, a
// pessoa já tinha preenchido o currículo pelo app.
const janelaCandidaturaNoCadastro = "24 hours"

// GetParticipacaoBanco retorna a escolha do cidadão sobre o banco de currículos,
// ou nil se ele ainda não tem currículo.
func (r *CurriculoRepository) GetParticipacaoBanco(ctx context.Context, cpf string) (*empregabilidade.ParticipacaoBanco, error) {
	var linhas []struct {
		DisponivelBanco           bool
		DisponivelBancoAlteradoEm sql.NullTime
		OrigemCadastro            sql.NullString
	}
	err := r.db.WithContext(ctx).Raw(
		`SELECT disponivel_banco, disponivel_banco_alterado_em, origem_cadastro FROM emp_curriculos WHERE cpf = ?`,
		apenasDigitos(cpf),
	).Scan(&linhas).Error
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar participação no banco de currículos: %w", err)
	}
	if len(linhas) == 0 {
		return nil, nil
	}

	linha := linhas[0]
	participacao := &empregabilidade.ParticipacaoBanco{Disponivel: linha.DisponivelBanco}
	if linha.DisponivelBancoAlteradoEm.Valid {
		alterado := linha.DisponivelBancoAlteradoEm.Time
		participacao.AlteradoEm = &alterado
	}
	if linha.OrigemCadastro.Valid {
		origem := linha.OrigemCadastro.String
		participacao.OrigemCadastro = &origem
	}
	return participacao, nil
}

// AtualizarParticipacaoBanco mostra ou oculta o currículo no banco. A data só
// muda quando o estado muda. Sem currículo, não há o que mudar.
func (r *CurriculoRepository) AtualizarParticipacaoBanco(ctx context.Context, cpf string, disponivel bool, quando time.Time) error {
	err := r.db.WithContext(ctx).Exec(
		`UPDATE emp_curriculos SET disponivel_banco = ?, disponivel_banco_alterado_em = ? WHERE cpf = ? AND disponivel_banco <> ?`,
		disponivel, quando, apenasDigitos(cpf), disponivel,
	).Error
	if err != nil {
		return fmt.Errorf("erro ao atualizar participação no banco de currículos: %w", err)
	}
	return nil
}

// RegistrarOrigemCadastro grava, na primeira candidatura própria, como o
// currículo chegou à base: pela vaga, se foi criado há até 24h, ou pelo app, se
// já existia antes. Depois disso não muda. Sem currículo, não há o que marcar.
func (r *CurriculoRepository) RegistrarOrigemCadastro(ctx context.Context, cpf string) error {
	err := r.db.WithContext(ctx).Exec(
		`UPDATE emp_curriculos SET origem_cadastro = CASE
			WHEN created_at >= CURRENT_TIMESTAMP - INTERVAL '`+janelaCandidaturaNoCadastro+`' THEN 'vaga'
			ELSE 'banco_curriculos'
		END
		WHERE cpf = ? AND origem_cadastro IS NULL`,
		apenasDigitos(cpf),
	).Error
	if err != nil {
		return fmt.Errorf("erro ao registrar origem do cadastro: %w", err)
	}
	return nil
}
