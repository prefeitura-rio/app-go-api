package models

type CertificateUpdateRequest struct {
	CertificateURL string `json:"certificate_url" binding:"omitempty,url,max=2048"`
}

type CertificateUpdateResponse struct {
	Message        string `json:"message"`
	CertificateURL string `json:"certificate_url"`
}
