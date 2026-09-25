package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// Quem só preencheu o currículo, sem inscrição nem candidatura, também precisa
// ter o cadastro sincronizado: é de lá que o banco de currículos tira o nome.
func TestCitizenSnapshotRepository_Integration_SincronizaQuemSoTemCurriculo(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()

	const cpf = "99999999911"
	require.NoError(t, tx.Exec(`INSERT INTO emp_curriculos (cpf) VALUES (?)`, cpf).Error)

	cpfs, err := repository.NewCitizenSnapshotRepository(tx).GetCPFsWithEnrollments(context.Background(), 24*time.Hour, 1000000)

	require.NoError(t, err)
	assert.Contains(t, cpfs, cpf)
}
