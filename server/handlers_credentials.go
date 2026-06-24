package server

import (
	"encoding/json"
	"net/http"

	"github.com/DatanoiseTV/tinyice/config"
)

// apiRegenerateCredential rotates a source or relay password and returns the
// new plaintext value once (passwords are only stored hashed for mounts).
func (s *Server) apiRegenerateCredential(w http.ResponseWriter, r *http.Request) {
	if !s.isCSRFSafe(r) {
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.checkAuth(r)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Target string `json:"target"` // default_source, mount, relay
		Mount  string `json:"mount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	plain, status, err := s.regenerateCredential(user, body.Target, body.Mount)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}

	s.Audit(r, "credential_regenerated", "credentials", body.Target+body.Mount, body.Target)

	jsonResponse(w, map[string]string{
		"password": plain,
		"target":   body.Target,
		"mount":    body.Mount,
	})
}

func (s *Server) regenerateCredential(user *config.User, target, mount string) (string, int, error) {
	plain := generateRandomString(12)

	switch target {
	case "default_source":
		if user.Role != config.RoleSuperAdmin {
			return "", http.StatusForbidden, errCred("superadmin required")
		}
		hashed, err := config.HashPassword(plain)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
		s.Config.DefaultSourcePassword = hashed
		if err := s.Config.SaveConfig(); err != nil {
			return "", http.StatusInternalServerError, err
		}
		return plain, http.StatusOK, nil

	case "mount":
		if mount == "" {
			return "", http.StatusBadRequest, errCred("mount is required")
		}
		if mount[0] != '/' {
			mount = "/" + mount
		}
		if !s.mountExists(mount) {
			return "", http.StatusNotFound, errCred("mount not found")
		}
		if !s.hasAccess(user, mount) {
			return "", http.StatusForbidden, errCred("forbidden")
		}
		hashed, err := config.HashPassword(plain)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
		if _, ok := s.Config.Mounts[mount]; ok {
			if user.Role != config.RoleSuperAdmin {
				return "", http.StatusForbidden, errCred("forbidden")
			}
			s.Config.Mounts[mount] = hashed
		} else {
			found := false
			for _, u := range s.Config.Users {
				if _, ok := u.Mounts[mount]; ok {
					if user.Role != config.RoleSuperAdmin && u.Username != user.Username {
						return "", http.StatusForbidden, errCred("forbidden")
					}
					u.Mounts[mount] = hashed
					found = true
					break
				}
			}
			if !found && user.Role != config.RoleSuperAdmin {
				return "", http.StatusNotFound, errCred("mount not found")
			}
		}
		if adv, ok := s.Config.AdvancedMounts[mount]; ok {
			adv.Password = hashed
			s.Config.AdvancedMounts[mount] = adv
		}
		if err := s.Config.SaveConfig(); err != nil {
			return "", http.StatusInternalServerError, err
		}
		return plain, http.StatusOK, nil

	case "relay":
		if user.Role != config.RoleSuperAdmin {
			return "", http.StatusForbidden, errCred("superadmin required")
		}
		if mount == "" {
			return "", http.StatusBadRequest, errCred("mount is required")
		}
		if mount[0] != '/' {
			mount = "/" + mount
		}
		var rc *config.RelayConfig
		for _, r := range s.Config.Relays {
			if r.Mount == mount {
				rc = r
				break
			}
		}
		if rc == nil {
			return "", http.StatusNotFound, errCred("relay not found")
		}
		rc.Password = plain
		if err := s.Config.SaveConfig(); err != nil {
			return "", http.StatusInternalServerError, err
		}
		s.RelayM.StopRelay(mount)
		s.RelayM.StartRelay(rc.URL, rc.Mount, plain, rc.BurstSize, s.Config.VisibleMounts[mount])
		return plain, http.StatusOK, nil

	default:
		return "", http.StatusBadRequest, errCred("target must be default_source, mount, or relay")
	}
}

type credError string

func (e credError) Error() string { return string(e) }

func errCred(msg string) error { return credError(msg) }

func (s *Server) mountExists(mount string) bool {
	if _, ok := s.Config.Mounts[mount]; ok {
		return true
	}
	if _, ok := s.Config.AdvancedMounts[mount]; ok {
		return true
	}
	for _, u := range s.Config.Users {
		if _, ok := u.Mounts[mount]; ok {
			return true
		}
	}
	return false
}

// generateMountPassword creates a hashed mount password, optionally using a
// caller-supplied plaintext. When plain is empty a random password is generated
// and returned.
func generateMountPassword(plain string) (hashed, generated string, err error) {
	if plain == "" {
		generated = generateRandomString(12)
		plain = generated
	}
	hashed, err = config.HashPassword(plain)
	return hashed, generated, err
}
