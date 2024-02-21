@echo off

if "%1" == "build" goto :build
if "%1" == "build-release" goto :build-release
if "%1" == "clean" goto :clean

REM Default target
if "%1" == "" goto :build-release

echo Invalid target: %1
echo Usage: .\make.bat [build^|build-release^|clean]
goto :eof

:build
	if not exist "build" mkdir build
	go build -o .\build\secure-login.exe
	goto :eof

:build-release
	if not exist "build" mkdir build
	go build --ldflags "-s -w" -o .\build\secure-login.exe
	goto :eof

:clean
	rd /s /q build
	goto :eof
