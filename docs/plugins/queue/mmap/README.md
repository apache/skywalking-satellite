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

## Configuration
[Configuration Params](../../../configuration/queue.md)

## Segment
Segments are a series of files of the same size. Each input data would cost `8Bit+Len(data)Bit` to store the raw bytes. The first 8Bit is equal to `Len(data)` for finding the ending position. 
### Swapper
For the performance and resources thinking, we define a page replacement policy.

- Keep the reading and writing segments on the memory.
- When the mmapcount is greater or equals to the max_in_mem_segments, we first scan the read scope and then scan the written scope to swap the segments to promise the reading or writing segments are always in memory.
    - Read scope: [reading_offset - max_in_mem_segments,reading_offset - 1]
    - Written scope: [writing_offset - max_in_mem_segments,writing_offset - 1]
    - Each displacement operation guarantees at least `max_in_mem_segments/2-1` capacity available. Subtract operation to subtract the amount of memory that must always exist.

## BenchmarkTest
Test machine: macbook pro 2018

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
BenchmarkPush
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:10000
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:10000-12         	   45946	     28185 ns/op	    9884 B/op	       9 allocs/op
BenchmarkPush/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000
BenchmarkPush/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000-12         	   68137	     19142 ns/op	    9838 B/op	       9 allocs/op
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:20_message:8KB_queueCapacity:10000
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:20_message:8KB_queueCapacity:10000-12         	   47361	     22318 ns/op	    9884 B/op	       9 allocs/op
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000-12        	   24478	     45501 ns/op	   18934 B/op	      10 allocs/op
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000-12        	   45691	     29413 ns/op	    9884 B/op	       9 allocs/op
PASS
```
### push and pop operation
```
goos: darwin
goarch: amd64
pkg: github.com/apache/skywalking-satellite/plugins/queue/mmap
BenchmarkPushAndPop
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:10000
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:10000-12         	   30657	     34182 ns/op	   28725 B/op	      41 allocs/op
BenchmarkPushAndPop/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000
BenchmarkPushAndPop/segmentSize:_256KB_maxInMemSegments:10_message:8KB_queueCapacity:10000-12         	   34617	     31619 ns/op	   28677 B/op	      41 allocs/op
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:20_message:8KB_queueCapacity:10000
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:20_message:8KB_queueCapacity:10000-12         	   32440	     38439 ns/op	   28726 B/op	      41 allocs/op
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:16KB_queueCapacity:10000-12        	   18554	     56840 ns/op	   54931 B/op	      42 allocs/op
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB_queueCapacity:100000-12        	   27303	     40016 ns/op	   28725 B/op	      41 allocs/op
PASS
```
