# What is the performance of the Satellite?
## Performance
The performance reduction of the mmap-queue  is mainly due to the file persistent operation to ensure data stability. However, the queue is used to collect some core telemetry data. We will continue to optimize the performance of this queue.

- 0.5 core supported 3000 ops throughput with memory queue.
- 0.5 core supported 1500 ops throughput with the memory mapped queue(Ensure data stability).


## Details
### Testing environment
1. machine: 
    -  cpu: INTEL Xeon E5-2650 V4 12C 2.2GHZ * 2
    - memory: INVENTEC PC4-19200 * 8
    - harddisk: INVENTEC SATA 4T 7.2K * 8
2. Kafka: 
    - region: the same region with the test machine in Baidu Cloud.
    - version.: 0.1.1.0
3. The input plugin: grpc-nativelog-receiver
4. resource limit:
    - cpu: 500m(0.5 core)
    - memory: 100M

### Performance Test With Memory Queue
|  Qps   |stack memory in use| heap memory in use  |no-heap memory in use | 
|  ----  | ----  | ----  | ----  |
| 400  | 2.13M | 11M |83K|
| 800  | 2.49M | 13.4M |83K|
| 1200  | 2.72M | 13.4M |83K|
| 1600  | 2.85M | 16.2M |83K|
| 2000  | 2.92M | 17.6M |83K|
| 2400  | 2.98M | 18.3M |83K|
| 2800  | 3.54M | 26.8M |83K|
| 3000  | 3.34M | 28M |83K|

### Performance Test With Mmap Queue
|  Qps   |stack memory in use| heap memory in use  |no-heap memory in use | 
|  ----  | ----  | ----  | ----  |
| 400  | 2.39M | 9.5M |83K|
| 800  | 2.43M | 12.1M |83K|
| 1200  | 2.49M | 12M |83K|
| 1600  | 2.62M | 13.3M |83K|
