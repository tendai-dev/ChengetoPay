package main

import (
	"context"
	"fmt"
	"time"
)

// Organization represents a tenant/organization
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Status      string    `json:"status"` // active, inactive, suspended
	Plan        string    `json:"plan"`   // basic, pro, enterprise
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Status       string    `json:"status"` // active, inactive, suspended
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Role represents a role with permissions
type Role struct {
	ID          string   `json:"id"`
	OrgID       string   `json:"org_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole represents a user's role assignment
type UserRole struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
	OrgID  string `json:"org_id"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	KeyHash   string    `json:"key_hash"`
	Scopes    []string  `json:"scopes"`
	Status    string    `json:"status"` // active, inactive, revoked
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Token represents an OAuth/JWT token
type Token struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	UserID       string    `json:"user_id"`
	TokenType    string    `json:"token_type"` // access, refresh
	TokenHash    string    `json:"token_hash"`
	Scopes       []string  `json:"scopes"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthenticateRequest represents an authentication request
type AuthenticateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgID    string `json:"org_id,omitempty"`
}

// TokenRequest represents a token generation request
type TokenRequest struct {
	GrantType    string `json:"grant_type"` // password, client_credentials, refresh_token
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
}

// CreateOrgRequest represents a request to create an organization
type CreateOrgRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Plan   string `json:"plan"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	OrgID     string `json:"org_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// CheckPermissionRequest represents a permission check request
type CheckPermissionRequest struct {
	UserID     string `json:"user_id"`
	OrgID      string `json:"org_id"`
	Permission string `json:"permission"`
	Resource   string `json:"resource,omitempty"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// OrgFilters represents filters for listing organizations
type OrgFilters struct {
	Status string `json:"status"`
	Plan   string `json:"plan"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// UserFilters represents filters for listing users
type UserFilters struct {
	OrgID  string `json:"org_id"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// RoleFilters represents filters for listing roles
type RoleFilters struct {
	OrgID string `json:"org_id"`
	Limit int    `json:"limit"`
	Offset int   `json:"offset"`
}

// Repository interface for auth data access
type Repository interface {
	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, id string) (*Organization, error)
	ListOrganizations(ctx context.Context, filters OrgFilters) ([]*Organization, error)
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context, filters UserFilters) ([]*User, error)
	GetUserByEmail(ctx context.Context, email, orgID string) (*User, error)
	CreateRole(ctx context.Context, role *Role) error
	ListRoles(ctx context.Context, filters RoleFilters) ([]*Role, error)
	AssignUserRole(ctx context.Context, userRole *UserRole) error
	GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
	CreateAPIKey(ctx context.Context, apiKey *APIKey) error
	GetAPIKey(ctx context.Context, keyHash string) (*APIKey, error)
	CreateToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, tokenHash string) (*Token, error)
	ValidatePassword(ctx context.Context, userID, password string) (bool, error)
}

// Service represents the auth business logic
type Service struct {
	repo   Repository
	logger interface{}
}

// NewService creates a new auth service
func NewService(repo Repository, logger interface{}) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Authenticate authenticates a user
func (s *Service) Authenticate(ctx context.Context, req *AuthenticateRequest) (*Token, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email, req.OrgID)
	if err != nil {
		return nil, err
	}

	valid, err := s.repo.ValidatePassword(ctx, user.ID, req.Password)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate token
	token := &Token{
		ID:        generateID(),
		OrgID:     user.OrgID,
		UserID:    user.ID,
		TokenType: "access",
		TokenHash: generateTokenHash(),
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// GenerateToken generates a new token
func (s *Service) GenerateToken(ctx context.Context, req *TokenRequest) (*Token, error) {
	// Implementation would handle different grant types
	token := &Token{
		ID:        generateID(),
		TokenType: "access",
		TokenHash: generateTokenHash(),
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateToken validates a token
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*Token, error) {
	return s.repo.GetToken(ctx, tokenString)
}

// CreateOrganization creates a new organization
func (s *Service) CreateOrganization(ctx context.Context, req *CreateOrgRequest) (*Organization, error) {
	org := &Organization{
		ID:        generateID(),
		Name:      req.Name,
		Domain:    req.Domain,
		Status:    "active",
		Plan:      req.Plan,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// ListOrganizations lists organizations with filters
func (s *Service) ListOrganizations(ctx context.Context, filters OrgFilters) ([]*Organization, error) {
	return s.repo.ListOrganizations(ctx, filters)
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	user := &User{
		ID:        generateID(),
		OrgID:     req.OrgID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ListUsers lists users with filters
func (s *Service) ListUsers(ctx context.Context, filters UserFilters) ([]*User, error) {
	return s.repo.ListUsers(ctx, filters)
}

// ListRoles lists roles with filters
func (s *Service) ListRoles(ctx context.Context, filters RoleFilters) ([]*Role, error) {
	return s.repo.ListRoles(ctx, filters)
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*PermissionResult, error) {
	roles, err := s.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission == req.Permission {
				return &PermissionResult{Allowed: true}, nil
			}
		}
	}

	return &PermissionResult{Allowed: false, Reason: "Permission not found"}, nil
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *Organization) error {
	return nil
}

func (m *MockRepository) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	return &Organization{
		ID:        id,
		Name:      "Test Organization",
		Domain:    "test.com",
		Status:    "active",
		Plan:      "pro",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListOrganizations(ctx context.Context, filters OrgFilters) ([]*Organization, error) {
	return []*Organization{}, nil
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	return nil
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*User, error) {
	return &User{
		ID:        id,
		OrgID:     "org_123",
		Email:     "user@test.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ListUsers(ctx context.Context, filters UserFilters) ([]*User, error) {
	return []*User{}, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email, orgID string) (*User, error) {
	return &User{
		ID:        "user_123",
		OrgID:     orgID,
		Email:     email,
		FirstName: "John",
		LastName:  "Doe",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) CreateRole(ctx context.Context, role *Role) error {
	return nil
}

func (m *MockRepository) ListRoles(ctx context.Context, filters RoleFilters) ([]*Role, error) {
	return []*Role{}, nil
}

func (m *MockRepository) AssignUserRole(ctx context.Context, userRole *UserRole) error {
	return nil
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	return []*Role{}, nil
}

func (m *MockRepository) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	return nil
}

func (m *MockRepository) GetAPIKey(ctx context.Context, keyHash string) (*APIKey, error) {
	return nil, nil
}

func (m *MockRepository) CreateToken(ctx context.Context, token *Token) error {
	return nil
}

func (m *MockRepository) GetToken(ctx context.Context, tokenHash string) (*Token, error) {
	return &Token{
		ID:        "token_123",
		OrgID:     "org_123",
		UserID:    "user_123",
		TokenType: "access",
		TokenHash: tokenHash,
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRepository) ValidatePassword(ctx context.Context, userID, password string) (bool, error) {
	return true, nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("auth_%d", time.Now().UnixNano())
}

func generateTokenHash() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}
