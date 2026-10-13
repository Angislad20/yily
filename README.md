<p align="center">
  <picture>
    <img src="logo.svg" alt="Overmind" />
  </picture>
</p>

**Yily** is a lightweight and fast **open-source secret management application** designed to help you securely store and manage sensitive information such as API keys, passwords, tokens, and other secrets.

The project was born from a simple need: having a **self-hosted secret management solution that is easy to deploy, simple to use, and efficient**, without the complexity that can come with more advanced solutions such as HashiCorp Vault or Infisical.

Yily aims to provide a good balance between **simplicity, performance, security, and self-hosting**. It is designed for developers and teams who want to keep control of their secrets and infrastructure while using a straightforward and lightweight solution.

The project is still in its **early stages of development**, and the goal is to build it together with the open-source community. Contributions, ideas, feedback, and discussions are very welcome.

If you are interested in contributing, please check out the [contributing guidelines](CONTRIBUTING.MD) and feel free to open an issue or submit a pull request.

## Technologies Used

| Technology | Description |
|------------|-------------|
| Go | The core programming language used for building Yily server and CLI. |
| Echo | A web framework for Go, used for building the Yily web interface. |
| SQLite | A lightweight relational database used for storing secrets and metadata. |
| SolidJS | A reactive JavaScript library used for building the Yily web interface. |
| Tailwind CSS | A utility-first CSS framework used for styling the Yily web interface. |
| Nix | A package manager and build system used for building and packaging Yily. |

## Project Structure

For this project, we have a monorepo structure that contains both the server and the web interface. The main directories are:

```text
.
├── client/                 # Interface web
└── server/                 # API, CLI et services backend
```

The `client` contains the web interface, while the `server` contains the API, the command-line interface, and the application's internal packages.

## Configuration

### Master Key (`YILY_MASTER_KEY`)

Yily encrypts secret values at rest using AES-256-GCM. The server requires a 32-byte Base64-encoded master key provided through the `YILY_MASTER_KEY` environment variable.

1. **Generate a key:**
   ```bash
   openssl rand -base64 32
   ```

2. **Configure the key:**
   - **Environment variable:** set it in your environment or process manager:
     ```bash
     export YILY_MASTER_KEY="<your-generated-base64-key>"
     ```
   **Local development with Air:** from the `server` directory, copy `.env.example` to `.env` and set `YILY_MASTER_KEY`. Air loads `.env` into the process environment when running `air`.
   > **Note:** The Go binary reads `YILY_MASTER_KEY` directly from process environment variables via `os.LookupEnv`. If running without Air (e.g. `go run ./cmd/server` or `./tmp/server`), export the variable in your shell.

3. **Important precautions:**
   - **Keep it secret:** Never commit the master key or a populated `.env` file to version control.
   - **Backup securely:** Store `YILY_MASTER_KEY` in a secure location separate from the SQLite database.
   - **Unrecoverable on loss:** If the master key is lost, all encrypted secrets become permanently unrecoverable.
   - **Persistence:** Do not generate a new key on every restart; reuse the same key to retain access to existing data.

## License

Yily is licensed under the MIT License. See the [LICENSE](LICENSE) file for more information.
