# ALS Load Balance

Using satellite as a load balancer in envoy and OAP can effectively prevent the problem of unbalanced messages received by OAP.

In this case, we mainly use memory queues for intermediate data storage. 

Deference Envoy Count, OAP performance could impact the Satellite transmit performance.

|Envoy Instance|Concurrent User|ALS OPS|Satellite CPU|Satellite Memory|
|--------------|---------------|-------|-------------|----------------|
|150|100|~50K|1.2C|0.5-1.0G|
|150|300|~80K|1.8C|1.0-1.5G|
|300|100|~50K|1.4C|0.8-1.2G|
|300|300|~100K|2.2C|1.3-2.0G|
|800|100|~50K|1.5C|0.9-1.5G|
|800|300|~100K|2.6C|1.7-2.7G|
|1500|100|~50K|1.7C|1.4-2.4G|
|1500|300|~100K|2.7C|2.3-3.0G|
|2300|150|~50K|1.8C|1.9-3.1G|
|2300|300|~90K|2.5C|2.3-4.0G|
|2300|500|~110K|3.2C|2.8-4.7G|

## Detail

Using GKE Environment, helm to build cluster.

|Module|Version|Replicate Count|CPU Limit|Memory Limit|Description|
|------|-------|---------------|---------|------------|-----------|
|OAP|8.9.0|6|12C|32Gi|Using ElasticSearch as Storage|
|Satellite|0.4.0|1|8C|16Gi||
|ElasticSearch|7.5.1|3|8|16Gi||

