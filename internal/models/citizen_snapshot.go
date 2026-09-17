package models

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"time"
)

// CitizenSnapshot represents cached citizen data from RMI API
type CitizenSnapshot struct {
	CPF            string           `json:"cpf" gorm:"type:varchar(11);primaryKey"`
	Nome           string           `json:"nome" gorm:"type:varchar(500)"`
	NomeSocial     string           `json:"nome_social" gorm:"type:varchar(500);column:nome_social"`
	Email          string           `json:"email" gorm:"type:varchar(500)"`
	Celular        string           `json:"celular" gorm:"type:varchar(50)"`
	DataNascimento *time.Time       `json:"data_nascimento" gorm:"type:date;column:data_nascimento"`
	Endereco       *CitizenEndereco `json:"endereco" gorm:"type:jsonb"`
	Raca           string           `json:"raca" gorm:"type:varchar(100)"`
	Genero         string           `json:"genero" gorm:"type:varchar(100)"`
	RendaFamiliar  string           `json:"renda_familiar" gorm:"type:varchar(100);column:renda_familiar"`
	Escolaridade   string           `json:"escolaridade" gorm:"type:varchar(100)"`
	Deficiencia    string           `json:"deficiencia" gorm:"type:varchar(500)"`
	LastSyncedAt   time.Time        `json:"last_synced_at" gorm:"not null;column:last_synced_at"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// CitizenEndereco represents the address structure from RMI
type CitizenEndereco struct {
	Logradouro     string `json:"logradouro"`
	TipoLogradouro string `json:"tipo_logradouro,omitempty"`
	Numero         string `json:"numero"`
	Complemento    string `json:"complemento,omitempty"`
	Bairro         string `json:"bairro"`
	Municipio      string `json:"municipio"`
	Estado         string `json:"estado"`
	CEP            string `json:"cep"`
}

// TableName specifies the table name for GORM
func (CitizenSnapshot) TableName() string {
	return "citizen_snapshots"
}

// Value implements driver.Valuer for CitizenEndereco (JSONB storage)
func (e CitizenEndereco) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements sql.Scanner for CitizenEndereco (JSONB reading)
func (e *CitizenEndereco) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, e)
}

// CitizenPersonalInfo represents the personal info returned in enrollment responses
type CitizenPersonalInfo struct {
	Nome           string           `json:"nome,omitempty"`
	NomeSocial     string           `json:"nome_social,omitempty"`
	Email          string           `json:"email"`
	Celular        string           `json:"celular"`
	DataNascimento *time.Time       `json:"data_nascimento,omitempty"`
	Endereco       *CitizenEndereco `json:"endereco,omitempty"`
	Raca           string           `json:"raca,omitempty"`
	Genero         string           `json:"genero,omitempty"`
	RendaFamiliar  string           `json:"renda_familiar,omitempty"`
	Escolaridade   string           `json:"escolaridade,omitempty"`
	Deficiencia    string           `json:"deficiencia,omitempty"`
	DataHash       string           `json:"data_hash,omitempty"`
}

// ComputeDataHash generates a deterministic SHA-256 hash of the snapshot's contact/profile data.
// This allows detecting when citizen data has drifted or been updated in RMI.
func (c *CitizenSnapshot) ComputeDataHash() string {
	if c == nil {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(c.Nome))
	h.Write([]byte(c.NomeSocial))
	h.Write([]byte(c.Email))
	h.Write([]byte(c.Celular))
	if c.DataNascimento != nil {
		h.Write([]byte(c.DataNascimento.Format("2006-01-02")))
	}
	if c.Endereco != nil {
		h.Write([]byte(c.Endereco.Logradouro))
		h.Write([]byte(c.Endereco.TipoLogradouro))
		h.Write([]byte(c.Endereco.Numero))
		h.Write([]byte(c.Endereco.Complemento))
		h.Write([]byte(c.Endereco.Bairro))
		h.Write([]byte(c.Endereco.Municipio))
		h.Write([]byte(c.Endereco.Estado))
		h.Write([]byte(c.Endereco.CEP))
	}
	h.Write([]byte(c.Raca))
	h.Write([]byte(c.Genero))
	h.Write([]byte(c.RendaFamiliar))
	h.Write([]byte(c.Escolaridade))
	h.Write([]byte(c.Deficiencia))
	return hex.EncodeToString(h.Sum(nil))
}

// ToPersonalInfo converts a CitizenSnapshot to CitizenPersonalInfo for API responses
func (c *CitizenSnapshot) ToPersonalInfo() *CitizenPersonalInfo {
	if c == nil {
		return nil
	}
	return &CitizenPersonalInfo{
		Nome:           c.Nome,
		NomeSocial:     c.NomeSocial,
		Email:          c.Email,
		Celular:        c.Celular,
		DataNascimento: c.DataNascimento,
		Endereco:       c.Endereco,
		Raca:           c.Raca,
		Genero:         c.Genero,
		RendaFamiliar:  c.RendaFamiliar,
		Escolaridade:   c.Escolaridade,
		Deficiencia:    c.Deficiencia,
		DataHash:       c.ComputeDataHash(),
	}
}
