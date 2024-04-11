package train

import (
	v1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
)

// token issue response body
type TokenResponse struct {
	Token Token `json:"token"`
}

type Token struct {
	Methods   []string  `json:"methods"`
	User      User      `json:"user"`
	AuditIds  []string  `json:"audit_ids"`
	ExpiresAt string    `json:"expires_at"`
	IssuedAt  string    `json:"issued_at"`
	Domain    Domain    `json:"domain"`
	Roles     []Role    `json:"roles"`
	Catalog   []Catalog `json:"catalog"`
}

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Catalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
}

type Endpoint struct {
	ID        string `json:"id"`
	Interface string `json:"interface"`
	RegionID  string `json:"region_id"`
	URL       string `json:"url"`
	Region    string `json:"region"`
}

// keystone token issue request body
type AuthRequest struct {
	Auth AuthDetails `json:"auth"`
}

type AuthDetails struct {
	Identity IdentityDetails `json:"identity"`
	Scope    *ScopeDetails   `json:"scope,omitempty"`
}

type IdentityDetails struct {
	Methods  []string        `json:"methods"`
	Password PasswordDetails `json:"password"`
}

type PasswordDetails struct {
	User User `json:"user"`
}

type Domain struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ScopeDetails struct {
	Domain DomainDetails `json:"domain"`
}

type DomainDetails struct {
	Id string `json:"id"`
}

type User struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	DomainID          string            `json:"domain_id,omitempty"`
	ProjectID         string            `json:"project_id,omitempty"`
	Enabled           bool              `json:"enabled,omitempty"`
	PasswordExpiresAt string            `json:"password_expires_at,omitempty"`
	Password          string            `json:"password,omitempty"`
	Domain            *Domain           `json:"domain,omitempty"`
	Options           Options           `json:"options,omitempty"`
	Links             Links             `json:"links,omitempty"`
	OpenhydraUser     *v1.OpenHydraUser `json:"openhydra,omitempty"`
	Email             string            `json:"email,omitempty"`
}

type Options struct {
	IgnoreChangePasswordUponFirstUse bool `json:"ignore_change_password_upon_first_use,omitempty"`
	IgnorePasswordExpiry             bool `json:"ignore_password_expiry,omitempty"`
	IgnoreLockoutFailureAttempts     bool `json:"ignore_lockout_failure_attempts,omitempty"`
	LockPassword                     bool `json:"lock_password,omitempty"`
}

type Links struct {
	Self string `json:"self"`
}

type UserContainer struct {
	Users []User `json:"users"`
}
