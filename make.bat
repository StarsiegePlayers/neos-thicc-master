@echo off
rem windows makefile????

set version=1.0.0 Test
set outputdir=build
set output=mstrsvr
set packagename=neos-dummy-thicc-master-%version%
set ldflags=-X 'main.VERSION=%version%' -X 'main.TIME=%TIME%' -X 'main.DATE=%DATE%'

if "%1"=="" call :release
if "%1"=="release" call :release
if "%1"=="debug" call :debug
if "%1"=="generate" call :generate
if "%1"=="clean" call :clean
goto end

:generate
echo ---generate---
echo go generate
echo.
go generate
goto :eof

:debug
echo =====Debug Build started %DATE% %TIME%=====
echo.
call :clean
call :generate
if "%GOOS%"=="windows" set output=%output%.exe
if "%GOOS%"=="" set output=%output%.exe
echo ---debug---
echo go build -o %outputdir%\%output% -x -ldflags="%ldflags%"
echo.
go build -o %outputdir%\%output% -x -ldflags="%ldflags%"
goto :eof

:release
echo =====Release Build started %DATE% %TIME%=====
echo.
call :clean
call :generate
set ldflags=%ldflags% -s -w

:: *Nix/BSD/Windows
for %%G in (windows linux freebsd openbsd netbsd) do (
    for %%U in (386 amd64 arm arm64) do (
        call :buildRelease %%G %%U
    )
)

:: MacOS
for %%U in (amd64 arm64) do (
   call :buildRelease darwin %%U
)

:: Android
call :buildRelease android arm64

call :package
goto :eof


:buildRelease
echo --- release %1 %2 ---
echo go build -o %outputdir%\%2\%output%_%1 -ldflags="%ldflags%"
echo.
set GOOS=%1
set GOARCH=%2
go build -o %outputdir%\%2\%output%_%1 -ldflags="%ldflags%"
if "%1"=="windows" (
    echo move %outputdir%\%2\%output%_%1 %outputdir%\%2\%output%.exe
    move %outputdir%\%2\%output%_%1 %outputdir%\%2\%output%.exe
)
goto :eof

:clean
echo ---clean---
echo rmdir /S /Q build\
echo.
rmdir /S /Q build\ >nul 2>nul
echo del /F /Q *.syso
echo.
del /F /Q *.syso >nul 2>nul
goto :eof

:package
copy %output%.yaml build\
cd %outputdir%
move 386 x86
move amd64 x86_64
7z a "%packagename%-release.7z" *
goto :eof

:end
