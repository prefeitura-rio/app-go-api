package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// Test helper functions and conversion logic

func TestCitizenInfoToSnapshot_AllFields(t *testing.T) {
	// Create a worker for testing the conversion method
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:           "12345678901",
		Nome:          "João Silva",
		NomeSocial:    "João",
		Raca:          "Branca",
		Genero:        "Masculino",
		RendaFamiliar: "1-2 salários",
		Escolaridade:  "Superior Completo",
		Deficiencia:   "Nenhuma",
	}

	// Set email and telefone values
	citizenInfo.Email.Indicador = true
	citizenInfo.Email.Principal.Valor = "joao@example.com"

	citizenInfo.Telefone.Indicador = true
	citizenInfo.Telefone.Principal.DDI = "55"
	citizenInfo.Telefone.Principal.DDD = "21"
	citizenInfo.Telefone.Principal.Valor = "987654321"

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.Equal(t, "12345678901", snapshot.CPF)
	assert.Equal(t, "João Silva", snapshot.Nome)
	assert.Equal(t, "João", snapshot.NomeSocial)
	assert.Equal(t, "joao@example.com", snapshot.Email)
	assert.Equal(t, "5521987654321", snapshot.Celular)
	assert.Equal(t, "Branca", snapshot.Raca)
	assert.Equal(t, "Masculino", snapshot.Genero)
	assert.Equal(t, "1-2 salários", snapshot.RendaFamiliar)
	assert.Equal(t, "Superior Completo", snapshot.Escolaridade)
	assert.Equal(t, "Nenhuma", snapshot.Deficiencia)
}

func TestCitizenInfoToSnapshot_WithAddress(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:  "12345678901",
		Nome: "Maria Santos",
		Endereco: &models.CitizenEnderecoInfo{
			Indicador: true,
			Principal: struct {
				Logradouro     string `json:"logradouro"`
				TipoLogradouro string `json:"tipo_logradouro"`
				Numero         string `json:"numero"`
				Complemento    string `json:"complemento"`
				Bairro         string `json:"bairro"`
				Municipio      string `json:"municipio"`
				Estado         string `json:"estado"`
				CEP            string `json:"cep"`
				Origem         string `json:"origem"`
				Sistema        string `json:"sistema"`
				UpdatedAt      string `json:"updated_at"`
			}{
				Logradouro:     "Avenida Atlântica",
				TipoLogradouro: "Avenida",
				Numero:         "1500",
				Complemento:    "Bloco A",
				Bairro:         "Copacabana",
				Municipio:      "Rio de Janeiro",
				Estado:         "RJ",
				CEP:            "22021-000",
			},
		},
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.NotNil(t, snapshot.Endereco)
	assert.Equal(t, "Avenida Atlântica", snapshot.Endereco.Logradouro)
	assert.Equal(t, "Avenida", snapshot.Endereco.TipoLogradouro)
	assert.Equal(t, "1500", snapshot.Endereco.Numero)
	assert.Equal(t, "Bloco A", snapshot.Endereco.Complemento)
	assert.Equal(t, "Copacabana", snapshot.Endereco.Bairro)
	assert.Equal(t, "Rio de Janeiro", snapshot.Endereco.Municipio)
	assert.Equal(t, "RJ", snapshot.Endereco.Estado)
	assert.Equal(t, "22021-000", snapshot.Endereco.CEP)
}

func TestCitizenInfoToSnapshot_WithBirthDate(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:  "12345678901",
		Nome: "Pedro Costa",
		Nascimento: &models.CitizenNascimento{
			Data: "1990-05-15",
		},
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.NotNil(t, snapshot.DataNascimento)
	assert.Equal(t, 1990, snapshot.DataNascimento.Year())
	assert.Equal(t, time.May, snapshot.DataNascimento.Month())
	assert.Equal(t, 15, snapshot.DataNascimento.Day())
}

func TestCitizenInfoToSnapshot_WithBrazilianDateFormat(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:  "12345678901",
		Nome: "Ana Souza",
		Nascimento: &models.CitizenNascimento{
			Data: "15/05/1990",
		},
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.NotNil(t, snapshot.DataNascimento)
	assert.Equal(t, 1990, snapshot.DataNascimento.Year())
	assert.Equal(t, time.May, snapshot.DataNascimento.Month())
	assert.Equal(t, 15, snapshot.DataNascimento.Day())
}

func TestCitizenInfoToSnapshot_InvalidDate(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:  "12345678901",
		Nome: "Carlos Lima",
		Nascimento: &models.CitizenNascimento{
			Data: "invalid-date",
		},
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	// Should handle invalid date gracefully - parseDate returns zero time on invalid input
	// which is not nil but represents an invalid date
	if snapshot.DataNascimento != nil {
		assert.True(t, snapshot.DataNascimento.IsZero() || snapshot.DataNascimento.Year() == 1)
	}
}

func TestCitizenInfoToSnapshot_NoAddress(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:      "12345678901",
		Nome:     "Test User",
		Endereco: nil,
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.Nil(t, snapshot.Endereco)
}

func TestCitizenInfoToSnapshot_AddressNotIndicator(t *testing.T) {
	worker := &CitizenSyncWorker{}

	citizenInfo := &models.CitizenContactInfo{
		CPF:  "12345678901",
		Nome: "Test User",
		Endereco: &models.CitizenEnderecoInfo{
			Indicador: false, // Not an active address
		},
	}

	snapshot := worker.citizenInfoToSnapshot("12345678901", citizenInfo)

	assert.Nil(t, snapshot.Endereco)
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "ISO 8601 date",
			dateStr: "2006-01-02",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "Brazilian date format",
			dateStr: "02/01/2006",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "ISO 8601 datetime with Z",
			dateStr: "2006-01-02T15:04:05Z",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "ISO 8601 datetime with timezone",
			dateStr: "2006-01-02T15:04:05-03:00",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "invalid date format",
			dateStr: "invalid",
			wantErr: false,
			wantNil: true,
		},
		{
			name:    "empty string",
			dateStr: "",
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.dateStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
			}
		})
	}
}

func TestParseDateSpecificValues(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected time.Time
	}{
		{
			name:     "ISO date",
			dateStr:  "2023-12-25",
			expected: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Brazilian date",
			dateStr:  "25/12/2023",
			expected: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.dateStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Year(), got.Year())
			assert.Equal(t, tt.expected.Month(), got.Month())
			assert.Equal(t, tt.expected.Day(), got.Day())
		})
	}
}

func TestParseDateTimezoneHandling(t *testing.T) {
	dateWithTimezone := "2023-06-15T14:30:00-03:00"
	result, err := parseDate(dateWithTimezone)

	assert.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.June, result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 14, result.Hour())
	assert.Equal(t, 30, result.Minute())
}

func TestParseDateUTCTimezone(t *testing.T) {
	dateUTC := "2023-06-15T14:30:00Z"
	result, err := parseDate(dateUTC)

	assert.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.June, result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 14, result.Hour())
	assert.Equal(t, 30, result.Minute())
}

func TestMaskCPF(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want string
	}{
		{
			name: "full CPF without formatting",
			cpf:  "12345678901",
			want: "123******01",
		},
		{
			name: "full CPF with dots",
			cpf:  "123.456.789-01",
			want: "123******01",
		},
		{
			name: "full CPF with dots only",
			cpf:  "123.456.78901",
			want: "123******01",
		},
		{
			name: "short CPF - too short to mask properly",
			cpf:  "1234",
			want: "***",
		},
		{
			name: "very short CPF",
			cpf:  "12",
			want: "***",
		},
		{
			name: "empty CPF",
			cpf:  "",
			want: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskCPF(tt.cpf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskCPFFormat(t *testing.T) {
	cpf := "12345678901"
	masked := maskCPF(cpf)

	assert.Equal(t, 11, len(masked), "Masked CPF should have 11 characters")
	assert.Equal(t, "123", masked[:3], "Should show first 3 digits")
	assert.Equal(t, "01", masked[len(masked)-2:], "Should show last 2 digits")
	assert.Contains(t, masked, "******", "Should contain 6 asterisks")
}

func TestMaskCPFRemovesAllFormatting(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want string
	}{
		{
			name: "CPF with dots and dash",
			cpf:  "123.456.789-01",
			want: "123******01",
		},
		{
			name: "CPF with spaces",
			cpf:  "123 456 789 01",
			want: "123******01",
		},
		{
			name: "CPF with mixed formatting",
			cpf:  "123.456.789 01",
			want: "123******01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskCPF(tt.cpf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaskCPFMinimumLength(t *testing.T) {
	// Test minimum length for proper masking
	cpf5 := "12345"
	masked5 := maskCPF(cpf5)
	assert.Equal(t, "123******45", masked5)

	// Test exactly 11 digits
	cpf11 := "00000000000"
	masked11 := maskCPF(cpf11)
	assert.Equal(t, "000******00", masked11)
}

func TestMaskCPFIdempotent(t *testing.T) {
	cpf := "12345678901"
	masked1 := maskCPF(cpf)
	masked2 := maskCPF(cpf)

	assert.Equal(t, masked1, masked2, "Multiple calls should produce same result")
}
