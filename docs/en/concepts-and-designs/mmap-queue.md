# Design
The mmap-queue is a big, fast, and persistent queue based on the memory-mapped files. One mmap-queue has a directory to store the whole data. The queue directory is made up of many segments and 1 metafile. This is originally implemented by [bigqueue](https://github.com/grandecola/bigqueue) project, we changed it a little for fitting the Satellite project requirements.

- Segment: Segment is the real data store center, that provides large-space storage and does not reduce read and write performance as much as possible by using mmap. And we will avoid deleting files by reusing them.
- Meta: The purpose of meta is to find the data that the consumer needs.

## Meta
Metadata only needs 80B to store the Metadata for the pipe. But for memory alignment, it takes at least one memory page size, which is generally 4K.
```
[    8Bit   ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit  ]
[metaVersion][  ID   ][ offset][  ID   ][ offset][  ID   ][ offset][  ID   ][ offset][capacity]
[metaVersion][writing   offset][watermark offset][committed offset][reading   offset][capacity]

```
### Transforming

![](https://skywalking.apache.org/blog/2020-11-25-skywalking-satellite-0.1.0-design/offset-convert.jpg)


## BenchmarkTest
Test machine: **macbook pro 2018**

```
Model Name:	MacBook Pro
Model Identifier:	MacBookPro15,1
Processor Name:	6-Core Intel Core i7
Processor Speed:	2.2 GHz
Number of Processors:	1
Total Number of Cores:	6
L2 Cache (per Core):	256 KB
L3 Cache:	9 MB
Hyper-Threading Technology:	Enabled
Memory:	16 GB
System Firmware Version:	1554.60.15.0.0 (iBridge: 18.16.13030.0.0,0
```

### push operation

```
goos: darwin
goarch: amd64
pkg: github.com/apache/skywalking-satellite/plugins/queue/mmap
BenchmarkEnqueue
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:18_message:8KB_queueCapacity:10000         	   10000	    106520 ns/op	    9888 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000         	   18536	     54331 ns/op	    9839 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_512KB_maxInMemSegments:6_message:8KB_queueCapacity:10000          	   27859	     43251 ns/op	    9815 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_256KB_maxInMemSegments:20_message:8KB_queueCapacity:10000         	   23673	     45910 ns/op	    9839 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000        	   10000	    131686 ns/op	   18941 B/op	      10 allocs/op
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000        	   23011	     47101 ns/op	    9887 B/op	       9 allocs/op
PASS
```
### push and pop operation
```
goos: darwin
goarch: amd64
pkg: github.com/apache/skywalking-satellite/plugins/queue/mmap
BenchmarkEnqueueAndDequeue
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:18_message:8KB_queueCapacity:10000         	   18895	     53056 ns/op	   28773 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000         	   24104	    117128 ns/op	   28725 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_512KB_maxInMemSegments:6_message:8KB_queueCapacity:10000          	   23733	     71632 ns/op	   28699 B/op	      41 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_256KB_maxInMemSegments:20_message:8KB_queueCapacity:10000         	   26286	     64377 ns/op	   28725 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000        	   10000	    118004 ns/op	   54978 B/op	      43 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000        	   16489	     64400 ns/op	   28772 B/op	      42 allocs/op
PASS
```
