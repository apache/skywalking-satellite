# Fallbacker/timer-fallbacker
## Description
This is a timer fallback trigger to process the forward failure data.
## DefaultConfig
```yaml
# The forwarder max attempt times.
max_attempts: 3
# The exponential_backoff is the standard retry duration, and the time for each retry is expanded
# by 2 times until the number of retries reaches the maximum.(Time unit is millisecond.)
exponential_backoff: 2000
# The max latency time used in retrying, which would override the latency time when the latency time
# with exponential increasing larger than it.(Time unit is millisecond.)
max_latency_time: 5000
```
