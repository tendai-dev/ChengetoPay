package vault

import (
	"context"
	"fmt"
	"log"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// VaultConfig holds Vault configuration
type VaultConfig struct {
	Address     string
	Token       string
	MountPath   string
	Timeout     time.Duration
	MaxRetries  int
}

// VaultClient represents the Vault client
type VaultClient struct {
	client   *vault.Client
	config   VaultConfig
}

// Secret represents a secret in Vault
type Secret struct {
	Path    string                 `json:"path"`
	Data    map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Version int                    `json:"version,omitempty"`
}

// DatabaseCredential represents database credentials
type DatabaseCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

// APICredential represents API credentials
type APICredential struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Endpoint  string `json:"endpoint"`
	Version   string `json:"version"`
}

// Certificate represents SSL/TLS certificates
type Certificate struct {
	CertPEM string `json:"cert_pem"`
	KeyPEM  string `json:"key_pem"`
	CAChain string `json:"ca_chain,omitempty"`
}

// NewVaultClient creates a new Vault client
func NewVaultClient(config VaultConfig) (*VaultClient, error) {
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = config.Address
	vaultConfig.Timeout = config.Timeout

	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set token
	client.SetToken(config.Token)

	// Test connection
	_, err = client.Sys().Health()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Vault: %w", err)
	}

	// Enable KV secrets engine if not already enabled
	if err := enableKVSecretsEngine(client, config.MountPath); err != nil {
		return nil, fmt.Errorf("failed to enable KV secrets engine: %w", err)
	}

	vaultClient := &VaultClient{
		client: client,
		config: config,
	}

	log.Println("✅ Vault client connected successfully")
	return vaultClient, nil
}

// enableKVSecretsEngine enables the KV secrets engine
func enableKVSecretsEngine(client *vault.Client, mountPath string) error {
	// Check if KV engine is already enabled
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("failed to list mounts: %w", err)
	}

	if _, exists := mounts[mountPath+"/"]; exists {
		return nil // Already enabled
	}

	// Enable KV v2 secrets engine
	err = client.Sys().Mount(mountPath, &vault.MountInput{
		Type:        "kv",
		Description: "KV Version 2 secret engine",
		Options: map[string]string{
			"version": "2",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to enable KV secrets engine: %w", err)
	}

	return nil
}

// StoreSecret stores a secret in Vault
func (v *VaultClient) StoreSecret(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := v.client.Logical().Write(v.config.MountPath+"/"+path, data)
	if err != nil {
		return fmt.Errorf("failed to store secret at %s: %w", path, err)
	}

	log.Printf("✅ Secret stored at: %s", path)
	return nil
}

// GetSecret retrieves a secret from Vault
func (v *VaultClient) GetSecret(ctx context.Context, path string) (*Secret, error) {
	secret, err := v.client.Logical().Read(v.config.MountPath + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from %s: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at %s", path)
	}

	// Extract data and metadata
	data := make(map[string]interface{})
	metadata := make(map[string]interface{})

	if dataRaw, ok := secret.Data["data"]; ok {
		if dataMap, ok := dataRaw.(map[string]interface{}); ok {
			data = dataMap
		}
	}

	if metadataRaw, ok := secret.Data["metadata"]; ok {
		if metadataMap, ok := metadataRaw.(map[string]interface{}); ok {
			metadata = metadataMap
		}
	}

	return &Secret{
		Path:     path,
		Data:     data,
		Metadata: metadata,
	}, nil
}

// DeleteSecret deletes a secret from Vault
func (v *VaultClient) DeleteSecret(ctx context.Context, path string) error {
	_, err := v.client.Logical().Delete(v.config.MountPath + "/" + path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at %s: %w", path, err)
	}

	log.Printf("✅ Secret deleted at: %s", path)
	return nil
}

// ListSecrets lists all secrets under a path
func (v *VaultClient) ListSecrets(ctx context.Context, path string) ([]string, error) {
	secrets, err := v.client.Logical().List(v.config.MountPath + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at %s: %w", path, err)
	}

	if secrets == nil || secrets.Data == nil {
		return []string{}, nil
	}

	if keysRaw, ok := secrets.Data["keys"]; ok {
		if keys, ok := keysRaw.([]interface{}); ok {
			result := make([]string, len(keys))
			for i, key := range keys {
				if keyStr, ok := key.(string); ok {
					result[i] = keyStr
				}
			}
			return result, nil
		}
	}

	return []string{}, nil
}

// StoreDatabaseCredential stores database credentials
func (v *VaultClient) StoreDatabaseCredential(ctx context.Context, name string, cred DatabaseCredential) error {
	data := map[string]interface{}{
		"username": cred.Username,
		"password": cred.Password,
		"host":     cred.Host,
		"port":     cred.Port,
		"database": cred.Database,
		"ssl_mode": cred.SSLMode,
	}

	return v.StoreSecret(ctx, fmt.Sprintf("database/%s", name), data)
}

// GetDatabaseCredential retrieves database credentials
func (v *VaultClient) GetDatabaseCredential(ctx context.Context, name string) (*DatabaseCredential, error) {
	secret, err := v.GetSecret(ctx, fmt.Sprintf("database/%s", name))
	if err != nil {
		return nil, err
	}

	cred := &DatabaseCredential{}
	if username, ok := secret.Data["username"].(string); ok {
		cred.Username = username
	}
	if password, ok := secret.Data["password"].(string); ok {
		cred.Password = password
	}
	if host, ok := secret.Data["host"].(string); ok {
		cred.Host = host
	}
	if port, ok := secret.Data["port"].(float64); ok {
		cred.Port = int(port)
	}
	if database, ok := secret.Data["database"].(string); ok {
		cred.Database = database
	}
	if sslMode, ok := secret.Data["ssl_mode"].(string); ok {
		cred.SSLMode = sslMode
	}

	return cred, nil
}

// StoreAPICredential stores API credentials
func (v *VaultClient) StoreAPICredential(ctx context.Context, name string, cred APICredential) error {
	data := map[string]interface{}{
		"api_key":    cred.APIKey,
		"api_secret": cred.APISecret,
		"endpoint":   cred.Endpoint,
		"version":    cred.Version,
	}

	return v.StoreSecret(ctx, fmt.Sprintf("api/%s", name), data)
}

// GetAPICredential retrieves API credentials
func (v *VaultClient) GetAPICredential(ctx context.Context, name string) (*APICredential, error) {
	secret, err := v.GetSecret(ctx, fmt.Sprintf("api/%s", name))
	if err != nil {
		return nil, err
	}

	cred := &APICredential{}
	if apiKey, ok := secret.Data["api_key"].(string); ok {
		cred.APIKey = apiKey
	}
	if apiSecret, ok := secret.Data["api_secret"].(string); ok {
		cred.APISecret = apiSecret
	}
	if endpoint, ok := secret.Data["endpoint"].(string); ok {
		cred.Endpoint = endpoint
	}
	if version, ok := secret.Data["version"].(string); ok {
		cred.Version = version
	}

	return cred, nil
}

// StoreCertificate stores SSL/TLS certificates
func (v *VaultClient) StoreCertificate(ctx context.Context, name string, cert Certificate) error {
	data := map[string]interface{}{
		"cert_pem": cert.CertPEM,
		"key_pem":  cert.KeyPEM,
		"ca_chain": cert.CAChain,
	}

	return v.StoreSecret(ctx, fmt.Sprintf("certificates/%s", name), data)
}

// GetCertificate retrieves SSL/TLS certificates
func (v *VaultClient) GetCertificate(ctx context.Context, name string) (*Certificate, error) {
	secret, err := v.GetSecret(ctx, fmt.Sprintf("certificates/%s", name))
	if err != nil {
		return nil, err
	}

	cert := &Certificate{}
	if certPEM, ok := secret.Data["cert_pem"].(string); ok {
		cert.CertPEM = certPEM
	}
	if keyPEM, ok := secret.Data["key_pem"].(string); ok {
		cert.KeyPEM = keyPEM
	}
	if caChain, ok := secret.Data["ca_chain"].(string); ok {
		cert.CAChain = caChain
	}

	return cert, nil
}

// GeneratePassword generates a secure password
func (v *VaultClient) GeneratePassword(ctx context.Context, length int) (string, error) {
	// Use Vault's password generation capability
	data := map[string]interface{}{
		"length": length,
		"charset": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*",
	}

	secret, err := v.client.Logical().Write("sys/tools/random/"+fmt.Sprintf("%d", length), data)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("failed to generate password: no data returned")
	}

	if randomBytes, ok := secret.Data["random_bytes"].(string); ok {
		return randomBytes, nil
	}

	return "", fmt.Errorf("failed to generate password: invalid response format")
}

// RotateSecret rotates a secret by generating a new one
func (v *VaultClient) RotateSecret(ctx context.Context, path string, generator func() (map[string]interface{}, error)) error {
	// Generate new secret
	newData, err := generator()
	if err != nil {
		return fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Store new secret
	if err := v.StoreSecret(ctx, path, newData); err != nil {
		return fmt.Errorf("failed to store new secret: %w", err)
	}

	log.Printf("✅ Secret rotated at: %s", path)
	return nil
}

// GetSecretMetadata gets metadata for a secret
func (v *VaultClient) GetSecretMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := v.GetSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	return secret.Metadata, nil
}

// HealthCheck performs a health check on Vault
func (v *VaultClient) HealthCheck(ctx context.Context) error {
	health, err := v.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	if !health.Initialized {
		return fmt.Errorf("vault is not initialized")
	}

	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}

	return nil
}

// GetVaultStatus returns Vault status information
func (v *VaultClient) GetVaultStatus(ctx context.Context) (map[string]interface{}, error) {
	health, err := v.client.Sys().Health()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault status: %w", err)
	}

	return map[string]interface{}{
		"initialized": health.Initialized,
		"sealed":      health.Sealed,
		"standby":     health.Standby,
		"replication_dr_mode": health.ReplicationDRMode,
		"replication_performance_mode": health.ReplicationPerformanceMode,
		"server_time_utc": health.ServerTimeUTC,
		"version": health.Version,
		"cluster_name": health.ClusterName,
		"cluster_id": health.ClusterID,
	}, nil
}
