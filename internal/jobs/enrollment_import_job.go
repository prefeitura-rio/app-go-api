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

// buildScheduleMap creates a map to find schedules by address, time, or days
// Key format: "address|time|days" (all normalized and lowercased)
// Also supports UUID-based keys for direct schedule/class lookup
func buildScheduleMap(locations []models.LocationClass, remoteClassLoaded bool, remoteClass *models.RemoteClass) map[string]struct {
	LocationID uuid.UUID
	ScheduleID uuid.UUID
} {
	scheduleMap := make(map[string]struct {
		LocationID uuid.UUID
		ScheduleID uuid.UUID
	})

	for _, loc := range locations {
		normalizedAddr := strings.ToLower(strings.TrimSpace(loc.Address))

		for _, schedule := range loc.Schedules {
			// Create multiple keys for better matching
			normalizedTime := strings.ToLower(strings.TrimSpace(schedule.ClassTime))
			normalizedDays := strings.ToLower(strings.TrimSpace(schedule.ClassDays))

			// Key by UUID (schedule_id)
			scheduleMap[schedule.ID.String()] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}

			// Key by UUID (location_id)
			scheduleMap[loc.ID.String()] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}

			// Key by address only
			addressKey := normalizedAddr
			scheduleMap[addressKey] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}

			// Key by address + time
			addressTimeKey := normalizedAddr + "|" + normalizedTime
			scheduleMap[addressTimeKey] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}

			// Key by address + days
			addressDaysKey := normalizedAddr + "|" + normalizedDays
			scheduleMap[addressDaysKey] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}

			// Key by address + time + days
			fullKey := normalizedAddr + "|" + normalizedTime + "|" + normalizedDays
			scheduleMap[fullKey] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{loc.ID, schedule.ID}
		}
	}

	// Add remote class schedules
	if remoteClassLoaded && remoteClass != nil {
		for _, schedule := range remoteClass.Schedules {
			normalizedTime := strings.ToLower(strings.TrimSpace(schedule.ClassTime))
			normalizedDays := strings.ToLower(strings.TrimSpace(schedule.ClassDays))

			// Key by UUID (schedule_id)
			scheduleMap[schedule.ID.String()] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{remoteClass.ID, schedule.ID}

			// Key by UUID (remote_class_id)
			scheduleMap[remoteClass.ID.String()] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{remoteClass.ID, schedule.ID}

			// Key by time + days (for remote courses without address)
			timeDaysKey := normalizedTime + "|" + normalizedDays
			scheduleMap[timeDaysKey] = struct {
				LocationID uuid.UUID
				ScheduleID uuid.UUID
			}{remoteClass.ID, schedule.ID}
		}
	}

	return scheduleMap
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
	if err := p.db.Where("curso_id = ?", metadata.CursoID).Preload("Schedules").Find(&locationClasses).Error; err != nil {
		log.Printf("Aviso: erro ao carregar location_classes: %v", err)
	}

	var remoteClass models.RemoteClass
	remoteClassLoaded := false
	if err := p.db.Where("curso_id = ?", metadata.CursoID).Preload("Schedules").First(&remoteClass).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("Aviso: erro ao carregar remote_class: %v", err)
		}
	} else {
		remoteClassLoaded = true
	}

	fieldMappings := buildFieldMappings(customFields)
	scheduleMap := buildScheduleMap(locationClasses, remoteClassLoaded, &remoteClass)

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

		enrollmentID, err := p.processRow(ctx, metadata.CursoID, row, customFields, scheduleMap, locationClasses, remoteClassLoaded, &remoteClass)
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

// findScheduleByTurma tries to match the "Turma" field from CSV to a schedule
// It tries multiple strategies: UUID match, exact match, partial match by address, address+time, address+days
func findScheduleByTurma(turma string, scheduleMap map[string]struct {
	LocationID uuid.UUID
	ScheduleID uuid.UUID
}, locations []models.LocationClass, remoteClassLoaded bool, remoteClass *models.RemoteClass) (*uuid.UUID, *uuid.UUID, error) {
	if turma == "" {
		return nil, nil, nil
	}

	turmaNorm := strings.ToLower(strings.TrimSpace(turma))

	// Try UUID parsing first (supports schedule_id or class_id)
	if parsedUUID, err := uuid.Parse(turmaNorm); err == nil {
		if match, exists := scheduleMap[parsedUUID.String()]; exists {
			return &match.LocationID, &match.ScheduleID, nil
		}
	}

	// Try exact match
	if match, exists := scheduleMap[turmaNorm]; exists {
		return &match.LocationID, &match.ScheduleID, nil
	}

	// Try partial matching - split by | or - or common separators
	parts := strings.FieldsFunc(turmaNorm, func(r rune) bool {
		return r == '|' || r == '-' || r == ',' || r == ';'
	})

	if len(parts) >= 2 {
		// Try address|time OR time|days (for remote courses)
		key := strings.TrimSpace(parts[0]) + "|" + strings.TrimSpace(parts[1])
		if match, exists := scheduleMap[key]; exists {
			return &match.LocationID, &match.ScheduleID, nil
		}
	}

	if len(parts) >= 3 {
		// Try address|time|days
		key := strings.TrimSpace(parts[0]) + "|" + strings.TrimSpace(parts[1]) + "|" + strings.TrimSpace(parts[2])
		if match, exists := scheduleMap[key]; exists {
			return &match.LocationID, &match.ScheduleID, nil
		}
	}

	// Fallback: fuzzy match by address for presential courses
	for _, loc := range locations {
		locAddrNorm := strings.ToLower(strings.TrimSpace(loc.Address))
		if strings.Contains(locAddrNorm, turmaNorm) || strings.Contains(turmaNorm, locAddrNorm) {
			// Return first schedule of this location
			if len(loc.Schedules) > 0 {
				return &loc.ID, &loc.Schedules[0].ID, nil
			}
		}
	}

	// Fallback: fuzzy match by time+days for remote courses
	if remoteClassLoaded && remoteClass != nil && len(remoteClass.Schedules) > 0 {
		for _, schedule := range remoteClass.Schedules {
			timeNorm := strings.ToLower(strings.TrimSpace(schedule.ClassTime))
			daysNorm := strings.ToLower(strings.TrimSpace(schedule.ClassDays))

			if (strings.Contains(turmaNorm, timeNorm) && strings.Contains(turmaNorm, daysNorm)) ||
			   (strings.Contains(timeNorm, turmaNorm) || strings.Contains(daysNorm, turmaNorm)) {
				return &remoteClass.ID, &schedule.ID, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("turma '%s' não encontrada para este curso", turma)
}

func (p *EnrollmentImportProcessor) processRow(ctx context.Context, cursoID int, row EnrollmentRow, customFields []models.CustomField, scheduleMap map[string]struct {
	LocationID uuid.UUID
	ScheduleID uuid.UUID
}, locationClasses []models.LocationClass, remoteClassLoaded bool, remoteClass *models.RemoteClass) (*uuid.UUID, error) {
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

	// Count total schedules across all locations and remote classes
	totalSchedules := 0
	for _, loc := range locationClasses {
		totalSchedules += len(loc.Schedules)
	}
	if remoteClassLoaded && remoteClass != nil {
		totalSchedules += len(remoteClass.Schedules)
	}

	if totalSchedules > 1 && row.Turma == "" {
		return nil, fmt.Errorf("coluna 'Turma' é obrigatória quando o curso tem múltiplas turmas")
	}

	var enrolledUnit *models.EnrolledUnit
	var scheduleID *uuid.UUID

	if row.Turma != "" {
		locationID, foundScheduleID, err := findScheduleByTurma(row.Turma, scheduleMap, locationClasses, remoteClassLoaded, remoteClass)
		if err != nil {
			return nil, err
		}

		if locationID != nil && foundScheduleID != nil {
			scheduleID = foundScheduleID

			// Try to find in location classes
			found := false
			for _, loc := range locationClasses {
				if loc.ID == *locationID {
					// Convert schedules to EnrolledUnitSchedule format
					enrolledSchedules := make([]models.EnrolledUnitSchedule, 0, len(loc.Schedules))
					for _, schedule := range loc.Schedules {
						enrolledSchedules = append(enrolledSchedules, models.EnrolledUnitSchedule{
							ID:             schedule.ID.String(),
							LocationID:     loc.ID.String(),
							Vacancies:      schedule.Vacancies,
							ClassStartDate: schedule.ClassStartDate.Format("2006-01-02"),
							ClassEndDate:   schedule.ClassEndDate.Format("2006-01-02"),
							ClassTime:      schedule.ClassTime,
							ClassDays:      schedule.ClassDays,
							CreatedAt:      schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
							UpdatedAt:      schedule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
						})
					}

					enrolledUnit = &models.EnrolledUnit{
						ID:           loc.ID.String(),
						CursoID:      cursoID,
						Address:      loc.Address,
						Neighborhood: loc.Neighborhood,
						Schedules:    enrolledSchedules,
						CreatedAt:    loc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
						UpdatedAt:    loc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					}
					found = true
					break
				}
			}

			// If not found in location classes, try remote class
			if !found && remoteClassLoaded && remoteClass != nil && remoteClass.ID == *locationID {
				// Convert remote schedules to EnrolledUnitSchedule format
				enrolledSchedules := make([]models.EnrolledUnitSchedule, 0, len(remoteClass.Schedules))
				for _, schedule := range remoteClass.Schedules {
					enrolledSchedules = append(enrolledSchedules, models.EnrolledUnitSchedule{
						ID:             schedule.ID.String(),
						LocationID:     remoteClass.ID.String(),
						Vacancies:      schedule.Vacancies,
						ClassStartDate: schedule.ClassStartDate.Format("2006-01-02"),
						ClassEndDate:   schedule.ClassEndDate.Format("2006-01-02"),
						ClassTime:      schedule.ClassTime,
						ClassDays:      schedule.ClassDays,
						CreatedAt:      schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
						UpdatedAt:      schedule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					})
				}

				enrolledUnit = &models.EnrolledUnit{
					ID:        remoteClass.ID.String(),
					CursoID:   cursoID,
					Schedules: enrolledSchedules,
					CreatedAt: remoteClass.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt: remoteClass.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				}
			}
		}
	} else if totalSchedules == 1 {
		// Auto-select if only one schedule total (either location or remote)
		if len(locationClasses) == 1 && len(locationClasses[0].Schedules) == 1 {
			// Single location class with one schedule
			loc := locationClasses[0]
			schedule := loc.Schedules[0]
			scheduleID = &schedule.ID

			enrolledSchedules := []models.EnrolledUnitSchedule{
				{
					ID:             schedule.ID.String(),
					LocationID:     loc.ID.String(),
					Vacancies:      schedule.Vacancies,
					ClassStartDate: schedule.ClassStartDate.Format("2006-01-02"),
					ClassEndDate:   schedule.ClassEndDate.Format("2006-01-02"),
					ClassTime:      schedule.ClassTime,
					ClassDays:      schedule.ClassDays,
					CreatedAt:      schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:      schedule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				},
			}

			enrolledUnit = &models.EnrolledUnit{
				ID:           loc.ID.String(),
				CursoID:      cursoID,
				Address:      loc.Address,
				Neighborhood: loc.Neighborhood,
				Schedules:    enrolledSchedules,
				CreatedAt:    loc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:    loc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		} else if remoteClassLoaded && remoteClass != nil && len(remoteClass.Schedules) == 1 {
			// Single remote class with one schedule
			schedule := remoteClass.Schedules[0]
			scheduleID = &schedule.ID

			enrolledSchedules := []models.EnrolledUnitSchedule{
				{
					ID:             schedule.ID.String(),
					LocationID:     remoteClass.ID.String(),
					Vacancies:      schedule.Vacancies,
					ClassStartDate: schedule.ClassStartDate.Format("2006-01-02"),
					ClassEndDate:   schedule.ClassEndDate.Format("2006-01-02"),
					ClassTime:      schedule.ClassTime,
					ClassDays:      schedule.ClassDays,
					CreatedAt:      schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:      schedule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				},
			}

			enrolledUnit = &models.EnrolledUnit{
				ID:        remoteClass.ID.String(),
				CursoID:   cursoID,
				Schedules: enrolledSchedules,
				CreatedAt: remoteClass.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: remoteClass.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
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
		ScheduleID:       scheduleID,
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
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing CSV file: %v\n", err)
		}
	}()

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
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing XLSX file: %v\n", err)
		}
	}()

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
