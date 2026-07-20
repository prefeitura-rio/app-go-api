package repository_test

import (
	"context"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// TestCustomFieldOrder_Integration verifies that custom_fields are persisted and
// read back in the exact order sent, across create, update and duplicate flows.
// This exercises the real Postgres path (display_order column + ORDER BY on preload).
func TestCustomFieldOrder_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	ctx := context.Background()
	repo := repository.NewCursoRepository(db)

	titlesOf := func(fields []models.CustomField) []string {
		out := make([]string, len(fields))
		for i, f := range fields {
			out[i] = f.Title
		}
		return out
	}
	assertOrder := func(t *testing.T, got []models.CustomField, want []string) {
		t.Helper()
		gotTitles := titlesOf(got)
		if len(gotTitles) != len(want) {
			t.Fatalf("expected %d custom fields, got %d (%v)", len(want), len(gotTitles), gotTitles)
		}
		for i := range want {
			if gotTitles[i] != want[i] {
				t.Fatalf("order mismatch at index %d: want %v, got %v", i, want, gotTitles)
			}
		}
	}

	// Deliberately non-alphabetical order — a naive query on a UUID PK would not
	// return this order without display_order.
	inputOrder := []string{"Zeta", "Alpha", "Mike", "Bravo", "Yankee"}

	curso := &models.Curso{
		Titulo:     "Curso Ordem Custom Fields Integration",
		Modalidade: models.ModalidadeRemoto,
		Status:     models.StatusCursoDraft,
	}
	for _, title := range inputOrder {
		curso.CustomFields = append(curso.CustomFields, models.CustomField{
			Title:     title,
			FieldType: "text",
		})
	}

	// --- Create flow (mirrors CursoService.Create) ---
	id, err := repo.Create(ctx, curso)
	if err != nil {
		t.Fatalf("Create curso: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM custom_fields WHERE curso_id = ?", id)
		db.Exec("DELETE FROM cursos WHERE id = ?", id)
	})
	for i := range curso.CustomFields {
		curso.CustomFields[i].CursoID = id
	}
	if err := repo.CreateCustomFields(ctx, curso.CustomFields); err != nil {
		t.Fatalf("CreateCustomFields: %v", err)
	}

	// GET after create — must match input order
	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}
	assertOrder(t, got.CustomFields, inputOrder)
	t.Logf("after create: %v", titlesOf(got.CustomFields))

	// Run GET a few more times to catch any non-deterministic ordering.
	for n := 0; n < 5; n++ {
		again, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID repeat %d: %v", n, err)
		}
		assertOrder(t, again.CustomFields, inputOrder)
	}

	// --- Update flow (delete + recreate via updateCustomFieldsWithTx) ---
	reordered := []string{"Bravo", "Yankee", "Zeta", "Alpha", "Mike"}
	updateCurso := got // reuse loaded course to keep required fields
	updateCurso.CustomFields = nil
	for _, title := range reordered {
		updateCurso.CustomFields = append(updateCurso.CustomFields, models.CustomField{
			Title:     title,
			FieldType: "text",
		})
	}
	if err := repo.Update(ctx, updateCurso); err != nil {
		t.Fatalf("Update curso: %v", err)
	}

	gotAfterUpdate, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	assertOrder(t, gotAfterUpdate.CustomFields, reordered)
	t.Logf("after update: %v", titlesOf(gotAfterUpdate.CustomFields))

	// --- Duplicate flow (copy array in received order, drop IDs, create new) ---
	dup := &models.Curso{
		Titulo:     "Curso Ordem Custom Fields Integration (Dup)",
		Modalidade: models.ModalidadeRemoto,
		Status:     models.StatusCursoDraft,
	}
	for _, f := range gotAfterUpdate.CustomFields {
		dup.CustomFields = append(dup.CustomFields, models.CustomField{
			Title:     f.Title,
			FieldType: f.FieldType,
		})
	}
	dupID, err := repo.Create(ctx, dup)
	if err != nil {
		t.Fatalf("Create dup curso: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM custom_fields WHERE curso_id = ?", dupID)
		db.Exec("DELETE FROM cursos WHERE id = ?", dupID)
	})
	for i := range dup.CustomFields {
		dup.CustomFields[i].CursoID = dupID
	}
	if err := repo.CreateCustomFields(ctx, dup.CustomFields); err != nil {
		t.Fatalf("CreateCustomFields dup: %v", err)
	}
	gotDup, err := repo.GetByID(ctx, dupID)
	if err != nil {
		t.Fatalf("GetByID dup: %v", err)
	}
	assertOrder(t, gotDup.CustomFields, reordered)
	t.Logf("after duplicate: %v", titlesOf(gotDup.CustomFields))
}
