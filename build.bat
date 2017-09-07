go test ./...

call :build linux arm  
call :build linux amd64  
call :build linux 386  
call :build darwin amd64  
call :build darwin 386
goto :end

:build
	set GOOS=%1
	set GOARCH=%2
	go tool dist install pkg/runtime
	go install -a std
	go build -buildmode=plugin -o build/metering-%1-%2%3 -i .
:end