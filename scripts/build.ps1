param(
    [ValidateSet('build', 'test','test-race','lint','run','docker-build','docker-run','clean','help')]
    [string]$Command = 'help',
    [string]$Args
)

$BINARY_NAME = 'gocrawl'
$IMAGE_NAME = 'gocrawl'

Write-Host "Running: $Command" -ForegroundColor Cyan

switch ($Command) {
    'help' {
    Write-Host "Available commands:" -ForegroundColor Cyan
    Write-Host "  build        - Build binary"
    Write-Host "  test         - Run tests"
    Write-Host "  test-race    - Run tests with race detector"
    Write-Host "  lint         - Run linter"
    Write-Host "  run          - Run application"
    Write-Host "  docker-build - Build Docker image"
    Write-Host "  docker-run   - Run Docker container"
    Write-Host "  clean        - Cleanup"
    Write-Host "  help         - Show help"
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    Write-Host "  .\scripts\build.ps1 -Command <command> [-Args <arguments>]"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Cyan
    Write-Host "  .\scripts\build.ps1 -Command build"
    Write-Host "  .\scripts\build.ps1 -Command run -Args `"-file test.txt`""
    Write-Host "  .\scripts\build.ps1 -Command docker-run -Args `"-file /data/test.txt`""
             }
    'build'{
        Write-Host "Building $BINARY_NAME..." -ForegroundColor Yellow
        go build -o $BINARY_NAME ./cmd/gocrawl
        if ($LASTEXITCODE -eq 0){
            Write-Host "Build completed successfully" -ForegroundColor Green
        }
    }
    'test'{
        Write-Host "Running tests..." -ForegroundColor Yellow
        go test ./...
    }
    'test-race'{
        Write-Host "Running tests with race detector..." -ForegroundColor Yellow
        go test -race ./...
    }
    'lint'{
        Write-Host "Running linter..." -ForegroundColor Yellow
        golangci-lint run ./...
    }
    'run'{
        Write-Host "Running application..." -ForegroundColor Yellow
        go run ./cmd/gocrawl $Args
    }
    'docker-build'{
        Write-Host "Building Docker image..." -ForegroundColor Yellow
        docker build -t $IMAGE_NAME
    }
    'docker-run'{
        Write-Host "Running docker container..." -ForegroundColor Yellow
        docker run --rm -it -v ${PWD}:/data $IMAGE_NAME $Args
    }
    'clean'{
        Write-Host "Cleaning..." -ForegroundColor Yellow
        Remove-Item -Force $BINARY_NAME -ErrorAction SilentlyContinue
        docker rmi $IMAGE_NAME -Force 2>$null
        Write-Host "Clean completed" -ForegroundColor Green
    }
}