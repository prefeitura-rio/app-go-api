package models_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestCategoria_TableName(t *testing.T) {
	cat := &models.Categoria{}
	if got := cat.TableName(); got != "categorias" {
		t.Errorf("Categoria.TableName() = %v, want categorias", got)
	}
}

func TestCategoria_SetID(t *testing.T) {
	cat := &models.Categoria{}
	cat.SetID(42)
	if cat.ID != 42 {
		t.Errorf("SetID(42): ID = %d, want 42", cat.ID)
	}
}

func TestAcessibilidade_TableName(t *testing.T) {
	aces := &models.Acessibilidade{}
	if got := aces.TableName(); got != "acessibilidades" {
		t.Errorf("Acessibilidade.TableName() = %v, want acessibilidades", got)
	}
}

func TestAcessibilidade_SetID(t *testing.T) {
	aces := &models.Acessibilidade{}
	aces.SetID(10)
	if aces.ID != 10 {
		t.Errorf("SetID(10): ID = %d, want 10", aces.ID)
	}
}

func TestEscolaridade_TableName(t *testing.T) {
	esc := &models.Escolaridade{}
	if got := esc.TableName(); got != "escolaridades" {
		t.Errorf("Escolaridade.TableName() = %v, want escolaridades", got)
	}
}

func TestEscolaridade_SetID(t *testing.T) {
	esc := &models.Escolaridade{}
	esc.SetID(5)
	if esc.ID != 5 {
		t.Errorf("SetID(5): ID = %d, want 5", esc.ID)
	}
}

func TestEmpresa_TableName(t *testing.T) {
	emp := &models.Empresa{}
	if got := emp.TableName(); got != "empresas" {
		t.Errorf("Empresa.TableName() = %v, want empresas", got)
	}
}

func TestEmpresa_SetID(t *testing.T) {
	emp := &models.Empresa{}
	emp.SetID(15)
	if emp.ID != 15 {
		t.Errorf("SetID(15): ID = %d, want 15", emp.ID)
	}
}
