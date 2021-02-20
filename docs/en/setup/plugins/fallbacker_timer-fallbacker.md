# Fallbacker/timer-fallbacker
## Description
This is a timer fallback trigger to process the forward failure data.
## DefaultConfig
```yaml
# The forwarder max retry times.
max_times: 3
# The latency_factor is the standard retry duration, and the time for each retry is expanded by 2 times until the number 
# of retries reaches the maximum.(Time unit is millisecond.)
latency_factor: 2000
# The max retry latency time.(Time unit is millisecond.)
max_latency_time: 5000
```
