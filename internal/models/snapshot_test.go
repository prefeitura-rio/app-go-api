package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCitizenSnapshot_TableName(t *testing.T) {
	snapshot := &CitizenSnapshot{}
	assert.Equal(t, "citizen_snapshots", snapshot.TableName())
}

func TestCitizenSnapshot_ToPersonalInfo(t *testing.T) {
	t.Run("complete_snapshot", func(t *testing.T) {
		dataNascimento := time.Now()
		snapshot := &CitizenSnapshot{
			CPF:            "12345678901",
			Nome:           "João Silva",
			NomeSocial:     "João",
			Email:          "joao@example.com",
			Celular:        "21987654321",
			DataNascimento: &dataNascimento,
			Raca:           "Branca",
			Genero:         "Masculino",
			RendaFamiliar:  "3000",
			Escolaridade:   "Superior",
			Deficiencia:    "Nenhuma",
		}

		info := snapshot.ToPersonalInfo()

		assert.NotNil(t, info)
		assert.Equal(t, "João Silva", info.Nome)
		assert.Equal(t, "João", info.NomeSocial)
		assert.Equal(t, "joao@example.com", info.Email)
		assert.Equal(t, "21987654321", info.Celular)
		assert.Equal(t, dataNascimento, *info.DataNascimento)
		assert.Equal(t, "Branca", info.Raca)
		assert.Equal(t, "Masculino", info.Genero)
		assert.Equal(t, "3000", info.RendaFamiliar)
		assert.Equal(t, "Superior", info.Escolaridade)
		assert.Equal(t, "Nenhuma", info.Deficiencia)
	})

	t.Run("nil_snapshot", func(t *testing.T) {
		var snapshot *CitizenSnapshot
		info := snapshot.ToPersonalInfo()
		assert.Nil(t, info)
	})

	t.Run("empty_snapshot", func(t *testing.T) {
		snapshot := &CitizenSnapshot{}
		info := snapshot.ToPersonalInfo()
		assert.NotNil(t, info)
		assert.Equal(t, "", info.Nome)
		assert.Equal(t, "", info.Email)
	})
}

func TestOrgaoSnapshot_TableName(t *testing.T) {
	snapshot := &OrgaoSnapshot{}
	assert.Equal(t, "orgao_snapshots", snapshot.TableName())
}

func TestOrgaoMetadata_Scan(t *testing.T) {
	t.Run("valid_json", func(t *testing.T) {
		var metadata OrgaoMetadata
		err := metadata.Scan([]byte(`{"key1": "value1", "key2": 123}`))

		assert.NoError(t, err)
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, float64(123), metadata["key2"])
	})

	t.Run("nil_value", func(t *testing.T) {
		var metadata OrgaoMetadata
		err := metadata.Scan(nil)

		assert.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.Len(t, metadata, 0)
	})

	t.Run("empty_bytes", func(t *testing.T) {
		var metadata OrgaoMetadata
		err := metadata.Scan([]byte{})

		// Empty bytes are invalid JSON, should return error
		assert.Error(t, err)
	})

	t.Run("invalid_json", func(t *testing.T) {
		var metadata OrgaoMetadata
		err := metadata.Scan([]byte(`{invalid`))

		assert.Error(t, err)
	})

	t.Run("non_byte_slice", func(t *testing.T) {
		var metadata OrgaoMetadata
		err := metadata.Scan("not a byte slice")

		assert.NoError(t, err)
	})
}

func TestOrgaoMetadata_Value(t *testing.T) {
	t.Run("valid_metadata", func(t *testing.T) {
		metadata := OrgaoMetadata{
			"key1": "value1",
			"key2": 123,
		}

		value, err := metadata.Value()

		assert.NoError(t, err)
		assert.NotNil(t, value)
	})

	t.Run("nil_metadata", func(t *testing.T) {
		var metadata OrgaoMetadata
		value, err := metadata.Value()

		assert.NoError(t, err)
		assert.Nil(t, value)
	})
}

func TestInstituicaoEnsino_SetID(t *testing.T) {
	inst := &InstituicaoEnsino{}
	inst.SetID(42)
	assert.Equal(t, 42, inst.ID)
}

func TestCursoCategoria_TableName(t *testing.T) {
	rel := &CursoCategoria{}
	assert.Equal(t, "cursos_categorias", rel.TableName())
}

func TestCursoAcessibilidade_TableName(t *testing.T) {
	rel := &CursoAcessibilidade{}
	assert.Equal(t, "cursos_acessibilidades", rel.TableName())
}

func TestCitizenEndereco_Value(t *testing.T) {
	endereco := CitizenEndereco{
		Logradouro: "Rua das Flores",
		Numero:     "123",
		Bairro:     "Centro",
		Municipio:  "Rio de Janeiro",
		Estado:     "RJ",
		CEP:        "20000-000",
	}

	value, err := endereco.Value()

	assert.NoError(t, err)
	assert.NotNil(t, value)
}

func TestCitizenEndereco_Scan(t *testing.T) {
	t.Run("valid_json", func(t *testing.T) {
		var endereco CitizenEndereco
		err := endereco.Scan([]byte(`{
			"logradouro": "Rua das Flores",
			"numero": "123",
			"bairro": "Centro"
		}`))

		assert.NoError(t, err)
		assert.Equal(t, "Rua das Flores", endereco.Logradouro)
		assert.Equal(t, "123", endereco.Numero)
		assert.Equal(t, "Centro", endereco.Bairro)
	})

	t.Run("nil_value", func(t *testing.T) {
		var endereco CitizenEndereco
		err := endereco.Scan(nil)

		assert.NoError(t, err)
	})

	t.Run("non_byte_slice", func(t *testing.T) {
		var endereco CitizenEndereco
		err := endereco.Scan(123)

		assert.NoError(t, err)
	})
}
