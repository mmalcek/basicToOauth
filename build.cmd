set appVersion="1-0-3"

@ECHO OFF
IF [%1]==[windows] CALL :WINDOWS
IF [%1]==[linux] CALL :LINUX
IF [%1]==[clean] CALL :CLEAN
IF [%1]==[all] CALL :ALL
EXIT /B 1

:ALL
 CALL :CLEAN
 CALL :WINDOWS
 CALL :LINUX 
EXIT /B 0

:WINDOWS
	mkdir build\basicToOauth_Windows_amd64_%appVersion%
	go build -trimpath -ldflags "-s -w" -a -o build\basicToOauth_Windows_amd64_%appVersion%\basicToOauth.exe
    CALL COPY LICENSE build\basicToOauth_Windows_amd64_%appVersion%\LICENSE
    CALL COPY config.yaml build\basicToOauth_Windows_amd64_%appVersion%\config.yaml
    CALL COPY README.md build\basicToOauth_Windows_amd64_%appVersion%\README.md
	PUSHD build
	tar.exe -a -c -f basicToOauth_Windows_amd64_%appVersion%.zip basicToOauth_Windows_amd64_%appVersion%
	POPD
EXIT /B 0

:LINUX
    mkdir build\basicToOauth_Linux_amd64_%appVersion%
	set GOOS=linux&& set GOARCH=amd64&& go build -trimpath -ldflags "-s -w" -a -o build\basicToOauth_Linux_amd64_%appVersion%\basicToOauth
    CALL COPY LICENSE build\basicToOauth_Linux_amd64_%appVersion%\LICENSE
    CALL COPY config.yaml build\basicToOauth_Linux_amd64_%appVersion%\config.yaml
    CALL COPY README.md build\basicToOauth_Linux_amd64_%appVersion%\README.md
	PUSHD build
	tar.exe -a -c -f basicToOauth_Linux_amd64_%appVersion%.zip basicToOauth_Linux_amd64_%appVersion%
	POPD
EXIT /B 0

:CLEAN
    rmdir /s /q build
EXIT /B 0