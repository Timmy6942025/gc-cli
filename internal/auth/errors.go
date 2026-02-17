package auth

import "fmt"

var ErrNoToken = fmt.Errorf("no saved OAuth token found")
var ErrMissingClientID = fmt.Errorf("missing OAuth client id; set GC_OAUTH_CLIENT_ID or config oauth_client_id")

type ScopesRequiredError struct {
	Missing []string
}

func (e ScopesRequiredError) Error() string {
	return fmt.Sprintf("missing OAuth scopes: %v", e.Missing)
}
