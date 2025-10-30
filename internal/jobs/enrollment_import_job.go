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
	Turma        string
	CustomFields map[string]string
}

func normalizeFieldName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "ã", "a")
	name = strings.ReplaceAll(name, "á", "a")
	name = strings.ReplaceAll(name, "â", "a")
	name = strings.ReplaceAll(name, "à", "a")
	name = strings.ReplaceAll(name, "é", "e")
	name = strings.ReplaceAll(name, "ê", "e")
	name = strings.ReplaceAll(name, "í", "i")
	name = strings.ReplaceAll(name, "ó", "o")
	name = strings.ReplaceAll(name, "ô", "o")
	name = strings.ReplaceAll(name, "õ", "o")
	name = strings.ReplaceAll(name, "ú", "u")
	name = strings.ReplaceAll(name, "ç", "c")
	return name
}

func matchesFieldName(csvColumn, targetField string) bool {
	return normalizeFieldName(csvColumn) == normalizeFieldName(targetField)
}

func buildFieldMappings(customFields []models.CustomField) map[string]string {
	mappings := make(map[string]string)
	for _, field := range customFields {
		normalizedKey := normalizeFieldName(field.Title)
		mappings[normalizedKey] = field.Title
	}
	return mappings
}

func buildLocationMap(locations []models.LocationClass) map[string]uuid.UUID {
	locationMap := make(map[string]uuid.UUID)
	for _, loc := range locations {
		normalizedAddr := strings.ToLower(strings.TrimSpace(loc.Address))
		locationMap[normalizedAddr] = loc.ID
	}
	return locationMap
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

	var customFields []models.CustomField
	if err := p.db.Where("curso_id = ?", metadata.CursoID).Find(&customFields).Error; err != nil {
		log.Printf("Aviso: erro ao carregar custom_fields: %v", err)
	}

	var locationClasses []models.LocationClass
	if err := p.db.Where("curso_id = ?", metadata.CursoID).Find(&locationClasses).Error; err != nil {
		log.Printf("Aviso: erro ao carregar location_classes: %v", err)
	}

	fieldMappings := buildFieldMappings(customFields)
	locationMap := buildLocationMap(locationClasses)

	var rows []EnrollmentRow
	var parseErr error

	if strings.HasSuffix(strings.ToLower(metadata.FileName), ".csv") {
		rows, parseErr = p.parseCSV(metadata.FileName, fieldMappings)
	} else if strings.HasSuffix(strings.ToLower(metadata.FileName), ".xlsx") {
		rows, parseErr = p.parseXLSX(metadata.FileName, fieldMappings)
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

		enrollmentID, err := p.processRow(ctx, metadata.CursoID, row, customFields, locationMap, locationClasses)
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

func validateCustomFields(row EnrollmentRow, customFields []models.CustomField) error {
	for _, field := range customFields {
		value, exists := row.CustomFields[field.Title]

		if field.Required && (!exists || strings.TrimSpace(value) == "") {
			return fmt.Errorf("campo obrigatório '%s' não fornecido ou vazio", field.Title)
		}

		if !exists || value == "" {
			continue
		}

		if err := validateFieldType(value, field); err != nil {
			return fmt.Errorf("campo '%s': %w", field.Title, err)
		}
	}

	return nil
}

func validateFieldType(value string, field models.CustomField) error {
	value = strings.TrimSpace(value)

	switch field.FieldType {
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("deve ser um número válido")
		}

	case "select", "radio":
		if field.Options != nil {
			var options []string
			if err := json.Unmarshal(field.Options, &options); err == nil {
				valueNorm := strings.ToLower(strings.TrimSpace(value))
				found := false
				for _, opt := range options {
					if strings.ToLower(strings.TrimSpace(opt)) == valueNorm {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("valor '%s' não está entre as opções permitidas: %v", value, options)
				}
			}
		}

	case "multiselect", "checkbox":
		if field.Options != nil {
			var options []string
			if err := json.Unmarshal(field.Options, &options); err == nil {
				values := strings.Split(value, ",")
				for _, v := range values {
					valueNorm := strings.ToLower(strings.TrimSpace(v))
					found := false
					for _, opt := range options {
						if strings.ToLower(strings.TrimSpace(opt)) == valueNorm {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("valor '%s' não está entre as opções permitidas: %v", v, options)
					}
				}
			}
		}

	case "email":
		if !strings.Contains(value, "@") {
			return fmt.Errorf("deve ser um email válido")
		}

	case "tel", "phone":
		cleaned := strings.ReplaceAll(value, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		cleaned = strings.ReplaceAll(cleaned, "(", "")
		cleaned = strings.ReplaceAll(cleaned, ")", "")
		if len(cleaned) < 8 {
			return fmt.Errorf("deve ser um telefone válido")
		}
	}

	return nil
}

func findLocationByAddress(address string, locationMap map[string]uuid.UUID, locations []models.LocationClass) (*uuid.UUID, error) {
	if address == "" {
		return nil, nil
	}

	addressNorm := strings.ToLower(strings.TrimSpace(address))

	if locationID, exists := locationMap[addressNorm]; exists {
		return &locationID, nil
	}

	for _, loc := range locations {
		locAddrNorm := strings.ToLower(strings.TrimSpace(loc.Address))
		if strings.Contains(locAddrNorm, addressNorm) || strings.Contains(addressNorm, locAddrNorm) {
			return &loc.ID, nil
		}
	}

	return nil, fmt.Errorf("turma/localização '%s' não encontrada para este curso", address)
}

func (p *EnrollmentImportProcessor) processRow(ctx context.Context, cursoID int, row EnrollmentRow, customFields []models.CustomField, locationMap map[string]uuid.UUID, locationClasses []models.LocationClass) (*uuid.UUID, error) {
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

	if err := validateCustomFields(row, customFields); err != nil {
		return nil, err
	}

	if len(locationClasses) > 1 && row.Turma == "" {
		return nil, fmt.Errorf("coluna 'Turma' é obrigatória quando o curso tem múltiplas turmas/localizações")
	}

	var enrolledUnit *models.EnrolledUnit
	if row.Turma != "" {
		locationID, err := findLocationByAddress(row.Turma, locationMap, locationClasses)
		if err != nil {
			return nil, err
		}

		if locationID != nil {
			for _, loc := range locationClasses {
				if loc.ID == *locationID {
					enrolledUnit = &models.EnrolledUnit{
						ID:             locationID.String(),
						CursoID:        cursoID,
						Address:        loc.Address,
						Neighborhood:   loc.Neighborhood,
						Vacancies:      loc.Vacancies,
						ClassStartDate: loc.ClassStartDate.Format("2006-01-02"),
						ClassEndDate:   loc.ClassEndDate.Format("2006-01-02"),
						ClassTime:      loc.ClassTime,
						ClassDays:      loc.ClassDays,
					}
					break
				}
			}
		}
	} else if len(locationClasses) == 1 {
		loc := locationClasses[0]
		enrolledUnit = &models.EnrolledUnit{
			ID:             loc.ID.String(),
			CursoID:        cursoID,
			Address:        loc.Address,
			Neighborhood:   loc.Neighborhood,
			Vacancies:      loc.Vacancies,
			ClassStartDate: loc.ClassStartDate.Format("2006-01-02"),
			ClassEndDate:   loc.ClassEndDate.Format("2006-01-02"),
			ClassTime:      loc.ClassTime,
			ClassDays:      loc.ClassDays,
		}
	}

	var customFieldsData datatypes.JSON
	if len(row.CustomFields) > 0 {
		customFieldsMap := make(map[string]interface{})
		for key, value := range row.CustomFields {
			if value != "" {
				customFieldsMap[key] = value
			}
		}
		if len(customFieldsMap) > 0 {
			customFieldsBytes, err := json.Marshal(customFieldsMap)
			if err != nil {
				return nil, fmt.Errorf("erro ao serializar custom fields: %w", err)
			}
			customFieldsData = datatypes.JSON(customFieldsBytes)
		}
	}

	inscricao := &models.Inscricao{
		CursoID:          cursoID,
		CPF:              cpf,
		Name:             strings.TrimSpace(row.NomeCompleto),
		Email:            strings.TrimSpace(row.Email),
		Phone:            strings.TrimSpace(row.Telefone),
		Age:              row.Idade,
		Address:          strings.TrimSpace(row.Endereco),
		Neighborhood:     strings.TrimSpace(row.Bairro),
		CustomFieldsData: customFieldsData,
		EnrolledUnit:     enrolledUnit,
	}

	if err := p.inscricaoService.Create(ctx, inscricao); err != nil {
		return nil, err
	}

	return &inscricao.ID, nil
}

func (p *EnrollmentImportProcessor) parseCSV(filePath string, fieldMappings map[string]string) ([]EnrollmentRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	columnMap := make(map[string]int)
	for i, colName := range header {
		normalizedName := normalizeFieldName(colName)
		columnMap[normalizedName] = i
	}

	hasRequiredField := func(variants []string) bool {
		for _, variant := range variants {
			if _, exists := columnMap[normalizeFieldName(variant)]; exists {
				return true
			}
		}
		return false
	}

	if !hasRequiredField([]string{"nome_completo", "nome", "name"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna de nome (nome_completo, nome ou name)")
	}
	if !hasRequiredField([]string{"cpf"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna 'cpf'")
	}
	if !hasRequiredField([]string{"email"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna 'email'")
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

		if len(record) == 0 {
			continue
		}

		row := EnrollmentRow{
			CustomFields: make(map[string]string),
		}

		for colName, idx := range columnMap {
			if idx >= len(record) {
				continue
			}

			value := strings.TrimSpace(record[idx])

			if matchesFieldName(colName, "nome_completo") || matchesFieldName(colName, "nome") || matchesFieldName(colName, "name") {
				row.NomeCompleto = value
			} else if matchesFieldName(colName, "cpf") {
				row.CPF = value
			} else if matchesFieldName(colName, "idade") || matchesFieldName(colName, "age") {
				row.Idade, _ = strconv.Atoi(value)
			} else if matchesFieldName(colName, "telefone") || matchesFieldName(colName, "phone") {
				row.Telefone = value
			} else if matchesFieldName(colName, "email") {
				row.Email = value
			} else if matchesFieldName(colName, "endereco") || matchesFieldName(colName, "address") {
				row.Endereco = value
			} else if matchesFieldName(colName, "bairro") || matchesFieldName(colName, "neighborhood") {
				row.Bairro = value
			} else if matchesFieldName(colName, "turma") || matchesFieldName(colName, "location") || matchesFieldName(colName, "classe") {
				row.Turma = value
			} else {
				if originalTitle, exists := fieldMappings[colName]; exists {
					row.CustomFields[originalTitle] = value
				} else {
					row.CustomFields[header[idx]] = value
				}
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (p *EnrollmentImportProcessor) parseXLSX(filePath string, fieldMappings map[string]string) ([]EnrollmentRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo XLSX: %w", err)
	}
	defer f.Close()

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

	header := allRows[0]

	columnMap := make(map[string]int)
	for i, colName := range header {
		normalizedName := normalizeFieldName(colName)
		columnMap[normalizedName] = i
	}

	hasRequiredField := func(variants []string) bool {
		for _, variant := range variants {
			if _, exists := columnMap[normalizeFieldName(variant)]; exists {
				return true
			}
		}
		return false
	}

	if !hasRequiredField([]string{"nome_completo", "nome", "name"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna de nome (nome_completo, nome ou name)")
	}
	if !hasRequiredField([]string{"cpf"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna 'cpf'")
	}
	if !hasRequiredField([]string{"email"}) {
		return nil, fmt.Errorf("cabeçalho deve conter coluna 'email'")
	}

	var rows []EnrollmentRow

	for i := 1; i < len(allRows); i++ {
		record := allRows[i]

		if len(record) == 0 {
			continue
		}

		row := EnrollmentRow{
			CustomFields: make(map[string]string),
		}

		for colName, idx := range columnMap {
			if idx >= len(record) {
				continue
			}

			value := strings.TrimSpace(record[idx])

			if matchesFieldName(colName, "nome_completo") || matchesFieldName(colName, "nome") || matchesFieldName(colName, "name") {
				row.NomeCompleto = value
			} else if matchesFieldName(colName, "cpf") {
				row.CPF = value
			} else if matchesFieldName(colName, "idade") || matchesFieldName(colName, "age") {
				row.Idade, _ = strconv.Atoi(value)
			} else if matchesFieldName(colName, "telefone") || matchesFieldName(colName, "phone") {
				row.Telefone = value
			} else if matchesFieldName(colName, "email") {
				row.Email = value
			} else if matchesFieldName(colName, "endereco") || matchesFieldName(colName, "address") {
				row.Endereco = value
			} else if matchesFieldName(colName, "bairro") || matchesFieldName(colName, "neighborhood") {
				row.Bairro = value
			} else if matchesFieldName(colName, "turma") || matchesFieldName(colName, "location") || matchesFieldName(colName, "classe") {
				row.Turma = value
			} else {
				if originalTitle, exists := fieldMappings[colName]; exists {
					row.CustomFields[originalTitle] = value
				} else {
					row.CustomFields[header[idx]] = value
				}
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}
