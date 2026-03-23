package models

type CertificateUpdateRequest struct {
	CertificateURL string `json:"certificate_url" binding:"omitempty"`
}

type CertificateUpdateResponse struct {
	Message        string `json:"message"`
	CertificateURL string `json:"certificate_url"`
}
