# Contributing to Shugur Relay

First off, thank you for considering contributing to Shugur Relay! It's people like you that make the open source community such a great place. We welcome any form of contribution, from reporting bugs and suggesting features to writing code and improving documentation.

## Ways to Contribute

- **Reporting Bugs**: If you find a bug, please open an issue on GitHub. Include as much detail as possible, such as steps to reproduce, expected behavior, and actual behavior.
- **Suggesting Enhancements**: If you have an idea for a new feature or an improvement to an existing one, open an issue to discuss it.
- **Pull Requests**: If you're ready to contribute code or documentation, we'd love to see your pull requests.

## Getting Started

1. **Fork the repository** on GitHub.
2. **Clone your fork** to your local machine:

    ```bash
    git clone https://github.com/your-username/Relay.git
    cd Relay
    ```

3. **Create a new branch** for your changes:

    ```bash
    git checkout -b feature/your-awesome-feature
    ```

4. **Make your changes**. Be sure to follow the coding style and add tests where appropriate.
5. **Commit your changes**:

    ```bash
    git commit -m "feat: Add your awesome feature"
    ```

6. **Push to your branch**:

    ```bash
    git push origin feature/your-awesome-feature
    ```

7. **Open a pull request** from your fork to the `main` branch of the Shugur Relay repository.

## Pull Request Guidelines

- **Keep it focused**: Each pull request should address a single issue or feature.
- **Write clear commit messages**: Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.
- **Update documentation**: If you're adding a new feature or changing an existing one, be sure to update the documentation.
- **Add tests**: All new features should have corresponding tests.
- **Ensure tests pass**: Run `go test ./...` before submitting your pull request.
- **Link to an issue**: If your pull request addresses an existing issue, be sure to link to it in the description (e.g., `Fixes #123`).

## Development Setup

> **💻 Development Tip**: Use a consistent development environment. Consider using Docker for database dependencies to avoid conflicts between different projects.

### Prerequisites

- Go 1.21 or later
- CockroachDB (for local development)
- Docker and Docker Compose (optional)

### Local Development

1. **Set up CockroachDB locally**:

   ```bash
   # Using Docker
   docker run -d --name cockroach -p 26257:26257 cockroachdb/cockroach:latest start-single-node --insecure
   ```

2. **Configure the relay**:

   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your database settings
   ```

3. **Run the relay**:

   ```bash
   go run ./cmd
   ```

## Testing

Run the test suite before submitting changes:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific NIP tests
cd tests/nips && ./test_nip01.sh
```

## Code Style

We follow the standard Go coding style. Use `gofmt` to format your code before committing.

Additional guidelines:

- Write clear, descriptive variable and function names
- Add comments for complex logic
- Keep functions small and focused
- Follow the existing project structure

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Help

If you have any questions:

- Open an issue for bug reports or feature requests
- Check our [documentation](https://github.com/Shugur-Network/docs) for detailed guides
- Review existing issues and pull requests
