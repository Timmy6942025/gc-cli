package auth

// Default OAuth client for seamless public installs.
// Overrides:
// - GC_OAUTH_CLIENT_ID
// - GC_OAUTH_CLIENT_SECRET
const (
	DefaultOAuthClientID = "597878429548-hhf8isfl206qekrt51dlsiejnlblmk8g.apps.googleusercontent.com"
)

// Inject this at build time with:
//
//	-ldflags="-X github.com/timothy/gc-cli/internal/auth.DefaultOAuthClientSecret=..."
var DefaultOAuthClientSecret = ""
