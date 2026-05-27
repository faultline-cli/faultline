package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"faultline/internal/teams"
)

const maxSyncArtifactBytes = 5 * 1024 * 1024

// SyncOptions controls a Team artifact sync operation.
type SyncOptions struct {
	ProjectID string
	Source    string
	Branch    string
	CommitSHA string
}

type syncCredentials struct {
	Token    string
	TeamSlug string
	APIURL   string
}

// Sync pushes the artifact from `faultline analyze --json` to Faultline Teams.
func (Service) Sync(ctx context.Context, r io.Reader, opts SyncOptions, w io.Writer) error {
	projectID := strings.TrimSpace(firstNonEmptySync(opts.ProjectID, os.Getenv("FAULTLINE_PROJECT")))
	if projectID == "" {
		return fmt.Errorf("--project is required (or set FAULTLINE_PROJECT)")
	}

	creds, err := resolveSyncCredentials()
	if err != nil {
		return err
	}

	data, err := readSyncArtifactInput(r)
	if err != nil {
		return err
	}
	artifact, err := extractSyncArtifact(data)
	if err != nil {
		return err
	}

	req := teams.SyncRequest{
		ProjectID: projectID,
		Source:    opts.Source,
		Branch:    opts.Branch,
		CommitSHA: opts.CommitSHA,
		Artifact:  artifact,
	}
	result, err := teams.NewClient(creds.APIURL).Sync(ctx, creds.Token, creds.TeamSlug, req)
	if err != nil {
		return err
	}

	if result.IsNew {
		_, err = fmt.Fprintf(w, "synced  new        %s\n", result.ID)
	} else {
		_, err = fmt.Fprintf(w, "synced  duplicate  %s\n", result.ID)
	}
	return err
}

func resolveSyncCredentials() (syncCredentials, error) {
	apiURL := firstNonEmptySync(os.Getenv("FAULTLINE_API_URL"), teams.DefaultAPIURL)
	token := os.Getenv("FAULTLINE_TOKEN")
	if token != "" {
		if !strings.HasPrefix(token, "ft_") {
			return syncCredentials{}, fmt.Errorf("invalid token: must start with 'ft_'")
		}
		return syncCredentials{Token: token, APIURL: apiURL}, nil
	}

	creds, err := teams.LoadCredentials()
	if err != nil {
		return syncCredentials{}, fmt.Errorf("load credentials: %w", err)
	}
	if creds == nil || creds.Token == "" {
		return syncCredentials{}, fmt.Errorf("not authenticated; run 'faultline auth login' or set FAULTLINE_TOKEN")
	}
	if !strings.HasPrefix(creds.Token, "ft_") {
		return syncCredentials{}, fmt.Errorf("invalid stored token: must start with 'ft_'")
	}
	if creds.APIURL != "" && os.Getenv("FAULTLINE_API_URL") == "" {
		apiURL = creds.APIURL
	}
	return syncCredentials{
		Token:    creds.Token,
		TeamSlug: creds.TeamSlug,
		APIURL:   apiURL,
	}, nil
}

func firstNonEmptySync(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func readSyncArtifactInput(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxSyncArtifactBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read artifact: %w", err)
	}
	if len(data) > maxSyncArtifactBytes {
		return nil, fmt.Errorf("artifact input exceeds maximum size")
	}
	return data, nil
}

func extractSyncArtifact(data []byte) (json.RawMessage, error) {
	var wrapper struct {
		Matched  *bool           `json:"matched"`
		Artifact json.RawMessage `json:"artifact"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse artifact JSON: %w", err)
	}
	if len(wrapper.Artifact) == 0 {
		return nil, fmt.Errorf("no artifact field in JSON; re-run 'faultline analyze --json' to regenerate")
	}
	if bytes.Equal(bytes.TrimSpace(wrapper.Artifact), []byte("null")) {
		if wrapper.Matched != nil && !*wrapper.Matched {
			return nil, fmt.Errorf("no artifact to sync: the log was not matched by any playbook")
		}
		return nil, fmt.Errorf("no artifact field in JSON; re-run 'faultline analyze --json' to regenerate")
	}

	var artifact struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.Unmarshal(wrapper.Artifact, &artifact); err != nil {
		return nil, fmt.Errorf("artifact must be a JSON object: %w", err)
	}
	if strings.TrimSpace(artifact.Fingerprint) == "" {
		return nil, errors.New("artifact is missing fingerprint; re-run 'faultline analyze --json' to regenerate")
	}
	return wrapper.Artifact, nil
}
