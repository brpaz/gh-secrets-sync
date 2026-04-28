# gh-secrets-sync

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/brpaz/gh-secrets-sync?style=for-the-badge)
![Go Report Card](https://goreportcard.com/badge/github.com/brpaz/gh-secrets-sync?style=for-the-badge)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/brpaz/brpaz/gh-secrets-sync/ci.yml?branch=main&style=for-the-badge)](https://github.com/brpaz/gh-secrets-sync/actions)
![License](https://img.shields.io/github/license/brpaz/gh-secrets-sync?style=for-the-badge)](./LICENSE)

> Github CLI extension that syncs GitHub secrets across different repositories.

## 🎯 Motivation

I faced a few situations where I have common secrets that I want to deploy and keep in sync across multiple repositories. For example, GitHub Apps bot tokens, or tokens to interact with external services like NPM. When having a few repos that requires the same token, having to manually set it up in each repository is a pain, and it's easy to forget to update it when the token rotates. 

That´s why I decided to build this tool, to simplify this process and have a centralized source of truth that can be easily updated and propagated to all repositories that need it.

## 🗃️ Features

- Secrets are stored in a local configuration file.
- Commands to add, update, and delete secrets from the configuration file, simplify the management process.
- Sync command to propagate changes to all repositories that are using the secrets, ensuring that all repositories are always up-to-date with the latest secrets.

## 🚀 Getting Started

### Installation

You can install `gh-secrets-sync` using the GitHub CLI:

```bash
gh extension install brpaz/gh-secrets-sync
```

## Usage

After installing the extension, you can start it using `gh secrets-sync` command. At startup, the extension will create a configuration file at `~/.config/gh-secrets-sync/secrets.yaml` if it doesn't exist. This is the file were you will place your secrets.

The configuration file should have the following structure:

```yaml
secrets:
  - name: "SECRET_NAME"
    value: "SECRET_VALUE"
    target_name: "SECRET_NAME_IN_GITHUB"
    repositories:
      - "owner/repo1"
      - "owner/repo2"    
    
## 🤝 Contributing

All contributions are welcome. Please check [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 🫶 Support

If you find this project helpful and would like to support its development, there are a few ways you can contribute:

[![Sponsor me on GitHub](https://img.shields.io/badge/Sponsor-%E2%9D%A4-%23db61a2.svg?&logo=github&logoColor=red&&style=for-the-badge&labelColor=white)](https://github.com/sponsors/brpaz)

<a href="https://www.buymeacoffee.com/Z1Bu6asGV" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>

## 👱 Contributors

- [Bruno Paz](https://brunopaz.dev) - Creator and maintainer

## ❤️ Acknowledgements

## 📃 License

Distributed under the MIT License. See [LICENSE](LICENSE) file for details.

## 📩 Contact

- 📧 **Email**: [oss@brunopaz.dev](mailto:oss@brunopaz.dev)
- 🐞 **Issues**: [GitHub Issues](https://github.com/brpaz/gh-secrets-sync/issues)
- 🖇️ **Source**: [https://github.com/brpaz/gh-secrets-sync](https://github.com/brpaz/gh-secrets-sync)
