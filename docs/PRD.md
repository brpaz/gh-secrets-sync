# PRD

## Motivation

I faced a few situations where I have common secrets that I want to deploy and keep in sync across multiple repositories. For example, GitHub Apps bot tokens, or tokens to interact with external services like NPM. When having a few repos that requires the same token, having to manually set it up in each repository is a pain, and it's easy to forget to update it when the token rotates.

## Solution

A Github CLI extension that syncs GitHub secrets across different repositories. The extension will allow users to manage their secrets in a local configuration file and then propagate those secrets to all repositories that are using them with a simple command.

## Secrets storage

The secrets will be stored in a local configuration file located at `~/.config/gh-secrets-sync/secrets.yaml` for linux. It should find the config directory standard for the OS. The configuration file will have the following structure:


```yaml
secrets:
  - name: "SECRET_NAME"
    value: "SECRET_VALUE"
    repositories:
      - "owner/repo1"
      - "owner/repo2"    
```

This file should be created by default the first time this extension is run, if it doesn't exist. Users can then edit this file to add, update, or delete secrets as needed.

## Available Commands

The extension should provide the following commands to manage secrets:

`gh secrets-sync` - 
`gh secrets-sync add` - Add a new secret to the configuration file.
`gh secrets-sync update` - Update an existing secret in the configuration file.
`gh secrets-sync delete` - Delete a secret from the configuration file.
`gh secrets-sync sync` - Sync the secrets from the configuration file to all repositories that are using them.
`gh secrets-sync list` - List all secrets in the configuration file and the repositories they are associated with.
`gh secrets-sync config` - Open the secrets config in the configured $EDITOR for manual editing.

## Tech Stack

- Golang
- GitHub CLI framework for building extensions - github.com/cli/go-gh/v2
- Urfave/cli for command line parsing - github.com/urfave/cli/v3
