package main

import (
	"encoding/json"
	"testing"
	"time"
)

// pgxLikeDecode reproduz como o driver (pgx v5, protocolo binário) monta um
// timestamptz: a partir do instante absoluto, via time.Unix(), que devolve o
// horário em time.Local. É o mesmo caminho que faz a API serializar as datas.
func pgxLikeDecode(instant time.Time) time.Time {
	return time.Unix(instant.Unix(), int64(instant.Nanosecond()))
}

func TestSetupTimezone_ApiSerializesWithBrasiliaOffset(t *testing.T) {
	original := time.Local
	t.Cleanup(func() { time.Local = original })

	// Instante armazenado no banco: 26/06/2026 23:59 BRT == 27/06 02:59 UTC.
	stored := time.Date(2026, 6, 27, 2, 59, 0, 0, time.UTC)

	t.Run("sem fuso configurado (UTC) a API devolve Z (bug)", func(t *testing.T) {
		time.Local = time.UTC
		b, _ := json.Marshal(pgxLikeDecode(stored))
		if got := string(b); got != `"2026-06-27T02:59:00Z"` {
			t.Fatalf("baseline inesperado: %s", got)
		}
	})

	t.Run("com America/Sao_Paulo a API devolve -03:00", func(t *testing.T) {
		setupTimezone("America/Sao_Paulo")
		b, _ := json.Marshal(pgxLikeDecode(stored))
		want := `"2026-06-26T23:59:00-03:00"`
		if got := string(b); got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
}

func TestSetupTimezone_InvalidKeepsLocal(t *testing.T) {
	original := time.Local
	t.Cleanup(func() { time.Local = original })

	time.Local = time.UTC
	setupTimezone("Nao/Existe")
	if time.Local != time.UTC {
		t.Fatalf("fuso inválido não deveria alterar time.Local, virou %s", time.Local)
	}
}
