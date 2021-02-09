# Common configuration
The common configuration has 2 parts, which are logger
configuration and the telemetry configuration.

## Logger
|  Config   |Default| Description  |
|  ----  | ----  | ----  |
| log_pattern  | %time [%level][%field] - %msg | The log format pattern configuration.|
| time_pattern  | 2006-01-02 15:04:05.000 |The time format pattern configuration.|
| level  | info |The lowest level of printing allowed.|

## Self Telemetry
|  Config   |Default| Description  |
|  ----  | ----  | ----  |
| cluster  | default-cluster | The space concept for the deployment, such as the namespace concept in the kubernetes.|
| service  | default-service | The group concept for the deployment, such as the service resource concept in the kubernetes.|
| instance  | default-instance |The minimum running unit, such as the pod concept in the kubernetes.|

