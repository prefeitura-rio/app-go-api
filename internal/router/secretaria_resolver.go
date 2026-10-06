package router

// mergeSecretariaOrgaoIDs unions orgao_ids resolved from orgao_snapshots with the
// raw cd_ua values from the CPF-Secretaria vínculo.
//
// Snapshot IDs cover legacy courses whose orgao_id is not the cd_ua.
// cd_uas cover secretarias that have no snapshot yet (e.g. SEIM / 5000 before
// any curso, emprego, or MEI opportunity references them), so editors linked
// only to those organs still get a non-empty scope for CourseAuthorization.
func mergeSecretariaOrgaoIDs(snapshotOrgaoIDs, cdUAs []string) []string {
	seen := make(map[string]struct{}, len(snapshotOrgaoIDs)+len(cdUAs))
	var result []string

	for _, id := range snapshotOrgaoIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	for _, cdUA := range cdUAs {
		if cdUA == "" {
			continue
		}
		if _, ok := seen[cdUA]; ok {
			continue
		}
		seen[cdUA] = struct{}{}
		result = append(result, cdUA)
	}

	return result
}
