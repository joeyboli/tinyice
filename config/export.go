package config

import (
	"encoding/json"
)

// Export returns a JSON snapshot of the config with secrets redacted.
func (c *Config) Export() ([]byte, error) {
	clone, err := c.clone()
	if err != nil {
		return nil, err
	}
	clone.redactSecrets()
	return json.MarshalIndent(clone, "", "    ")
}

func (c *Config) clone() (*Config, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	var out Config
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Config) redactSecrets() {
	c.AdminPassword = RedactedSecret
	for _, u := range c.Users {
		if u != nil {
			u.Password = RedactedSecret
		}
	}
	for _, tok := range c.APITokens {
		tok.TokenHash = RedactedSecret
	}
	for _, p := range c.OIDCProviders {
		p.ClientSecret = RedactedSecret
	}
	for mount := range c.Mounts {
		c.Mounts[mount] = RedactedSecret
	}
	for mount, ms := range c.AdvancedMounts {
		ms.Password = RedactedSecret
		c.AdvancedMounts[mount] = ms
	}
	for _, r := range c.Relays {
		r.Password = RedactedSecret
	}
	for _, adj := range c.AutoDJs {
		adj.MPDPassword = RedactedSecret
	}
	if c.SMTP != nil {
		c.SMTP.Password = RedactedSecret
	}
}

type secretSnapshot struct {
	adminPassword string
	mounts        map[string]string
	advanced      map[string]string
	users         map[string]string
	tokens        map[string]string
	oidc          map[string]string
	smtpPassword  string
	relayPassword map[string]string
	mpdPassword   map[string]string
}

func (c *Config) snapshotSecrets() secretSnapshot {
	s := secretSnapshot{
		mounts:        map[string]string{},
		advanced:      map[string]string{},
		users:         map[string]string{},
		tokens:        map[string]string{},
		oidc:          map[string]string{},
		relayPassword: map[string]string{},
		mpdPassword:   map[string]string{},
		adminPassword: c.AdminPassword,
	}
	for k, v := range c.Mounts {
		s.mounts[k] = v
	}
	for k, v := range c.AdvancedMounts {
		s.advanced[k] = v.Password
	}
	for k, u := range c.Users {
		if u != nil {
			s.users[k] = u.Password
		}
	}
	for _, t := range c.APITokens {
		s.tokens[t.ID] = t.TokenHash
	}
	for _, p := range c.OIDCProviders {
		s.oidc[p.ID] = p.ClientSecret
	}
	if c.SMTP != nil {
		s.smtpPassword = c.SMTP.Password
	}
	for _, r := range c.Relays {
		key := r.URL + "|" + r.Mount
		s.relayPassword[key] = r.Password
	}
	for _, adj := range c.AutoDJs {
		s.mpdPassword[adj.Mount] = adj.MPDPassword
	}
	return s
}

func (c *Config) restoreSecrets(s secretSnapshot) {
	if isRedactedSecret(c.AdminPassword) {
		c.AdminPassword = s.adminPassword
	}
	for mount, pw := range c.Mounts {
		if isRedactedSecret(pw) {
			if old, ok := s.mounts[mount]; ok {
				c.Mounts[mount] = old
			}
		}
	}
	for mount, ms := range c.AdvancedMounts {
		if isRedactedSecret(ms.Password) {
			if old, ok := s.advanced[mount]; ok {
				ms.Password = old
				c.AdvancedMounts[mount] = ms
			}
		}
	}
	for name, u := range c.Users {
		if u != nil && isRedactedSecret(u.Password) {
			if old, ok := s.users[name]; ok {
				u.Password = old
			}
		}
	}
	for _, tok := range c.APITokens {
		if isRedactedSecret(tok.TokenHash) {
			if old, ok := s.tokens[tok.ID]; ok {
				tok.TokenHash = old
			}
		}
	}
	for _, p := range c.OIDCProviders {
		if isRedactedSecret(p.ClientSecret) {
			if old, ok := s.oidc[p.ID]; ok {
				p.ClientSecret = old
			}
		}
	}
	if c.SMTP != nil && isRedactedSecret(c.SMTP.Password) {
		c.SMTP.Password = s.smtpPassword
	}
	for _, r := range c.Relays {
		key := r.URL + "|" + r.Mount
		if isRedactedSecret(r.Password) {
			if old, ok := s.relayPassword[key]; ok {
				r.Password = old
			}
		}
	}
	for _, adj := range c.AutoDJs {
		if isRedactedSecret(adj.MPDPassword) {
			if old, ok := s.mpdPassword[adj.Mount]; ok {
				adj.MPDPassword = old
			}
		}
	}
}

// MergeImport overlays incoming config onto the current one, preserving
// existing secrets when the import carries empty or redacted values.
func (c *Config) MergeImport(in *Config) error {
	if in == nil {
		return nil
	}
	secrets := c.snapshotSecrets()
	path := c.ConfigPath

	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	var merged Config
	if err := json.Unmarshal(raw, &merged); err != nil {
		return err
	}
	merged.ConfigPath = path
	*c = merged
	c.ConfigPath = path
	c.restoreSecrets(secrets)
	c.setDefaults()
	return nil
}
