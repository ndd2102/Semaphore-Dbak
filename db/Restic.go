package db

import (
    "errors"
)

// ResticConfig stores configuration information for Restic
type ResticConfig struct {
    ID        int    `db:"id" json:"id"`
    ProjectID int    `db:"project_id" json:"project_id"`
    Name      string `db:"name" json:"name" binding:"required"`
    URL       string `db:"url" json:"url" binding:"required"`
    Bucket    string `db:"bucket" json:"bucket" binding:"required"`
    SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required" backup:"-"`
    ResticKey string `db:"restic_key" json:"restic_key" binding:"required"`

	SSHKey AccessKey `db:"-" json:"-" backup:"-"`
}

// Validate checks if the ResticConfig fields are valid
func (m ResticConfig) Validate() error {
    if m.Name == "" {
        return errors.New("Restic configuration name can't be empty")
    }
    if m.URL == "" {
        return errors.New("Restic URL can't be empty")
    }
    if m.Bucket == "" {
        return errors.New("Restic Bucket can't be empty")
    }
    if m.ResticKey == "" {
        return errors.New("Restic Secret Key can't be empty")
    }
    return nil
}
