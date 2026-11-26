package empregabilidade

import (
	"strings"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// situacoesProcurandoEmprego são as situações atuais em que o app exige o
// tempo procurando emprego.
var situacoesProcurandoEmprego = func() map[string]bool {
	m := make(map[string]bool, len(empregabilidade.SituacoesProcurandoEmprego))
	for _, s := range empregabilidade.SituacoesProcurandoEmprego {
		m[s] = true
	}
	return m
}()

// pendenciasCurriculoMinimo diz o que falta para o currículo ter o mínimo que o
// app exige para se candidatar: situação atual (com o tempo procurando emprego,
// para quem está sem trabalho) e termos de uso aceitos. A escolaridade, também
// exigida pelo app, vem do cadastro (RMI) e fica de fora para a regra não
// depender dele. Vale para a associação pelo banco, para entrar no banco e,
// em SQL, para a listagem do banco (condicaoCurriculoMinimo, no repositório).
func pendenciasCurriculoMinimo(curriculo *empregabilidade.CurriculoCompleto, termosAceitos bool) []string {
	var pendencias []string

	var situacao *empregabilidade.CurriculoSituacaoInteresses
	if curriculo != nil {
		situacao = curriculo.SituacaoInteresses
	}
	switch {
	case situacao == nil || situacao.IDSituacao == nil:
		pendencias = append(pendencias, "Situação atual não informada")
	case situacao.Situacao != nil && situacoesProcurandoEmprego[situacao.Situacao.Descricao] &&
		strings.TrimSpace(situacao.TempoProcurandoEmprego) == "":
		pendencias = append(pendencias, "Tempo procurando emprego não informado")
	}

	if !termosAceitos {
		pendencias = append(pendencias, "Termos de uso não aceitos")
	}
	return pendencias
}
