package empregabilidade_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func meses(m int) *int { return &m }

func experiencia(comprovadaCT bool, tempoMeses *int) *empregabilidade.CurriculoExperiencia {
	return &empregabilidade.CurriculoExperiencia{ExperienciaComprovadaCT: comprovadaCT, TempoExperienciaMeses: tempoMeses}
}

func TestExperienciaCarteira_IsValid(t *testing.T) {
	assert.True(t, empregabilidade.ExperienciaCarteiraQualquerTempo.IsValid())
	assert.True(t, empregabilidade.ExperienciaCarteiraMinimo6Meses.IsValid())
	assert.Equal(t, empregabilidade.ExperienciaCarteira("qualquer_tempo"), empregabilidade.ExperienciaCarteiraQualquerTempo)
	assert.Equal(t, empregabilidade.ExperienciaCarteira("minimo_6_meses"), empregabilidade.ExperienciaCarteiraMinimo6Meses)
	assert.False(t, empregabilidade.ExperienciaCarteira("").IsValid())
	assert.False(t, empregabilidade.ExperienciaCarteira("minimo_12_meses").IsValid())
	assert.False(t, empregabilidade.ExperienciaCarteira("Qualquer_Tempo").IsValid())
}

func TestCriterioBloqueioExperienciaCarteira(t *testing.T) {
	assert.Equal(t, empregabilidade.CriterioBloqueio("experiencia_carteira"), empregabilidade.CriterioBloqueioExperienciaCarteira)
	assert.True(t, empregabilidade.CriterioBloqueioExperienciaCarteira.IsValid())

	got, err := empregabilidade.NormalizarCriteriosBloqueio([]empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioExperienciaCarteira,
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioExperienciaCarteira,
	})
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioExperienciaCarteira,
	}, got)
}

func TestExperienciaCarteira_Atende(t *testing.T) {
	cases := []struct {
		name         string
		exigencia    empregabilidade.ExperienciaCarteira
		experiencias []*empregabilidade.CurriculoExperiencia
		want         bool
	}{
		{name: "qualquer tempo sem experiências", exigencia: empregabilidade.ExperienciaCarteiraQualquerTempo, want: false},
		{
			name:         "qualquer tempo só com experiência fora da carteira",
			exigencia:    empregabilidade.ExperienciaCarteiraQualquerTempo,
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(false, meses(36))},
			want:         false,
		},
		{
			name:         "qualquer tempo com uma experiência em carteira de 1 mês",
			exigencia:    empregabilidade.ExperienciaCarteiraQualquerTempo,
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(false, meses(36)), experiencia(true, meses(1))},
			want:         true,
		},
		{
			name:         "qualquer tempo com experiência em carteira sem tempo informado",
			exigencia:    empregabilidade.ExperienciaCarteiraQualquerTempo,
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(true, nil)},
			want:         true,
		},
		{name: "6 meses sem experiências", exigencia: empregabilidade.ExperienciaCarteiraMinimo6Meses, want: false},
		{
			name:         "6 meses com 5 meses em carteira",
			exigencia:    empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(true, meses(5))},
			want:         false,
		},
		{
			name:         "6 meses com exatamente 6 meses numa experiência",
			exigencia:    empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(true, meses(6))},
			want:         true,
		},
		{
			name:      "6 meses somados entre experiências em carteira",
			exigencia: empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{
				experiencia(true, meses(2)), experiencia(true, meses(4)),
			},
			want: true,
		},
		{
			name:      "6 meses não soma experiência fora da carteira",
			exigencia: empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{
				experiencia(true, meses(3)), experiencia(false, meses(24)),
			},
			want: false,
		},
		{
			name:      "6 meses: experiência sem tempo informado conta zero",
			exigencia: empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{
				experiencia(true, nil), experiencia(true, meses(5)),
			},
			want: false,
		},
		{
			name:      "6 meses: tempo negativo não desconta",
			exigencia: empregabilidade.ExperienciaCarteiraMinimo6Meses,
			experiencias: []*empregabilidade.CurriculoExperiencia{
				experiencia(true, meses(-10)), experiencia(true, meses(6)),
			},
			want: true,
		},
		{
			name:         "experiência nil é ignorada",
			exigencia:    empregabilidade.ExperienciaCarteiraQualquerTempo,
			experiencias: []*empregabilidade.CurriculoExperiencia{nil},
			want:         false,
		},
		{
			name:         "exigência desconhecida nunca atende",
			exigencia:    empregabilidade.ExperienciaCarteira("x"),
			experiencias: []*empregabilidade.CurriculoExperiencia{experiencia(true, meses(120))},
			want:         false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.exigencia.Atende(tt.experiencias))
		})
	}
}

func TestExperienciaCarteira_Requisito(t *testing.T) {
	assert.Equal(t, "Experiência comprovada em carteira", empregabilidade.ExperienciaCarteiraQualquerTempo.Requisito())
	assert.Equal(t, "Experiência comprovada em carteira — mínimo de 6 meses", empregabilidade.ExperienciaCarteiraMinimo6Meses.Requisito())
}

func TestVaga_ExperienciaCarteira_JSON(t *testing.T) {
	t.Run("sem exigência é devolvido como null", func(t *testing.T) {
		v := &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Repositor"}
		require.NoError(t, v.AfterFind(nil))
		body, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"experiencia_carteira":null`)
	})
	t.Run("com exigência", func(t *testing.T) {
		e := empregabilidade.ExperienciaCarteiraMinimo6Meses
		v := &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Repositor", ExperienciaCarteira: &e}
		body, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"experiencia_carteira":"minimo_6_meses"`)
	})
	t.Run("leitura do corpo", func(t *testing.T) {
		var v empregabilidade.Vaga
		require.NoError(t, json.Unmarshal([]byte(`{"titulo":"x"}`), &v))
		assert.Nil(t, v.ExperienciaCarteira)

		require.NoError(t, json.Unmarshal([]byte(`{"experiencia_carteira":"qualquer_tempo"}`), &v))
		require.NotNil(t, v.ExperienciaCarteira)
		assert.Equal(t, empregabilidade.ExperienciaCarteiraQualquerTempo, *v.ExperienciaCarteira)
	})
}
