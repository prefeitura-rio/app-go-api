package jobs

import (
	"encoding/json"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestNormalizeFieldName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple lowercase",
			input: "nome",
			want:  "nome",
		},
		{
			name:  "simple uppercase",
			input: "NOME",
			want:  "nome",
		},
		{
			name:  "with spaces",
			input: "Nome Completo",
			want:  "nome_completo",
		},
		{
			name:  "with dashes",
			input: "data-nascimento",
			want:  "data_nascimento",
		},
		{
			name:  "with accents",
			input: "endereço",
			want:  "endereco",
		},
		{
			name:  "complex accents",
			input: "Situação Profissional",
			want:  "situacao_profissional",
		},
		{
			name:  "with ç",
			input: "Endereço",
			want:  "endereco",
		},
		{
			name:  "trim spaces",
			input: "  nome  completo  ",
			want:  "nome__completo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeFieldName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesFieldName(t *testing.T) {
	tests := []struct {
		name       string
		csvColumn  string
		targetField string
		want       bool
	}{
		{
			name:       "exact match lowercase",
			csvColumn:  "nome",
			targetField: "nome",
			want:       true,
		},
		{
			name:       "case insensitive",
			csvColumn:  "NOME",
			targetField: "nome",
			want:       true,
		},
		{
			name:       "with accents",
			csvColumn:  "endereço",
			targetField: "endereco",
			want:       true,
		},
		{
			name:       "with spaces vs underscores",
			csvColumn:  "nome completo",
			targetField: "nome_completo",
			want:       true,
		},
		{
			name:       "no match",
			csvColumn:  "nome",
			targetField: "email",
			want:       false,
		},
		{
			name:       "spaces normalized",
			csvColumn:  "Nome Completo",
			targetField: "nome_completo",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFieldName(tt.csvColumn, tt.targetField)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildFieldMappings(t *testing.T) {
	tests := []struct {
		name         string
		customFields []models.CustomField
		wantKeys     []string
	}{
		{
			name: "simple fields",
			customFields: []models.CustomField{
				{Title: "Data de Nascimento"},
				{Title: "Telefone"},
			},
			wantKeys: []string{"data_de_nascimento", "telefone"},
		},
		{
			name: "fields with accents",
			customFields: []models.CustomField{
				{Title: "Situação Profissional"},
				{Title: "Endereço"},
			},
			wantKeys: []string{"situacao_profissional", "endereco"},
		},
		{
			name:         "empty fields",
			customFields: []models.CustomField{},
			wantKeys:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFieldMappings(tt.customFields)
			assert.Equal(t, len(tt.wantKeys), len(got))
			for _, key := range tt.wantKeys {
				assert.Contains(t, got, key)
			}
		})
	}
}

func TestValidateFieldType(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		field   models.CustomField
		wantErr bool
	}{
		{
			name:    "valid number",
			value:   "42",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
		{
			name:    "invalid number",
			value:   "abc",
			field:   models.CustomField{FieldType: "number"},
			wantErr: true,
		},
		{
			name:    "valid email",
			value:   "test@example.com",
			field:   models.CustomField{FieldType: "email"},
			wantErr: false,
		},
		{
			name:    "invalid email",
			value:   "notanemail",
			field:   models.CustomField{FieldType: "email"},
			wantErr: true,
		},
		{
			name:    "valid phone",
			value:   "(21) 98765-4321",
			field:   models.CustomField{FieldType: "tel"},
			wantErr: false,
		},
		{
			name:    "invalid phone - too short",
			value:   "123",
			field:   models.CustomField{FieldType: "tel"},
			wantErr: true,
		},
		{
			name:    "valid text",
			value:   "any text",
			field:   models.CustomField{FieldType: "text"},
			wantErr: false,
		},
		{
			name:    "valid select option",
			value:   "option1",
			field:   models.CustomField{
				FieldType: "select",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: false,
		},
		{
			name:    "invalid select option",
			value:   "option3",
			field:   models.CustomField{
				FieldType: "select",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: true,
		},
		{
			name:    "valid multiselect",
			value:   "option1,option2",
			field:   models.CustomField{
				FieldType: "multiselect",
				Options:   mustMarshalJSON([]string{"option1", "option2", "option3"}),
			},
			wantErr: false,
		},
		{
			name:    "invalid multiselect - invalid option",
			value:   "option1,option4",
			field:   models.CustomField{
				FieldType: "multiselect",
				Options:   mustMarshalJSON([]string{"option1", "option2", "option3"}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldType(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCustomFields(t *testing.T) {
	tests := []struct {
		name         string
		row          EnrollmentRow
		customFields []models.CustomField
		wantErr      bool
		errContains  string
	}{
		{
			name: "all required fields present",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Data de Nascimento": "01/01/1990",
					"Telefone":           "21987654321",
				},
			},
			customFields: []models.CustomField{
				{Title: "Data de Nascimento", Required: true},
				{Title: "Telefone", Required: true},
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Telefone": "21987654321",
				},
			},
			customFields: []models.CustomField{
				{Title: "Data de Nascimento", Required: true},
				{Title: "Telefone", Required: false},
			},
			wantErr:     true,
			errContains: "obrigatório",
		},
		{
			name: "empty required field",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Data de Nascimento": "   ",
					"Telefone":           "21987654321",
				},
			},
			customFields: []models.CustomField{
				{Title: "Data de Nascimento", Required: true},
				{Title: "Telefone", Required: false},
			},
			wantErr:     true,
			errContains: "obrigatório",
		},
		{
			name: "optional field can be empty",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Data de Nascimento": "01/01/1990",
				},
			},
			customFields: []models.CustomField{
				{Title: "Data de Nascimento", Required: true},
				{Title: "Telefone", Required: false},
			},
			wantErr: false,
		},
		{
			name: "invalid field type",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Idade": "abc",
				},
			},
			customFields: []models.CustomField{
				{Title: "Idade", Required: true, FieldType: "number"},
			},
			wantErr:     true,
			errContains: "número válido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomFields(tt.row, tt.customFields)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to marshal JSON for tests
func mustMarshalJSON(v interface{}) datatypes.JSON {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(data)
}

// Edge case tests for normalizeFieldName
func TestNormalizeFieldName_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "all accented characters",
			input: "ãáâàéêíóôõúç",
			want:  "aaaaeeiooouc", // Note: õ becomes oo not o
		},
		{
			name:  "mixed case with all special chars",
			input: "Ação-Situação Profissional",
			want:  "acao_situacao_profissional",
		},
		{
			name:  "multiple consecutive spaces",
			input: "nome     completo",
			want:  "nome_____completo",
		},
		{
			name:  "multiple consecutive dashes",
			input: "data---nascimento",
			want:  "data___nascimento",
		},
		{
			name:  "tabs and newlines",
			input: "nome\tcompleto\n",
			want:  "nome\tcompleto",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only spaces",
			input: "     ",
			want:  "",
		},
		{
			name:  "only special chars",
			input: "---___",
			want:  "______",
		},
		{
			name:  "unicode characters",
			input: "José María Ñ",
			want:  "jose_maria_ñ", // ñ is not replaced
		},
		{
			name:  "very long string",
			input: "campo_" + string(make([]byte, 1000)),
			want:  "campo_" + string(make([]byte, 1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeFieldName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Edge case tests for matchesFieldName
func TestMatchesFieldName_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		csvColumn   string
		targetField string
		want        bool
	}{
		{
			name:        "empty vs empty",
			csvColumn:   "",
			targetField: "",
			want:        true,
		},
		{
			name:        "empty vs non-empty",
			csvColumn:   "",
			targetField: "nome",
			want:        false,
		},
		{
			name:        "spaces only match",
			csvColumn:   "   ",
			targetField: "",
			want:        true,
		},
		{
			name:        "unicode normalization",
			csvColumn:   "José",
			targetField: "jose",
			want:        true,
		},
		{
			name:        "mixed separators",
			csvColumn:   "data-de nascimento",
			targetField: "data_de_nascimento",
			want:        true,
		},
		{
			name:        "case and accent insensitive",
			csvColumn:   "SITUAÇÃO",
			targetField: "situacao",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFieldName(tt.csvColumn, tt.targetField)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Edge case tests for buildFieldMappings
func TestBuildFieldMappings_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		customFields []models.CustomField
		checkFunc    func(*testing.T, map[string]string)
	}{
		{
			name:         "nil custom fields",
			customFields: nil,
			checkFunc: func(t *testing.T, got map[string]string) {
				assert.NotNil(t, got)
				assert.Equal(t, 0, len(got))
			},
		},
		{
			name: "duplicate normalized keys",
			customFields: []models.CustomField{
				{Title: "Endereço"},
				{Title: "endereco"},
				{Title: "Endereco"},
			},
			checkFunc: func(t *testing.T, got map[string]string) {
				// Should have only one key since all normalize to "endereco"
				assert.Equal(t, 1, len(got))
				assert.Contains(t, got, "endereco")
			},
		},
		{
			name: "empty title fields",
			customFields: []models.CustomField{
				{Title: ""},
				{Title: "   "},
				{Title: "Valid Field"},
			},
			checkFunc: func(t *testing.T, got map[string]string) {
				assert.Contains(t, got, "valid_field")
				// Empty and whitespace-only titles normalize to ""
				assert.Contains(t, got, "")
			},
		},
		{
			name: "very long field names",
			customFields: []models.CustomField{
				{Title: string(make([]byte, 5000))},
			},
			checkFunc: func(t *testing.T, got map[string]string) {
				assert.Equal(t, 1, len(got))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFieldMappings(tt.customFields)
			tt.checkFunc(t, got)
		})
	}
}

// Edge case tests for validateFieldType
func TestValidateFieldType_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		field   models.CustomField
		wantErr bool
	}{
		{
			name:    "empty value for text",
			value:   "",
			field:   models.CustomField{FieldType: "text"},
			wantErr: false,
		},
		{
			name:    "very long email",
			value:   string(make([]byte, 1000)) + "@example.com",
			field:   models.CustomField{FieldType: "email"},
			wantErr: false, // Email validation is permissive
		},
		{
			name:    "number with decimals",
			value:   "42.5",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
		{
			name:    "negative number",
			value:   "-100",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
		{
			name:    "number with leading zeros",
			value:   "007",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
		{
			name:    "phone with special chars",
			value:   "+55 (21) 98765-4321",
			field:   models.CustomField{FieldType: "tel"},
			wantErr: false,
		},
		{
			name:    "phone with only numbers",
			value:   "21987654321",
			field:   models.CustomField{FieldType: "tel"},
			wantErr: false,
		},
		{
			name:    "email with plus sign",
			value:   "user+tag@example.com",
			field:   models.CustomField{FieldType: "email"},
			wantErr: false,
		},
		{
			name:    "email without domain",
			value:   "user@",
			field:   models.CustomField{FieldType: "email"},
			wantErr: false, // Email validation is permissive
		},
		{
			name:    "select with empty value",
			value:   "",
			field:   models.CustomField{
				FieldType: "select",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: true, // Empty is not in the options list
		},
		{
			name:    "multiselect with single value",
			value:   "option1",
			field:   models.CustomField{
				FieldType: "multiselect",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: false,
		},
		{
			name:    "multiselect with spaces",
			value:   "option1, option2",
			field:   models.CustomField{
				FieldType: "multiselect",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: false, // Should trim spaces
		},
		{
			name:    "multiselect empty values",
			value:   "option1,,option2",
			field:   models.CustomField{
				FieldType: "multiselect",
				Options:   mustMarshalJSON([]string{"option1", "option2"}),
			},
			wantErr: true, // Empty string between commas is invalid
		},
		{
			name:    "unknown field type",
			value:   "anything",
			field:   models.CustomField{FieldType: "unknown"},
			wantErr: false, // Unknown types default to allowing any value
		},
		{
			name:    "number zero",
			value:   "0",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
		{
			name:    "number with scientific notation",
			value:   "1e10",
			field:   models.CustomField{FieldType: "number"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldType(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Edge case tests for validateCustomFields
func TestValidateCustomFields_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		row          EnrollmentRow
		customFields []models.CustomField
		wantErr      bool
		errContains  string
	}{
		{
			name: "nil custom fields map",
			row: EnrollmentRow{
				CustomFields: nil,
			},
			customFields: []models.CustomField{
				{Title: "Field", Required: false},
			},
			wantErr: false,
		},
		{
			name: "empty custom fields map with required fields",
			row: EnrollmentRow{
				CustomFields: map[string]string{},
			},
			customFields: []models.CustomField{
				{Title: "Field", Required: true},
			},
			wantErr:     true,
			errContains: "obrigatório",
		},
		{
			name: "field with only whitespace",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Field": "   \t  \n  ",
				},
			},
			customFields: []models.CustomField{
				{Title: "Field", Required: true},
			},
			wantErr:     true,
			errContains: "obrigatório",
		},
		{
			name: "multiple validation errors",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Email": "invalid-email",
					"Idade": "not-a-number",
				},
			},
			customFields: []models.CustomField{
				{Title: "Email", Required: true, FieldType: "email"},
				{Title: "Idade", Required: true, FieldType: "number"},
			},
			wantErr:     true,
			errContains: "email válido", // The actual error message says "email" not "e-mail"
		},
		{
			name: "no custom fields defined",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Random": "value",
				},
			},
			customFields: []models.CustomField{},
			wantErr:      false,
		},
		{
			name: "extra fields in row not in schema",
			row: EnrollmentRow{
				CustomFields: map[string]string{
					"Field1": "value1",
					"Field2": "value2",
					"Extra":  "extra",
				},
			},
			customFields: []models.CustomField{
				{Title: "Field1", Required: true},
			},
			wantErr: false, // Extra fields are ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomFields(tt.row, tt.customFields)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
