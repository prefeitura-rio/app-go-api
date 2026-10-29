package empregabilidade

import "time"

// OrigemCadastro diz como o currículo chegou à base de currículos.
type OrigemCadastro string

const (
	// OrigemCadastroVaga é o currículo montado na hora de se candidatar a uma vaga.
	OrigemCadastroVaga OrigemCadastro = "vaga"
	// OrigemCadastroBancoCurriculos é o currículo preenchido direto no app, antes
	// de qualquer candidatura.
	OrigemCadastroBancoCurriculos OrigemCadastro = "banco_curriculos"
)

// ParticipacaoBanco é a escolha do cidadão de deixar o currículo visível para os
// recrutadores no banco de currículos (visível por padrão).
type ParticipacaoBanco struct {
	Disponivel     bool
	AlteradoEm     *time.Time
	OrigemCadastro *string
}

// ParticipacaoBancoStatus é o que o app precisa para montar o card do banco de
// currículos e o de "Meu currículo".
type ParticipacaoBancoStatus struct {
	// TemCurriculo é falso enquanto o cidadão não salvou nenhuma seção.
	TemCurriculo bool `json:"tem_curriculo"`
	// CurriculoCompleto diz se o currículo tem o mínimo exigido para se
	// candidatar; Pendencias diz o que falta. Não impede a participação no banco,
	// que mostra o que a pessoa conseguiu salvar.
	CurriculoCompleto bool     `json:"curriculo_completo"`
	Pendencias        []string `json:"pendencias"`
	// CurriculoAtualizadoEm é quando o cidadão salvou o currículo pela última vez.
	CurriculoAtualizadoEm *time.Time `json:"curriculo_atualizado_em"`
	Disponivel            bool       `json:"disponivel"`
	DisponivelAlteradoEm  *time.Time `json:"disponivel_alterado_em"`
	OrigemCadastro        *string    `json:"origem_cadastro"`
}

// AtualizarParticipacaoBancoRequest mostra ou oculta o currículo no banco.
type AtualizarParticipacaoBancoRequest struct {
	Disponivel *bool `json:"disponivel" binding:"required"`
}
