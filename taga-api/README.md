# taga-api

A Go web API built with Gin and Zap logging.

## Features

- **Gin**: Fast HTTP web framework
- **Zap**: Structured, leveled logging

## Getting Started

### Prerequisites

- Go 1.19 or higher

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Available Endpoints

- `GET /` - Welcome message
- `GET /health` - Health check

### Building

```bash
go build -o taga-api
./taga-api
```

## Development

### Project Structure

```
taga-api/
├── main.go          # Main application file
├── go.mod           # Go module file
├── go.sum           # Go dependencies
├── .gitignore       # Git ignore rules
└── README.md        # This file
```
<!-- SAMPLE TEXT -->