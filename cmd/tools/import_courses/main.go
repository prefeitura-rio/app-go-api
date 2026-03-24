package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type CourseSchedule struct {
	Vacancies      int    `json:"vacancies"`
	ClassStartDate string `json:"class_start_date"`
	ClassEndDate   string `json:"class_end_date"`
	ClassDays      string `json:"class_days"`
	ClassTime      string `json:"class_time"`
}

type LocationClass struct {
	Address      string           `json:"address"`
	Neighborhood string           `json:"neighborhood"`
	Schedules    []CourseSchedule `json:"schedules"`
}

type FieldOption struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type CustomField struct {
	Title     string        `json:"title"`
	FieldType string        `json:"field_type"`
	Required  bool          `json:"required"`
	Options   []FieldOption `json:"options,omitempty"`
}

type Categoria struct {
	ID   int    `json:"id"`
	Nome string `json:"nome,omitempty"`
}

type CoursePayload struct {
	Title                string          `json:"title"`
	Theme                string          `json:"theme,omitempty"`
	Description          string          `json:"description,omitempty"`
	Modalidade           string          `json:"modalidade"`
	Status               string          `json:"status"`
	EnrollmentStartDate  *string         `json:"enrollment_start_date,omitempty"`
	EnrollmentEndDate    *string         `json:"enrollment_end_date,omitempty"`
	Workload             string          `json:"workload,omitempty"`
	TargetAudience       string          `json:"target_audience,omitempty"`
	Accessibility        string          `json:"accessibility,omitempty"`
	HasCertificate       bool            `json:"has_certificate"`
	PreRequisitos        string          `json:"pre_requisitos,omitempty"`
	InstitutionalLogo    string          `json:"institutional_logo,omitempty"`
	IsVisible            *bool           `json:"is_visible,omitempty"`
	OrgaoID              string          `json:"orgao_id,omitempty"`
	IsExternalPartner    *bool           `json:"is_external_partner,omitempty"`
	CourseManagementType string          `json:"course_management_type,omitempty"`
	Categorias           []Categoria     `json:"categorias,omitempty"`
	Locations            []LocationClass `json:"locations,omitempty"`
	CustomFields         []CustomField   `json:"custom_fields,omitempty"`
}

type ImportResult struct {
	Line     int    `json:"line"`
	Title    string `json:"title"`
	Success  bool   `json:"success"`
	CourseID int    `json:"course_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type ImportReport struct {
	TotalProcessed int            `json:"total_processed"`
	Successful     int            `json:"successful"`
	Failed         int            `json:"failed"`
	Results        []ImportResult `json:"results"`
	StartedAt      time.Time      `json:"started_at"`
	FinishedAt     time.Time      `json:"finished_at"`
	Duration       string         `json:"duration"`
}

var (
	filePath    string
	apiURL      string
	dryRun      bool
	startLine   int
	limitLines  int
	verbose     bool
	portForward bool
	k8sContext  string
	k8sNS       string
	k8sSvc      string
	localPort   string
	remotePort  string

	categoriaMap = map[string]int{}
)

var portForwardCmd *exec.Cmd

func init() {
	flag.StringVar(&filePath, "file", "", "Caminho para o arquivo TSV")
	flag.StringVar(&apiURL, "api", "http://localhost:8080", "URL base da API")
	flag.BoolVar(&dryRun, "dry-run", false, "Apenas validar sem enviar para a API")
	flag.IntVar(&startLine, "start", 1, "Linha inicial (1-based)")
	flag.IntVar(&limitLines, "limit", 0, "Limite de linhas (0 = todas)")
	flag.BoolVar(&verbose, "verbose", false, "Mostrar payloads detalhados")
	flag.BoolVar(&portForward, "pf", false, "Criar port-forward automaticamente via kubectl")
	flag.StringVar(&k8sContext, "context", "", "Contexto do Kubernetes (ex: gke_rj-superapp_us-central1_application)")
	flag.StringVar(&k8sNS, "ns", "go", "Namespace do Kubernetes")
	flag.StringVar(&k8sSvc, "svc", "go", "Nome do serviço no Kubernetes")
	flag.StringVar(&localPort, "local-port", "8080", "Porta local para o port-forward")
	flag.StringVar(&remotePort, "remote-port", "80", "Porta remota do serviço")
}

func main() {
	flag.Parse()

	if filePath == "" {
		fmt.Println("Erro: --file é obrigatório")
		flag.Usage()
		os.Exit(1)
	}

	setupSignalHandler()

	if portForward && !dryRun {
		if err := startPortForward(); err != nil {
			fmt.Printf("Erro ao iniciar port-forward: %v\n", err)
			os.Exit(1)
		}
		defer stopPortForward()
	}

	if !dryRun {
		if err := loadCategorias(); err != nil {
			fmt.Printf("Erro ao carregar categorias: %v\n", err)
			os.Exit(1)
		}
	}

	report, err := importCourses()
	if err != nil {
		fmt.Printf("Erro fatal: %v\n", err)
		os.Exit(1)
	}

	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	reportFile := fmt.Sprintf("import_report_%s.json", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(reportFile, reportJSON, 0644); err != nil {
		fmt.Printf("Aviso: não foi possível salvar relatório: %v\n", err)
	} else {
		fmt.Printf("\nRelatório salvo em: %s\n", reportFile)
	}

	fmt.Printf("\n=== RESUMO ===\n")
	fmt.Printf("Total processado: %d\n", report.TotalProcessed)
	fmt.Printf("Sucesso: %d\n", report.Successful)
	fmt.Printf("Falhas: %d\n", report.Failed)
	fmt.Printf("Duração: %s\n", report.Duration)
}

type CourseGroup struct {
	Title    string
	Records  [][]string
	Lines    []int
	ColIndex *ColumnIndex
}

func importCourses() (*ImportReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Aviso: erro ao fechar arquivo: %v\n", err)
		}
	}()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	colIndex := buildColumnIndex(header)

	groups := make(map[string]*CourseGroup)
	groupOrder := []string{}
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Erro linha %d: %v\n", lineNum+2, err)
			continue
		}
		lineNum++

		if lineNum < startLine {
			continue
		}

		titulo := ""
		if idx, ok := colIndex.byName["titulo"]; ok && idx < len(record) {
			titulo = strings.TrimSpace(record[idx])
		}
		if titulo == "" {
			continue
		}

		if _, exists := groups[titulo]; !exists {
			groups[titulo] = &CourseGroup{
				Title:    titulo,
				Records:  [][]string{},
				Lines:    []int{},
				ColIndex: colIndex,
			}
			groupOrder = append(groupOrder, titulo)
		}

		groups[titulo].Records = append(groups[titulo].Records, record)
		groups[titulo].Lines = append(groups[titulo].Lines, lineNum+1)
	}

	report := &ImportReport{StartedAt: time.Now(), Results: []ImportResult{}}
	processed := 0

	for _, titulo := range groupOrder {
		if limitLines > 0 && processed >= limitLines {
			break
		}

		group := groups[titulo]
		result := processGroup(group)
		report.Results = append(report.Results, result)
		processed++

		if result.Success {
			report.Successful++
		} else {
			report.Failed++
		}
		report.TotalProcessed = processed

		if !dryRun {
			time.Sleep(100 * time.Millisecond)
		}
	}

	report.FinishedAt = time.Now()
	report.Duration = report.FinishedAt.Sub(report.StartedAt).String()
	return report, nil
}

func processGroup(group *CourseGroup) ImportResult {
	result := ImportResult{
		Line:  group.Lines[0],
		Title: group.Title,
	}

	payload, err := buildGroupPayload(group)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if verbose {
		fmt.Printf("Payload: %s\n", string(payloadJSON))
	}

	turmasInfo := fmt.Sprintf("%d turma(s)", len(payload.Locations))

	if dryRun {
		fmt.Printf("[DRY-RUN] %s - %s (linhas: %v)\n", payload.Title, turmasInfo, group.Lines)
		result.Success = true
		return result
	}

	courseID, err := sendToAPI(payloadJSON)
	if err != nil {
		result.Error = err.Error()
		fmt.Printf("[ERRO] %s - %v\n", payload.Title, err)
		return result
	}

	result.Success = true
	result.CourseID = courseID
	fmt.Printf("[OK] %s - %s (ID: %d)\n", payload.Title, turmasInfo, courseID)
	return result
}

func buildGroupPayload(group *CourseGroup) (*CoursePayload, error) {
	if len(group.Records) == 0 {
		return nil, fmt.Errorf("grupo vazio")
	}

	firstRecord := group.Records[0]
	colIndex := group.ColIndex

	payload := &CoursePayload{
		Title:                group.Title,
		Theme:                getField(firstRecord, colIndex, "theme"),
		Description:          getField(firstRecord, colIndex, "descricao"),
		Modalidade:           getField(firstRecord, colIndex, "modalidade"),
		Status:               "draft",
		Workload:             getField(firstRecord, colIndex, "workload"),
		TargetAudience:       getField(firstRecord, colIndex, "target_audience"),
		Accessibility:        getField(firstRecord, colIndex, "accessibility"),
		PreRequisitos:        getField(firstRecord, colIndex, "pre_requisitos"),
		InstitutionalLogo:    getField(firstRecord, colIndex, "institutional_logo"),
		OrgaoID:              getField(firstRecord, colIndex, "orgao_id"),
		CourseManagementType: getField(firstRecord, colIndex, "course_management_type"),
	}

	if payload.Modalidade == "" {
		payload.Modalidade = "Presencial"
	}
	if payload.CourseManagementType == "" {
		payload.CourseManagementType = "OWN_ORG"
	}

	if es := getField(firstRecord, colIndex, "enrollment_start_date"); es != "" {
		f := formatDateTime(es)
		payload.EnrollmentStartDate = &f
	}
	if ee := getField(firstRecord, colIndex, "enrollment_end_date"); ee != "" {
		f := formatDateTime(ee)
		payload.EnrollmentEndDate = &f
	}

	payload.HasCertificate = parseBool(getField(firstRecord, colIndex, "has_certificate"))
	isVisible := parseBool(getField(firstRecord, colIndex, "is_visible"))
	payload.IsVisible = &isVisible
	isExternal := parseBool(getField(firstRecord, colIndex, "is_external_partner"))
	payload.IsExternalPartner = &isExternal

	if catStr := getField(firstRecord, colIndex, "categoria"); catStr != "" {
		catNames := strings.Split(catStr, ",")
		for _, catName := range catNames {
			catName = strings.TrimSpace(catName)
			if catName == "" {
				continue
			}
			catID, err := getOrCreateCategoria(catName)
			if err != nil {
				return nil, fmt.Errorf("erro ao obter/criar categoria %s: %w", catName, err)
			}
			if catID > 0 {
				payload.Categorias = append(payload.Categorias, Categoria{ID: catID})
			}
		}
	}

	for _, record := range group.Records {
		address := getField(record, colIndex, "location_classes")
		neighborhood := getField(record, colIndex, "location_classes.neighborhood")

		if address != "" && neighborhood != "" {
			vacancies := 30
			if v, err := strconv.Atoi(getField(record, colIndex, "course_schedules.vacancies")); err == nil {
				vacancies = v
			}

			schedule := CourseSchedule{
				ClassStartDate: formatDate(getField(record, colIndex, "course_schedules.class_start_date")),
				ClassEndDate:   formatDate(getField(record, colIndex, "course_schedules.class_end_date")),
				ClassDays:      getField(record, colIndex, "course_schedules.class_days"),
				ClassTime:      getField(record, colIndex, "course_schedules.class_time"),
				Vacancies:      vacancies,
			}

			payload.Locations = append(payload.Locations, LocationClass{
				Address:      address,
				Neighborhood: neighborhood,
				Schedules:    []CourseSchedule{schedule},
			})
		}
	}

	payload.CustomFields = parseCustomFields(firstRecord, colIndex)
	return payload, nil
}

type ColumnIndex struct {
	byName           map[string]int
	customFieldStart int
	header           []string
}

func buildColumnIndex(header []string) *ColumnIndex {
	index := &ColumnIndex{
		byName:           make(map[string]int),
		customFieldStart: -1,
		header:           header,
	}
	for i, col := range header {
		col = strings.TrimSpace(col)
		if strings.HasPrefix(col, "custom_fields.") {
			if index.customFieldStart == -1 {
				index.customFieldStart = i
			}
		} else {
			index.byName[col] = i
		}
	}
	return index
}

func getField(record []string, colIndex *ColumnIndex, fieldName string) string {
	if idx, ok := colIndex.byName[fieldName]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

func parseCustomFields(record []string, colIndex *ColumnIndex) []CustomField {
	var fields []CustomField

	if colIndex.customFieldStart == -1 {
		return fields
	}

	for i := colIndex.customFieldStart; i < len(record); i += 4 {
		if i >= len(record) {
			break
		}

		title := strings.TrimSpace(record[i])
		if title == "" {
			continue
		}

		fieldType := "text"
		if i+1 < len(record) {
			if ft := strings.TrimSpace(record[i+1]); ft != "" {
				fieldType = ft
			}
		}

		required := false
		if i+2 < len(record) {
			required = parseBool(record[i+2])
		}

		var options []string
		if i+3 < len(record) {
			optStr := strings.TrimSpace(record[i+3])
			if optStr != "" && (fieldType == "radio" || fieldType == "multi select") {
				options = splitOptions(optStr)
			}
		}

		field := CustomField{Title: title, FieldType: fieldType, Required: required}
		if len(options) > 0 {
			for _, opt := range options {
				field.Options = append(field.Options, FieldOption{
					ID:    uuid.New().String(),
					Value: opt,
				})
			}
		}
		fields = append(fields, field)
	}

	return fields
}

func splitOptions(optStr string) []string {
	var options []string

	var delimiter string
	switch {
	case strings.Contains(optStr, ";"):
		delimiter = ";"
	case strings.Contains(optStr, "|"):
		delimiter = "|"
	case strings.Contains(optStr, "\n"):
		delimiter = "\n"
	default:
		delimiter = " "
	}

	if delimiter == " " {
		optStr = strings.ReplaceAll(optStr, "  ", " ")
		for _, w := range strings.Fields(optStr) {
			if w = strings.TrimSpace(w); w != "" {
				options = append(options, w)
			}
		}
	} else {
		parts := strings.Split(optStr, delimiter)
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				options = append(options, p)
			}
		}
	}

	return options
}

func formatDateTime(dt string) string {
	dt = strings.TrimSpace(dt)
	if dt == "" {
		return ""
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, dt); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return dt
}

func formatDate(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return ""
	}
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05"}
	for _, l := range layouts {
		if t, err := time.Parse(l, d); err == nil {
			// Adiciona 3h para compensar GMT-3 (Brasília)
			// Assim 2026-03-03 vira 2026-03-03T03:00:00Z
			// que em BRT é 2026-03-03T00:00:00-03:00
			t = t.Add(3 * time.Hour)
			return t.Format(time.RFC3339)
		}
	}
	return d + "T03:00:00Z"
}

func parseBool(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	return s == "TRUE" || s == "1" || s == "SIM" || s == "YES"
}

type CategoriaResponse struct {
	Data []struct {
		ID   int    `json:"id"`
		Nome string `json:"nome"`
	} `json:"data"`
}

type CategoriaCreateResponse struct {
	ID int `json:"id"`
}

func loadCategorias() error {
	resp, err := http.Get(apiURL + "/api/v1/categorias?pageSize=100")
	if err != nil {
		return fmt.Errorf("erro ao buscar categorias: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var catResp CategoriaResponse
	if err := json.NewDecoder(resp.Body).Decode(&catResp); err != nil {
		return fmt.Errorf("erro ao decodificar categorias: %w", err)
	}

	for _, cat := range catResp.Data {
		categoriaMap[cat.Nome] = cat.ID
	}

	fmt.Printf("Categorias carregadas: %d\n", len(categoriaMap))
	return nil
}

func getOrCreateCategoria(nome string) (int, error) {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return 0, nil
	}

	if id, ok := categoriaMap[nome]; ok {
		return id, nil
	}

	payload := map[string]string{"nome": nome}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiURL+"/api/v1/categorias", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("erro ao criar categoria %s: %w", nome, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("erro ao criar categoria %s: status %d - %s", nome, resp.StatusCode, string(respBody))
	}

	var createResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return 0, fmt.Errorf("erro ao decodificar resposta de criação: %w", err)
	}

	var newID int
	if data, ok := createResp["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			newID = int(id)
		}
	}

	if newID == 0 {
		return 0, fmt.Errorf("não foi possível obter ID da categoria criada: %s", nome)
	}

	categoriaMap[nome] = newID
	fmt.Printf("Categoria criada: %s (ID: %d)\n", nome, newID)
	return newID, nil
}

func sendToAPI(payload []byte) (int, error) {
	url := fmt.Sprintf("%s/api/v1/courses/draft", apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	var resp *http.Response

	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = client.Do(req)
		if err != nil {
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return 0, err
		}
		break
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API retornou %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID int `json:"id"`
		} `json:"data"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if !result.Success {
		return 0, fmt.Errorf("API erro: %s", result.Error)
	}

	return result.Data.ID, nil
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nInterrompido. Encerrando...")
		stopPortForward()
		os.Exit(1)
	}()
}

func startPortForward() error {
	args := []string{"port-forward", "-n", k8sNS, fmt.Sprintf("svc/%s", k8sSvc), fmt.Sprintf("%s:%s", localPort, remotePort)}

	if k8sContext != "" {
		args = append([]string{"--context", k8sContext}, args...)
		fmt.Printf("Iniciando port-forward: kubectl --context %s port-forward -n %s svc/%s %s:%s\n", k8sContext, k8sNS, k8sSvc, localPort, remotePort)
	} else {
		fmt.Printf("Iniciando port-forward: kubectl port-forward -n %s svc/%s %s:%s\n", k8sNS, k8sSvc, localPort, remotePort)
	}

	portForwardCmd = exec.Command("kubectl", args...)
	portForwardCmd.Stdout = nil
	portForwardCmd.Stderr = nil

	if err := portForwardCmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar kubectl: %w", err)
	}

	fmt.Println("Aguardando port-forward ficar pronto...")

	apiURL = fmt.Sprintf("http://localhost:%s", localPort)

	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		if isAPIReady() {
			fmt.Println("Port-forward pronto!")
			return nil
		}
	}

	stopPortForward()
	return fmt.Errorf("timeout aguardando port-forward (15s)")
}

func stopPortForward() {
	if portForwardCmd != nil && portForwardCmd.Process != nil {
		fmt.Println("Encerrando port-forward...")
		_ = portForwardCmd.Process.Kill()
		_ = portForwardCmd.Wait()
		portForwardCmd = nil
	}
}

func isAPIReady() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/health", apiURL))
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}
