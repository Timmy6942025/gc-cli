# OAuth Verification Checklist (Public App)

## Required artifacts

- OAuth consent screen configured in Google Cloud project
- App domain verification complete
- Privacy policy URL and Terms URL published
- App logo and support email set
- Scope justification document for sensitive Classroom/Drive scopes
- Data usage, storage, deletion policy documentation

## Implementation alignment in this repo

- PKCE + state validation implemented
- Token storage in system keychain
- Progressive scope acquisition supported
- Logout revokes local credentials (token deletion)
- JSON errors include actionable hints and optional handoff URLs

## Release readiness checks

- [ ] Consent screen in production mode
- [ ] Test users removed/adjusted for public launch
- [ ] Privacy policy reviewed and accessible
- [ ] Terms reviewed and accessible
- [ ] Scope list minimized to required operations
- [ ] Security review of logs for token leakage
- [ ] Pen-test / abuse-case review complete
