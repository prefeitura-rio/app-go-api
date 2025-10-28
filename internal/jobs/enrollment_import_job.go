package jobs

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type EnrollmentImportProcessor struct {
	db               *gorm.DB
	jobService       *services.JobService
	inscricaoService *services.InscricaoService
	cursoService     *services.CursoService
}

func NewEnrollmentImportProcessor(
	db *gorm.DB,
	jobService *services.JobService,
	inscricaoService *services.InscricaoService,
	cursoService *services.CursoService,
) *EnrollmentImportProcessor {
	return &EnrollmentImportProcessor{
		db:               db,
		jobService:       jobService,
		inscricaoService: inscricaoService,
		cursoService:     cursoService,
	}
}

type EnrollmentRow struct {
	NomeCompleto string
	CPF          string
	Idade        int
	Telefone     string
	Email        string
	Endereco     string
	Bairro       string
}

func (p *EnrollmentImportProcessor) Process(ctx context.Context, job *models.Job) error {
	var metadata models.EnrollmentImportMetadata
	if err := json.Unmarshal(job.Metadata, &metadata); err != nil {
		return fmt.Errorf("erro ao decodificar metadata: %w", err)
	}

	curso, err := p.cursoService.GetByID(ctx, metadata.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao buscar curso: %w", err)
	}
	if curso == nil {
		return fmt.Errorf("curso não encontrado")
	}

	var rows []EnrollmentRow
	var parseErr error

	if strings.HasSuffix(strings.ToLower(metadata.FileName), ".csv") {
		rows, parseErr = p.parseCSV(metadata.FileName)
	} else if strings.HasSuffix(strings.ToLower(metadata.FileName), ".xlsx") {
		rows, parseErr = p.parseXLSX(metadata.FileName)
	} else {
		return fmt.Errorf("formato de arquivo não suportado: %s", metadata.FileName)
	}

	if parseErr != nil {
		return fmt.Errorf("erro ao parsear arquivo: %w", parseErr)
	}

	totalRecords := len(rows)

	// Update job with total records
	job.TotalRecords = totalRecords
	if err := p.jobService.Update(ctx, job); err != nil {
		log.Printf("Erro ao atualizar total de registros: %v", err)
	}

	var enrollmentErrors []models.JobError
	var enrollmentIDs []uuid.UUID
	successCount := 0
	errorCount := 0
	duplicateCount := 0

	for i, row := range rows {
		select {
		case <-ctx.Done():
			return fmt.Errorf("job cancelado")
		default:
		}

		lineNumber := i + 2 // +2 because of header row and 1-indexed

		enrollmentID, err := p.processRow(ctx, metadata.CursoID, row)
		if err != nil {
			errorCount++
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "já inscrito") {
				duplicateCount++
			}
			enrollmentErrors = append(enrollmentErrors, models.JobError{
				Line:    lineNumber,
				Message: errorMsg,
				Data:    fmt.Sprintf("%s - %s", row.NomeCompleto, row.CPF),
			})
		} else {
			successCount++
			enrollmentIDs = append(enrollmentIDs, *enrollmentID)
		}

		// Update progress every 10 records or at the end
		if (i+1)%10 == 0 || i == len(rows)-1 {
			progress := ((i + 1) * 100) / totalRecords
			if err := p.jobService.UpdateProgress(ctx, job.ID, progress, successCount, errorCount); err != nil {
				log.Printf("Erro ao atualizar progresso: %v", err)
			}
		}
	}

	// Update job with final results
	result := models.EnrollmentImportResult{
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		DuplicateCount: duplicateCount,
		EnrollmentIDs:  enrollmentIDs,
	}

	resultJSON, _ := json.Marshal(result)
	job.Result = datatypes.JSON(resultJSON)

	if len(enrollmentErrors) > 0 {
		errorsJSON, _ := json.Marshal(enrollmentErrors)
		job.Errors = datatypes.JSON(errorsJSON)
	}

	job.SuccessCount = successCount
	job.ErrorCount = errorCount
	job.Progress = 100

	if err := p.jobService.Update(ctx, job); err != nil {
		log.Printf("Erro ao atualizar job com resultados: %v", err)
	}

	// Clean up file
	if err := os.Remove(metadata.FileName); err != nil {
		log.Printf("Erro ao remover arquivo temporário: %v", err)
	}

	return nil
}

func (p *EnrollmentImportProcessor) processRow(ctx context.Context, cursoID int, row EnrollmentRow) (*uuid.UUID, error) {
	if row.NomeCompleto == "" {
		return nil, fmt.Errorf("nome completo é obrigatório")
	}
	if row.CPF == "" {
		return nil, fmt.Errorf("CPF é obrigatório")
	}
	if row.Email == "" {
		return nil, fmt.Errorf("email é obrigatório")
	}

	// Clean CPF (remove dots, dashes)
	cpf := strings.ReplaceAll(row.CPF, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	cpf = strings.TrimSpace(cpf)

	if len(cpf) != 11 {
		return nil, fmt.Errorf("CPF inválido")
	}

	inscricao := &models.Inscricao{
		CursoID:      cursoID,
		CPF:          cpf,
		Name:         strings.TrimSpace(row.NomeCompleto),
		Email:        strings.TrimSpace(row.Email),
		Phone:        strings.TrimSpace(row.Telefone),
		Age:          row.Idade,
		Address:      strings.TrimSpace(row.Endereco),
		Neighborhood: strings.TrimSpace(row.Bairro),
	}

	if err := p.inscricaoService.Create(ctx, inscricao); err != nil {
		return nil, err
	}

	return &inscricao.ID, nil
}

func (p *EnrollmentImportProcessor) parseCSV(filePath string) ([]EnrollmentRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	// Validate header
	expectedHeaders := []string{"nome_completo", "cpf", "idade", "telefone", "email", "endereco", "bairro"}
	if !p.validateHeaders(header, expectedHeaders) {
		return nil, fmt.Errorf("cabeçalho inválido. Esperado: %v", expectedHeaders)
	}

	var rows []EnrollmentRow

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha: %w", err)
		}

		if len(record) < 7 {
			continue
		}

		idade, _ := strconv.Atoi(strings.TrimSpace(record[2]))

		row := EnrollmentRow{
			NomeCompleto: record[0],
			CPF:          record[1],
			Idade:        idade,
			Telefone:     record[3],
			Email:        record[4],
			Endereco:     record[5],
			Bairro:       record[6],
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (p *EnrollmentImportProcessor) parseXLSX(filePath string) ([]EnrollmentRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo XLSX: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("nenhuma planilha encontrada")
	}

	allRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler planilha: %w", err)
	}

	if len(allRows) < 2 {
		return nil, fmt.Errorf("arquivo vazio ou sem dados")
	}

	// Validate header
	header := allRows[0]
	expectedHeaders := []string{"nome_completo", "cpf", "idade", "telefone", "email", "endereco", "bairro"}
	if !p.validateHeaders(header, expectedHeaders) {
		return nil, fmt.Errorf("cabeçalho inválido. Esperado: %v", expectedHeaders)
	}

	var rows []EnrollmentRow

	for i := 1; i < len(allRows); i++ {
		record := allRows[i]
		if len(record) < 7 {
			continue
		}

		idade, _ := strconv.Atoi(strings.TrimSpace(record[2]))

		row := EnrollmentRow{
			NomeCompleto: record[0],
			CPF:          record[1],
			Idade:        idade,
			Telefone:     record[3],
			Email:        record[4],
			Endereco:     record[5],
			Bairro:       record[6],
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (p *EnrollmentImportProcessor) validateHeaders(actual, expected []string) bool {
	if len(actual) < len(expected) {
		return false
	}

	for i, exp := range expected {
		if i >= len(actual) {
			return false
		}
		if strings.ToLower(strings.TrimSpace(actual[i])) != exp {
			return false
		}
	}

	return true
}
