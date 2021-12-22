@echo off
rem windows makefile????

set version=0.9.0 Test
set output=mstrsvr.exe
set ldflags=-X 'main.VERSION=%version%' -X 'main.TIME=%TIME%' -X 'main.DATE=%DATE%'
rem set GOOS=linux

if "%1"=="" call :release
if "%1"=="release" call :release
if "%1"=="debug" call :debug
if "%1"=="clean" call :clean
goto end

:debug
echo =====Debug Build started %DATE% %TIME%=====
echo.
call :clean
echo ---debug---
echo go build -o %output% -x -ldflags="%ldflags%"
echo.
go build -o %output% -x -ldflags="%ldflags%"
goto :eof

:release
echo =====Release Build started %DATE% %TIME%=====
echo.
call :clean
set ldflags=%ldflags% -s -w
echo ---release---
echo go build -o %output% -x -ldflags="%ldflags%"
echo.
go build -o %output% -x -ldflags="%ldflags%"
goto :eof

:clean
echo ---clean---
echo del *.exe *.syso
echo.
del *.exe *.syso >nul 2>nul
goto :eof

:end
