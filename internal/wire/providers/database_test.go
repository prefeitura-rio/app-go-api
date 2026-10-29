package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/prefeitura-rio/app-go-api/internal/config"
)

// capturingWriter guarda tudo o que o logger do GORM imprime.
type capturingWriter struct {
	mu  sync.Mutex
	out strings.Builder
}

func (w *capturingWriter) Printf(format string, args ...interface{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Fprintf(&w.out, format, args...)
	w.out.WriteString("\n")
}

func (w *capturingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.out.String()
}

// CPF, nome e e-mail sintéticos usados como termo de busca.
const (
	syntheticCPF   = "00011122233"
	syntheticName  = "Fulano Sintetico"
	syntheticEmail = "fulano.sintetico@example.invalid"
)

func openMockDB(t *testing.T, l logger.Interface) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:               l,
		DisableAutomaticPing: true,
	})
	require.NoError(t, err)
	return db, mock
}

func TestNewGormLogger_InfoLevel_DoesNotLogQueryParameters(t *testing.T) {
	writer := &capturingWriter{}
	db, mock := openMockDB(t, newGormLogger(writer, logger.Info))

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inscricoes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "inscricoes"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	searchTerm := "%" + syntheticCPF + "%"

	// Mesma forma da busca de inscrições por CPF/nome/e-mail (inscricao_repository).
	var total int64
	err := db.Table("inscricoes").
		Where("curso_id = ?", 2882).
		Where("cpf ILIKE ? OR name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm).
		Count(&total).Error
	require.NoError(t, err)

	var rows []map[string]interface{}
	err = db.Table("inscricoes").
		Where("name = ? AND email = ?", syntheticName, syntheticEmail).
		Find(&rows).Error
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	out := writer.String()
	// O SQL continua no log, para depuração, com os placeholders no lugar dos valores.
	assert.Contains(t, out, `SELECT count(*) FROM "inscricoes"`)
	assert.Contains(t, out, "cpf ILIKE $2")
	assert.Contains(t, out, "name = $1")
	assert.Contains(t, out, "email = $2")
	// Nenhum valor de parâmetro (dado pessoal) aparece no log.
	assert.NotContains(t, out, syntheticCPF)
	assert.NotContains(t, out, syntheticName)
	assert.NotContains(t, out, syntheticEmail)
	assert.NotContains(t, out, "2882")
}

func TestNewGormLogger_ErrorTrace_DoesNotLogQueryParameters(t *testing.T) {
	writer := &capturingWriter{}
	db, mock := openMockDB(t, newGormLogger(writer, logger.Error))

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inscricoes"`).
		WillReturnError(fmt.Errorf("falha simulada"))

	var total int64
	err := db.Table("inscricoes").
		Where("cpf ILIKE ?", "%"+syntheticCPF+"%").
		Count(&total).Error
	require.Error(t, err)

	out := writer.String()
	assert.Contains(t, out, "cpf ILIKE $1")
	assert.Contains(t, out, "falha simulada")
	assert.NotContains(t, out, syntheticCPF)
}

func TestGormLogger_ParameterizedQueries(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{"development (local e staging)", "development"},
		{"produção", "prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AppConfig{App: config.AppSettings{Environment: tt.env}}
			l := GormLogger(cfg)

			filter, ok := l.(gorm.ParamsFilter)
			require.True(t, ok, "o logger do GORM precisa filtrar os parâmetros das queries")

			sql, vars := filter.ParamsFilter(context.Background(), "SELECT * FROM inscricoes WHERE cpf = $1", syntheticCPF)
			assert.Equal(t, "SELECT * FROM inscricoes WHERE cpf = $1", sql)
			assert.Empty(t, vars)
		})
	}
}

func TestGormLogLevel(t *testing.T) {
	tests := []struct {
		env  string
		want logger.LogLevel
	}{
		{"development", logger.Info},
		{"prod", logger.Silent},
		{"production", logger.Silent},
		{"test", logger.Silent},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &config.AppConfig{App: config.AppSettings{Environment: tt.env}}
			assert.Equal(t, tt.want, gormLogLevel(cfg))
		})
	}
}

func TestGormTracingPlugin_DoesNotRecordQueryVariables(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	db, mock := openMockDB(t, logger.Discard)
	require.NoError(t, db.Use(gormTracingPlugin(otelgorm.WithTracerProvider(tp), otelgorm.WithoutMetrics())))

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inscricoes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	var total int64
	err := db.Table("inscricoes").
		Where("cpf ILIKE ? OR email ILIKE ?", "%"+syntheticCPF+"%", "%"+syntheticEmail+"%").
		Count(&total).Error
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	spans := recorder.Ended()
	require.NotEmpty(t, spans)

	var statement string
	for _, span := range spans {
		for _, attr := range span.Attributes() {
			if attr.Key == semconv.DBStatementKey {
				statement = attr.Value.AsString()
			}
			assert.NotContains(t, attr.Value.Emit(), syntheticCPF)
			assert.NotContains(t, attr.Value.Emit(), syntheticEmail)
		}
	}
	assert.Contains(t, statement, "cpf ILIKE '?'")
}
