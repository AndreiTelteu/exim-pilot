@echo off
REM Mock Exim binary for testing on Windows

if "%1"=="-bp" (
    echo    2h  1024 1b2c3d-000001-AB ^<sender@example.com^>
    echo          recipient@example.com
    echo    5m  2048 4e5f6g-000002-CD ^<user@test.com^>
    echo          admin@example.org
    echo          user2@test.com
    exit /b 0
)

if "%1"=="-Mvh" (
    echo From: sender@example.com
    echo To: recipient@example.com
    echo Subject: Test message
    echo Date: %date% %time%
    exit /b 0
)

if "%1"=="-Mvb" (
    echo This is a test message body.
    exit /b 0
)

if "%1"=="-Mvl" (
    exit /b 1
)

if "%1"=="-M" (
    echo Message %2 delivery initiated
    exit /b 0
)

if "%1"=="-Mf" (
    echo Message %2 frozen
    exit /b 0
)

if "%1"=="-Mt" (
    echo Message %2 thawed
    exit /b 0
)

if "%1"=="-Mrm" (
    echo Message %2 removed
    exit /b 0
)

REM Default version output
echo Mock Exim 4.96
exit /b 0