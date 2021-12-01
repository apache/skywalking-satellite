@REM
@REM  Licensed to the Apache Software Foundation (ASF) under one or more
@REM  contributor license agreements.  See the NOTICE file distributed with
@REM  this work for additional information regarding copyright ownership.
@REM  The ASF licenses this file to You under the Apache License, Version 2.0
@REM  (the "License"); you may not use this file except in compliance with
@REM  the License.  You may obtain a copy of the License at
@REM
@REM      http://www.apache.org/licenses/LICENSE-2.0
@REM
@REM  Unless required by applicable law or agreed to in writing, software
@REM  distributed under the License is distributed on an "AS IS" BASIS,
@REM  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
@REM  See the License for the specific language governing permissions and
@REM  limitations under the License.

@echo off

setlocal
set SATELLITE_PROCESS_TITLE=Skywalking-Satellite
set SATELLITE_HOME=%~dp0%..

for /F %%i in ('where /r "%SATELLITE_HOME%\bin" skywalking-satellite*windows*') do ( set SATELLITE_EXEC=%%i)

start "%SATELLITE_PROCESS_TITLE%" %SATELLITE_EXEC% start --config="%SATELLITE_HOME%\configs\satellite_config.yaml"

endlocal
