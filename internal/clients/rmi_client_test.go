package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestNormalizeCNPJ(t *testing.T) {
	t.Run("Valid CNPJ without formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12345678000195")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Valid CNPJ with standard formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12.345.678/0001-95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Valid CNPJ with partial formatting", func(t *testing.T) {
		result, err := NormalizeCNPJ("12345678/0001-95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("CNPJ with spaces", func(t *testing.T) {
		result, err := NormalizeCNPJ("12 345 678 0001 95")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != "12345678000195" {
			t.Errorf("Expected '12345678000195', got '%s'", result)
		}
	})

	t.Run("Invalid CNPJ - too short", func(t *testing.T) {
		_, err := NormalizeCNPJ("1234567800019")

		if err == nil {
			t.Error("Expected error for CNPJ with 13 digits")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - too long", func(t *testing.T) {
		_, err := NormalizeCNPJ("123456780001950")

		if err == nil {
			t.Error("Expected error for CNPJ with 15 digits")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - empty string", func(t *testing.T) {
		_, err := NormalizeCNPJ("")

		if err == nil {
			t.Error("Expected error for empty CNPJ")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - only letters", func(t *testing.T) {
		_, err := NormalizeCNPJ("abcdefghijklmn")

		if err == nil {
			t.Error("Expected error for CNPJ with only letters")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Invalid CNPJ - mixed letters and numbers", func(t *testing.T) {
		_, err := NormalizeCNPJ("12.345.678/ABCD-95")

		if err == nil {
			t.Error("Expected error for CNPJ with letters")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})
}

func TestNewRMIClient(t *testing.T) {
	t.Run("Creates client with correct timeout", func(t *testing.T) {
		baseURL := "http://test.example.com"
		timeout := 10 * time.Second
		client := NewRMIClient(baseURL, timeout)

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}
	})
}

func TestGetUserLegalEntities(t *testing.T) {
	t.Run("Success - single page", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/v1/citizen/") || !strings.Contains(r.URL.Path, "/legal-entities") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}

			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
			}

			response := models.LegalEntitiesResponse{
				Data: []models.LegalEntity{
					{
						CNPJ:         "12345678000195",
						RazaoSocial:  "Test Company",
						CNAEFiscal:   "1234-5/67",
						NomeFantasia: "Test Inc",
					},
				},
				Pagination: models.PaginationResponse{
					Page:       1,
					PerPage:    10,
					Total:      1,
					TotalPages: 1,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		entities, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(entities) != 1 {
			t.Errorf("Expected 1 entity, got %d", len(entities))
		}

		if entities[0].CNPJ != "12345678000195" {
			t.Errorf("Expected CNPJ '12345678000195', got '%s'", entities[0].CNPJ)
		}
	})

	t.Run("Success - multiple pages", func(t *testing.T) {
		pageCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pageCount++
			page := 1
			if strings.Contains(r.URL.RawQuery, "page=2") {
				page = 2
			}

			var response models.LegalEntitiesResponse
			if page == 1 {
				response = models.LegalEntitiesResponse{
					Data: []models.LegalEntity{
						{CNPJ: "11111111000111", RazaoSocial: "Company 1"},
					},
					Pagination: models.PaginationResponse{
						Page:       1,
						PerPage:    1,
						Total:      2,
						TotalPages: 2,
					},
				}
			} else {
				response = models.LegalEntitiesResponse{
					Data: []models.LegalEntity{
						{CNPJ: "22222222000122", RazaoSocial: "Company 2"},
					},
					Pagination: models.PaginationResponse{
						Page:       2,
						PerPage:    1,
						Total:      2,
						TotalPages: 2,
					},
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		entities, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expected 2 entities, got %d", len(entities))
		}

		if pageCount != 2 {
			t.Errorf("Expected 2 requests, got %d", pageCount)
		}
	})

	t.Run("Success - empty results", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := models.LegalEntitiesResponse{
				Data: []models.LegalEntity{},
				Pagination: models.PaginationResponse{
					Page:       1,
					PerPage:    10,
					Total:      0,
					TotalPages: 1,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		entities, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("Expected 0 entities, got %d", len(entities))
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}

		if !strings.Contains(err.Error(), "base URL not configured") {
			t.Errorf("Expected 'base URL not configured' error, got: %v", err)
		}
	})

	t.Run("Error - non-200 status", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for 500 status")
		}

		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("Expected status 500 error, got: %v", err)
		}
	})

	t.Run("Error - malformed JSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not valid json{"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}

		if !strings.Contains(err.Error(), "failed to decode") {
			t.Errorf("Expected decode error, got: %v", err)
		}
	})

	t.Run("Error - context cancellation", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})

	t.Run("Error - safety limit exceeded", func(t *testing.T) {
		requestCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			response := models.LegalEntitiesResponse{
				Data: make([]models.LegalEntity, 100),
				Pagination: models.PaginationResponse{
					Page:       requestCount,
					PerPage:    100,
					Total:      2000,
					TotalPages: 20,
				},
			}

			for i := 0; i < 100; i++ {
				response.Data[i] = models.LegalEntity{
					CNPJ:        "1234567800" + string(rune(i)),
					RazaoSocial: "Company " + string(rune(i)),
				}
			}

			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for exceeding safety limit")
		}

		if !strings.Contains(err.Error(), "exceeded maximum legal entities limit") {
			t.Errorf("Expected safety limit error, got: %v", err)
		}
	})

	t.Run("Error - invalid request creation", func(t *testing.T) {
		client := NewRMIClient("http://test.example.com", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for failed request")
		}
	})

	t.Run("Error - 404 on second page", func(t *testing.T) {
		callCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				response := models.LegalEntitiesResponse{
					Data: []models.LegalEntity{
						{CNPJ: "11111111000111", RazaoSocial: "Company 1"},
					},
					Pagination: models.PaginationResponse{
						Page:       1,
						PerPage:    1,
						Total:      2,
						TotalPages: 2,
					},
				}
				json.NewEncoder(w).Encode(response)
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Page not found"))
			}
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for 404 on second page")
		}

		if !strings.Contains(err.Error(), "status 404") {
			t.Errorf("Expected status 404 error, got: %v", err)
		}
	})
}

func TestGetOrgao(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/v1/departments/") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}

			orgao := models.Orgao{
				ID:      "123",
				SiglaUA: "SME",
				NomeUA:  "Secretaria Municipal de Educação",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(orgao)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		orgao, err := client.GetOrgao(ctx, "123")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if orgao.ID != "123" {
			t.Errorf("Expected ID '123', got '%s'", orgao.ID)
		}

		if orgao.SiglaUA != "SME" {
			t.Errorf("Expected SiglaUA 'SME', got '%s'", orgao.SiglaUA)
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetOrgao(ctx, "123")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}
	})

	t.Run("Error - 404 not found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetOrgao(ctx, "999")
		if err == nil {
			t.Error("Expected error for 404 status")
		}
	})

	t.Run("Error - malformed JSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetOrgao(ctx, "123")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})

	t.Run("Error - 500 internal server error", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal error"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetOrgao(ctx, "123")
		if err == nil {
			t.Error("Expected error for 500 status")
		}

		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("Expected status 500 error, got: %v", err)
		}
	})

	t.Run("Error - context cancellation", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.GetOrgao(ctx, "123")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

func TestGetCNPJOwnerInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
				legalEntity := models.LegalEntityDetails{
					CNPJ:        "12345678000195",
					RazaoSocial: "Test Company",
				}
				legalEntity.Socios = []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678900", NomeSocio: "John Doe"},
				}
				json.NewEncoder(w).Encode(legalEntity)
			} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
				citizen := models.CitizenContactInfo{
					CPF:  "12345678900",
					Nome: "John Doe",
				}
				citizen.Email.Indicador = true
				citizen.Email.Principal.Valor = "john@example.com"
				citizen.Telefone.Indicador = true
				citizen.Telefone.Principal.DDI = "55"
				citizen.Telefone.Principal.DDD = "21"
				citizen.Telefone.Principal.Valor = "999999999"
				json.NewEncoder(w).Encode(citizen)
			}
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		info, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12.345.678/0001-95")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if info.CNPJ != "12345678000195" {
			t.Errorf("Expected CNPJ '12345678000195', got '%s'", info.CNPJ)
		}

		if info.EmailPessoaFisica != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%s'", info.EmailPessoaFisica)
		}

		if info.CelularPessoaFisica != "5521999999999" {
			t.Errorf("Expected phone '5521999999999', got '%s'", info.CelularPessoaFisica)
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}
	})

	t.Run("Error - CNPJ not found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for not found")
		}
	})

	t.Run("Error - no socios found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
				Socios: []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{},
			}
			json.NewEncoder(w).Encode(legalEntity)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for no socios")
		}

		if !strings.Contains(err.Error(), "no socios found") {
			t.Errorf("Expected 'no socios found' error, got: %v", err)
		}
	})

	t.Run("Error - legal entity API error", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for API error")
		}

		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("Expected status 500 error, got: %v", err)
		}
	})

	t.Run("Error - malformed legal entity JSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})

	t.Run("Error - citizen not found", func(t *testing.T) {
		callCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				legalEntity := models.LegalEntityDetails{
					CNPJ:        "12345678000195",
					RazaoSocial: "Test Company",
				}
				legalEntity.Socios = []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678900", NomeSocio: "John Doe"},
				}
				json.NewEncoder(w).Encode(legalEntity)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for citizen not found")
		}

		if !strings.Contains(err.Error(), "citizen contact info not found") {
			t.Errorf("Expected 'citizen contact info not found' error, got: %v", err)
		}
	})

	t.Run("Error - citizen API error", func(t *testing.T) {
		callCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				legalEntity := models.LegalEntityDetails{
					CNPJ:        "12345678000195",
					RazaoSocial: "Test Company",
				}
				legalEntity.Socios = []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678900", NomeSocio: "John Doe"},
				}
				json.NewEncoder(w).Encode(legalEntity)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server error"))
			}
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for API error")
		}
	})

	t.Run("Error - malformed citizen JSON", func(t *testing.T) {
		callCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				legalEntity := models.LegalEntityDetails{
					CNPJ:        "12345678000195",
					RazaoSocial: "Test Company",
				}
				legalEntity.Socios = []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678900", NomeSocio: "John Doe"},
				}
				json.NewEncoder(w).Encode(legalEntity)
			} else {
				w.Write([]byte("invalid json"))
			}
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCNPJOwnerInfo(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})
}

func TestGetCitizenByCPF(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/v1/citizen/") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}

			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
				t.Errorf("Expected Bearer token in Authorization header")
			}

			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "Jane Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "jane@example.com"

			json.NewEncoder(w).Encode(citizen)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		citizen, err := client.GetCitizenByCPF(ctx, "test-token", "123.456.789-00")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if citizen.CPF != "12345678900" {
			t.Errorf("Expected CPF '12345678900', got '%s'", citizen.CPF)
		}

		if citizen.Nome != "Jane Doe" {
			t.Errorf("Expected Nome 'Jane Doe', got '%s'", citizen.Nome)
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCitizenByCPF(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}
	})

	t.Run("Error - 404 not found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCitizenByCPF(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for 404")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	t.Run("Error - 500 server error", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCitizenByCPF(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for 500")
		}
	})

	t.Run("Error - malformed JSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetCitizenByCPF(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})

	t.Run("Error - context deadline", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 50*time.Millisecond)
		ctx := context.Background()

		_, err := client.GetCitizenByCPF(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

func TestGetLegalEntityByCNPJ(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/v1/legal-entity/") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}

			razaoSocial := "Test Company Full"
			nomeFantasia := "Test Inc"
			cnae := "1234-5/67"

			entity := models.LegalEntityFull{
				CNPJ:            "12345678000195",
				RazaoSocial:     razaoSocial,
				NomeFantasia:    &nomeFantasia,
				CapitalSocial:   100000.50,
				CNAEFiscal:      &cnae,
				CNAESecundarias: []string{"2345-6/78"},
			}

			json.NewEncoder(w).Encode(entity)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		entity, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12.345.678/0001-95")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if entity.CNPJ != "12345678000195" {
			t.Errorf("Expected CNPJ '12345678000195', got '%s'", entity.CNPJ)
		}

		if entity.RazaoSocial != "Test Company Full" {
			t.Errorf("Expected RazaoSocial 'Test Company Full', got '%s'", entity.RazaoSocial)
		}

		if entity.NomeFantasia == nil || *entity.NomeFantasia != "Test Inc" {
			t.Error("Expected NomeFantasia to be 'Test Inc'")
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}
	})

	t.Run("Error - invalid CNPJ format", func(t *testing.T) {
		client := NewRMIClient("http://test.example.com", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "123")
		if err == nil {
			t.Error("Expected error for invalid CNPJ")
		}

		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Expected ErrInvalidCNPJ, got: %v", err)
		}
	})

	t.Run("Error - 404 not found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for 404")
		}

		if !errors.Is(err, ErrCNPJNotFound) {
			t.Errorf("Expected ErrCNPJNotFound, got: %v", err)
		}
	})

	t.Run("Error - 403 access denied", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for 403")
		}

		if !errors.Is(err, ErrCNPJAccessDenied) {
			t.Errorf("Expected ErrCNPJAccessDenied, got: %v", err)
		}
	})

	t.Run("Error - 400 bad request", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for 400")
		}

		if !strings.Contains(err.Error(), "status 400") {
			t.Errorf("Expected status 400 error, got: %v", err)
		}
	})

	t.Run("Error - 503 service unavailable", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for 503")
		}
	})

	t.Run("Error - malformed JSON response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{invalid json"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}

		if !strings.Contains(err.Error(), "failed to decode") {
			t.Errorf("Expected decode error, got: %v", err)
		}
	})

	t.Run("Error - timeout", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 50*time.Millisecond)
		ctx := context.Background()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected timeout error")
		}
	})

	t.Run("Error - context cancellation", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.GetLegalEntityByCNPJ(ctx, "test-token", "12345678000195")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

func TestGetCPFSecretarias(t *testing.T) {
	t.Run("Success - returns cd_uas", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/cpf-secretaria/") {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("Expected 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"cpf":    "12345678900",
				"cd_uas": []string{"UA001", "UA002"},
			})
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		cdUAs, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(cdUAs) != 2 {
			t.Errorf("Expected 2 cd_uas, got %d", len(cdUAs))
		}
		if cdUAs[0] != "UA001" {
			t.Errorf("Expected 'UA001', got '%s'", cdUAs[0])
		}
		if cdUAs[1] != "UA002" {
			t.Errorf("Expected 'UA002', got '%s'", cdUAs[1])
		}
	})

	t.Run("Success - empty cd_uas", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"cpf":    "12345678900",
				"cd_uas": []string{},
			})
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		cdUAs, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if len(cdUAs) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(cdUAs))
		}
	})

	t.Run("Success - nil cd_uas normalized to empty slice", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// omit cd_uas field — JSON decodes to nil, must be normalized
			json.NewEncoder(w).Encode(map[string]interface{}{"cpf": "12345678900"})
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		cdUAs, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if cdUAs == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(cdUAs) != 0 {
			t.Errorf("Expected 0 items, got %d", len(cdUAs))
		}
	})

	t.Run("Error - empty base URL", func(t *testing.T) {
		client := NewRMIClient("", 5*time.Second)
		_, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for empty base URL")
		}
		if !strings.Contains(err.Error(), "base URL not configured") {
			t.Errorf("Expected 'base URL not configured', got: %v", err)
		}
	})

	t.Run("Error - non-200 status", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		_, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for non-200 status")
		}
		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("Expected 'status 500', got: %v", err)
		}
	})

	t.Run("No mapping - 404 not found", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		cdUAs, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err != nil {
			t.Errorf("Expected no error for 404 (no mapping), got: %v", err)
		}
		if len(cdUAs) != 0 {
			t.Errorf("Expected empty slice for 404 (no mapping), got: %v", cdUAs)
		}
	})

	t.Run("Error - malformed JSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not valid json{"))
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		_, err := client.GetCPFSecretarias(context.Background(), "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
		if !strings.Contains(err.Error(), "failed to decode") {
			t.Errorf("Expected decode error, got: %v", err)
		}
	})

	t.Run("Error - context cancellation", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := client.GetCPFSecretarias(ctx, "test-token", "12345678900")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})

	t.Run("Success - Bearer token forwarded correctly", func(t *testing.T) {
		var receivedAuth string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"cpf":    "12345678900",
				"cd_uas": []string{"UA001"},
			})
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		_, err := client.GetCPFSecretarias(context.Background(), "my-service-token", "12345678900")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if receivedAuth != "Bearer my-service-token" {
			t.Errorf("Expected 'Bearer my-service-token', got '%s'", receivedAuth)
		}
	})

	t.Run("Success - CPF included in request path", func(t *testing.T) {
		var requestPath string
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"cpf":    "98765432100",
				"cd_uas": []string{"UA003"},
			})
		}))
		defer mockServer.Close()

		client := NewRMIClient(mockServer.URL, 5*time.Second)
		_, err := client.GetCPFSecretarias(context.Background(), "token", "98765432100")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !strings.Contains(requestPath, "98765432100") {
			t.Errorf("Expected CPF in path, got: %s", requestPath)
		}
	})
}

// Dados sintéticos: o corpo de erro simula um RMI que devolve dados do cidadão.
const syntheticRMIErrorBody = `{"error":"falha","cpf":"12345678900","nome":"Fulano de Teste"}`

func TestGetCPFSecretarias_ErrorStatus_OmitsResponseBody(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"500 sem novas tentativas", http.StatusInternalServerError},
		{"503 depois das novas tentativas", http.StatusServiceUnavailable},
		{"400 do RMI", http.StatusBadRequest},
		{"403 do RMI", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(syntheticRMIErrorBody))
			}))
			defer srv.Close()

			client := NewRMIClient(srv.URL, 5*time.Second,
				WithName(uniqueClientName(t, "secretaria-body-")),
				WithRetryBackoff(time.Millisecond, 2*time.Millisecond))
			_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
			if err == nil {
				t.Fatal("Expected error for non-200 status")
			}

			want := fmt.Sprintf("RMI API returned status %d", tt.status)
			if err.Error() != want {
				t.Errorf("Expected error %q, got %q", want, err.Error())
			}
			for _, leaked := range []string{"Fulano", "12345678900", "falha"} {
				if strings.Contains(err.Error(), leaked) {
					t.Errorf("Error message leaks response body content %q: %q", leaked, err.Error())
				}
			}
		})
	}
}

func TestGetOrgao_ErrorStatus_KeepsResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("orgao indisponivel"))
	}))
	defer srv.Close()

	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "orgao-body-")))
	_, err := client.GetOrgao(context.Background(), "ORG1")
	if err == nil {
		t.Fatal("Expected error for non-200 status")
	}
	if want := "RMI API returned status 500: orgao indisponivel"; err.Error() != want {
		t.Errorf("Expected error %q, got %q", want, err.Error())
	}
}

// Dados sintéticos: corpo de erro hostil, com CPF, CNPJ, e-mail e nome dentro e
// fora do campo "error".
const syntheticRMIErrorBodyWithPII = `{"error":"CPF 123.456.789-00 sem vinculo com 12.345.678/0001-90, contato fulano@example.com","cpf":"12345678900","nome":"Fulano de Teste","dados":{"email":"fulano@example.com"}}`

var syntheticPII = []string{
	"12345678900", "123.456.789-00", "12.345.678/0001-90", "12345678000190",
	"fulano@example.com", "Fulano de Teste", "Fulano",
}

func assertNoSyntheticPII(t *testing.T, err error) {
	t.Helper()
	for _, leaked := range syntheticPII {
		if strings.Contains(err.Error(), leaked) {
			t.Errorf("Error message leaks personal data %q: %q", leaked, err.Error())
		}
	}
}

// legalEntityWithSocioHandler answers the legal-entity step of GetCNPJOwnerInfo
// with one socio and the citizen step with status and body.
func legalEntityWithSocioHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{CNPJ: "12345678000195", RazaoSocial: "Empresa Teste"}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{{CPFSocio: "12345678900", NomeSocio: "Fulano de Teste"}}
			_ = json.NewEncoder(w).Encode(legalEntity)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestRMIClient_ErrorStatus_SanitizesResponseBody(t *testing.T) {
	ctx := context.Background()
	calls := []struct {
		name    string
		handler func(status int, body string) http.HandlerFunc
		call    func(c *RMIClient) error
	}{
		{"GetUserLegalEntities", nil, func(c *RMIClient) error {
			_, err := c.GetUserLegalEntities(ctx, "Bearer token", "12345678900")
			return err
		}},
		{"GetOrgao", nil, func(c *RMIClient) error {
			_, err := c.GetOrgao(ctx, "ORG1")
			return err
		}},
		{"GetCNPJOwnerInfo legal entity", nil, func(c *RMIClient) error {
			_, err := c.GetCNPJOwnerInfo(ctx, "token", "12345678000195")
			return err
		}},
		{"GetCNPJOwnerInfo citizen", legalEntityWithSocioHandler, func(c *RMIClient) error {
			_, err := c.GetCNPJOwnerInfo(ctx, "token", "12345678000195")
			return err
		}},
		{"GetCitizenByCPF", nil, func(c *RMIClient) error {
			_, err := c.GetCitizenByCPF(ctx, "token", "12345678900")
			return err
		}},
		{"GetLegalEntityByCNPJ", nil, func(c *RMIClient) error {
			_, err := c.GetLegalEntityByCNPJ(ctx, "token", "12345678000195")
			return err
		}},
	}

	bodies := []struct {
		name string
		body string
		want string
	}{
		{"campo error com dados pessoais", syntheticRMIErrorBodyWithPII,
			`RMI API returned status 500: {"error":"CPF [omitido] sem vinculo com [omitido], contato [omitido]"}`},
		{"JSON sem campo de mensagem", `{"cpf":"12345678900","nome":"Fulano de Teste"}`,
			"RMI API returned status 500"},
		{"texto puro com dados pessoais", "falha no upstream\n12345678900 fulano@example.com",
			"RMI API returned status 500: falha no upstream [omitido] [omitido]"},
	}

	for _, tc := range calls {
		for _, b := range bodies {
			t.Run(tc.name+"/"+b.name, func(t *testing.T) {
				handler := tc.handler
				if handler == nil {
					handler = func(status int, body string) http.HandlerFunc {
						return func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(status)
							_, _ = w.Write([]byte(body))
						}
					}
				}
				srv := httptest.NewServer(handler(http.StatusInternalServerError, b.body))
				defer srv.Close()

				client := NewRMIClient(srv.URL, 5*time.Second,
					WithName(uniqueClientName(t, "sanitize-body-")),
					WithRetryBackoff(time.Millisecond, 2*time.Millisecond))
				err := tc.call(client)
				if err == nil {
					t.Fatal("Expected error for non-200 status")
				}
				if !strings.HasSuffix(err.Error(), b.want) {
					t.Errorf("Expected error ending with %q, got %q", b.want, err.Error())
				}
				assertNoSyntheticPII(t, err)
				if strings.Contains(err.Error(), "\n") {
					t.Errorf("Error message must be a single line, got %q", err.Error())
				}
			})
		}
	}
}

func TestRMIClient_ErrorStatus_KeepsMessagesWithoutPersonalData(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		status int
		body   string
		call   func(c *RMIClient) error
		want   string
	}{
		{"GetOrgao 500 do sync de órgãos", http.StatusInternalServerError, `{"error":"Failed to retrieve department"}`,
			func(c *RMIClient) error { _, err := c.GetOrgao(ctx, "ORG1"); return err },
			`RMI API returned status 500: {"error":"Failed to retrieve department"}`},
		{"GetCitizenByCPF 400 do sync de cidadãos", http.StatusBadRequest, `{"error":"Invalid CPF format"}`,
			func(c *RMIClient) error { _, err := c.GetCitizenByCPF(ctx, "token", "123"); return err },
			`RMI API returned status 400: {"error":"Invalid CPF format"}`},
		{"GetLegalEntityByCNPJ 400 com message", http.StatusBadRequest, `{"message":"invalid request","error":"bad request"}` + "\n",
			func(c *RMIClient) error { _, err := c.GetLegalEntityByCNPJ(ctx, "token", "12345678000195"); return err },
			`RMI API returned status 400: {"error":"bad request","message":"invalid request"}`},
		{"GetOrgao corpo vazio", http.StatusInternalServerError, "",
			func(c *RMIClient) error { _, err := c.GetOrgao(ctx, "ORG1"); return err },
			"RMI API returned status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "keep-body-")))
			err := tt.call(client)
			if err == nil {
				t.Fatal("Expected error for non-200 status")
			}
			if err.Error() != tt.want {
				t.Errorf("Expected error %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestSanitizeRMIErrorBody(t *testing.T) {
	long := strings.Repeat("a", 500)
	tests := []struct {
		name string
		body string
		want string
	}{
		{"vazio", "", ""},
		{"só espaços", "  \n\t ", ""},
		{"JSON error sem dado pessoal", `{"error":"Failed to retrieve department"}`, `{"error":"Failed to retrieve department"}`},
		{"JSON error com CPF formatado", `{"error":"CPF 123.456.789-00 inválido"}`, `{"error":"CPF [omitido] inválido"}`},
		{"JSON error com telefone", `{"error":"telefone 21987654321 inválido"}`, `{"error":"telefone [omitido] inválido"}`},
		{"JSON preserva números curtos", `{"error":"timeout after 30s, code 504"}`, `{"error":"timeout after 30s, code 504"}`},
		{"JSON descarta campos fora de error/message", `{"error":"x","cpf":"12345678900","nome":"Fulano"}`, `{"error":"x"}`},
		{"JSON sem campo de mensagem", `{"cpf":"12345678900","nome":"Fulano"}`, ""},
		{"JSON error não string", `{"error":{"nome":"Fulano"}}`, ""},
		{"JSON array", `[{"nome":"Fulano","cpf":"12345678900"}]`, ""},
		{"JSON string", `"Fulano 12345678900"`, ""},
		{"JSON error vazio", `{"error":"  "}`, ""},
		{"JSON não escapa HTML", `{"error":"a < b & c"}`, `{"error":"a < b & c"}`},
		{"JSON error longo é truncado", `{"error":"` + long + `"}`, `{"error":"` + strings.Repeat("a", rmiErrorSnippetMaxRunes) + `…"}`},
		{"texto puro sem dado pessoal", "orgao indisponivel", "orgao indisponivel"},
		{"texto puro do proxy", "upstream connect error or disconnect/reset before headers. reset reason: connection failure",
			"upstream connect error or disconnect/reset before headers. reset reason: connection failure"},
		{"texto puro com CNPJ e e-mail", "CNPJ 12.345.678/0001-90 de fulano.teste+x@example.com.br",
			"CNPJ [omitido] de [omitido]"},
		{"texto puro em várias linhas vira uma", "linha 1\r\nlinha 2\n\nlinha 3", "linha 1 linha 2 linha 3"},
		{"texto puro longo é truncado", long, strings.Repeat("a", rmiErrorSnippetMaxRunes) + "…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeRMIErrorBody([]byte(tt.body)); got != tt.want {
				t.Errorf("sanitizeRMIErrorBody(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}

func TestRMIClient_ErrorStatus_LargeBodyIsBounded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("x", 1<<20)))
	}))
	defer srv.Close()

	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "large-body-")))
	_, err := client.GetOrgao(context.Background(), "ORG1")
	if err == nil {
		t.Fatal("Expected error for non-200 status")
	}
	if len(err.Error()) > 512 {
		t.Errorf("Error message should be bounded, got %d bytes", len(err.Error()))
	}
}
