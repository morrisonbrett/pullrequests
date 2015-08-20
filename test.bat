set GOPATH=%HOMEDRIVE%%HOMEPATH%
go vet
go fmt
..\..\bin\golint.exe
del coverage.out
go test -v -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out