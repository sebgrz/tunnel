package configuration

type Configuration struct {
	Certificates []Certificate `json:"certificates"`
}

type Certificate struct {
	CertPath    string `json:"cert_path"`
	CertKeyPath string `json:"cert_key_path"`
}
