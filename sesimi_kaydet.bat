@echo off
chcp 65001 >nul
cd /d "%~dp0"
set "PYTHONPATH=%~dp0src;%PYTHONPATH%"
if exist "%~dp0.venv\Scripts\python.exe" (set "TUBIE_PYTHON=%~dp0.venv\Scripts\python.exe") else (set "TUBIE_PYTHON=python")
"%TUBIE_PYTHON%" -m tubie_stickman record %*
exit /b %ERRORLEVEL%
