# Setting Override
SkyWalking Satellite supports setting overrides by system environment variables. 
You could override the settings in `satellite_config.yaml`


## System environment variables
- Example

  Override `log_pattern` in this setting segment through environment variables
  
```yaml
logger:
  log_pattern: ${SATELLITE_LOGGER_LOG_PATTERN:%time [%level][%field] - %msg}
  time_pattern: ${SATELLITE_LOGGER_TIME_PATTERN:2006-01-02 15:04:05.000}
  level: ${SATELLITE_LOGGER_LEVEL:info}
```

If the `SATELLITE_LOGGER_LOG_PATTERN ` environment variable exists in your operating system and its value is `%msg`, 
then the value of `log_pattern` here will be overwritten to `%msg`, otherwise, it will be set to `%time [%level][%field] - %msg`.





