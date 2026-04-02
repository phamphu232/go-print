# GO-PRINT

## Setup

```
# Install packages.
go mod download
go mod tidy

# Install live reload
go install github.com/air-verse/air@latest

# Live reload
air

# Install rsrc
go install github.com/akavel/rsrc@latest

# Build with icon
rsrc -manifest main.manifest -o main.syso

# Build
./build.sh
```