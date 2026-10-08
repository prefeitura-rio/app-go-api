package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeSecretariaOrgaoIDs(t *testing.T) {
	tests := []struct {
		name              string
		snapshotOrgaoIDs  []string
		cdUAs             []string
		want              []string
	}{
		{
			name:             "cd_ua without snapshot is included (SEIM case)",
			snapshotOrgaoIDs: nil,
			cdUAs:            []string{"5000"},
			want:             []string{"5000"},
		},
		{
			name:             "empty snapshot list still includes all cd_uas",
			snapshotOrgaoIDs: []string{},
			cdUAs:            []string{"5000", "1000"},
			want:             []string{"5000", "1000"},
		},
		{
			name:             "snapshot orgao_id different from cd_ua keeps both",
			snapshotOrgaoIDs: []string{"org-legacy-42"},
			cdUAs:            []string{"5000"},
			want:             []string{"org-legacy-42", "5000"},
		},
		{
			name:             "duplicate when orgao_id equals cd_ua is deduplicated",
			snapshotOrgaoIDs: []string{"5000"},
			cdUAs:            []string{"5000"},
			want:             []string{"5000"},
		},
		{
			name:             "empty inputs yield empty result",
			snapshotOrgaoIDs: nil,
			cdUAs:            nil,
			want:             nil,
		},
		{
			name:             "blank values are skipped",
			snapshotOrgaoIDs: []string{"", "org-1"},
			cdUAs:            []string{"", "5000"},
			want:             []string{"org-1", "5000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeSecretariaOrgaoIDs(tt.snapshotOrgaoIDs, tt.cdUAs)
			assert.Equal(t, tt.want, got)
		})
	}
}
