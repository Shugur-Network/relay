# GitHub Secrets Configuration

This document outlines the required GitHub secrets for the CI/CD workflows to function properly.

## Required Secrets

### Core Secrets

- **`GITHUB_TOKEN`** (Automatically provided)
  - Used for: Basic GitHub API operations, downloading artifacts, Docker registry authentication
  - Permissions: Automatically managed by GitHub

### Optional Secrets

- **`RELEASE_PLEASE_TOKEN`** (Optional)
  - Used for: Enhanced release-please functionality
  - Fallback: Falls back to `GITHUB_TOKEN` if not set
  - Required permissions: `contents: write`, `pull-requests: write`
  - Setup: Generate a Personal Access Token (classic) or Fine-grained Personal Access Token

## Setting Up Secrets

### Repository Secrets
1. Go to your repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add the secret name and value

### Organization Secrets (Optional)
For organization-wide secrets, configure at the organization level.

## Security Best Practices

- Regularly rotate tokens
- Use minimal required permissions
- Monitor secret usage in workflow runs
- Use environment-specific secrets when possible

## Troubleshooting

If you encounter permission errors:
1. Verify the secret exists and is spelled correctly
2. Check that the token has the required permissions
3. Verify the token hasn't expired
4. For PATs, ensure the user has appropriate repository access