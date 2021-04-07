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
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:18_message:8KB_queueCapacity:10000         	   27585	     43559 ns/op	    9889 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000         	   39326	     31773 ns/op	    9840 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_512KB_maxInMemSegments:6_message:8KB_queueCapacity:10000          	   56770	     22990 ns/op	    9816 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_256KB_maxInMemSegments:20_message:8KB_queueCapacity:10000         	   43803	     29778 ns/op	    9840 B/op	       9 allocs/op
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000        	   16870	     80576 ns/op	   18944 B/op	      10 allocs/op
BenchmarkEnqueue/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000        	   36922	     39085 ns/op	    9889 B/op	       9 allocs/op
PASS
```
### push and pop operation
```
goos: darwin
goarch: amd64
pkg: github.com/apache/skywalking-satellite/plugins/queue/mmap
BenchmarkEnqueueAndDequeue
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:18_message:8KB_queueCapacity:10000         	   21030	     60728 ns/op	   28774 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000         	   30327	     41274 ns/op	   28726 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_512KB_maxInMemSegments:6_message:8KB_queueCapacity:10000          	   32738	     37923 ns/op	   28700 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_256KB_maxInMemSegments:20_message:8KB_queueCapacity:10000         	   28209	     41169 ns/op	   28726 B/op	      42 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000        	   14677	     89637 ns/op	   54981 B/op	      43 allocs/op
BenchmarkEnqueueAndDequeue/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000        	   22228	     54963 ns/op	   28774 B/op	      42 allocs/op
PASS
```
