# ShortURL

ShortURL is a URL shortening service written in Go. It provides a simple and efficient way to generate short, shareable links for long URLs.

## Features

- Generate short URLs for long links.
- Redirect users from short URLs to their original destinations.
- Simple and lightweight implementation.
- Easy to deploy and extend.

## Requirements

- Go 1.24 or later
- A database (e.g., PostgreSQL, MySQL, or SQLite)

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/yourusername/shorturl.git
    cd shorturl
    ```

2. Build the project:
    ```bash
    go build
    ```

3. Run the application:
    ```bash
    ./shorturl
    ```

## Configuration

Update the `config.yaml` file to set up database credentials and other application settings.

## Usage

- Access the web interface at `http://localhost:8080`.
- Use the API to programmatically shorten URLs.

## License

This project is licensed under the BSD License.
