# Design
The mmap-queue is a big, fast and persistent queue based on the memory mapped files. One mmap-queue has a directory to store the whole data. The Queue directory is made up with many segments and 1 meta file. 

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
Base machine: macbook pro 2018 Intel Core i7 16 GB 2400 MHz DDR4 SSD

### push operation

```
BenchmarkPush
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:8KB-12         	   42470	     25098 ns/op	     466 B/op	      10 allocs/op

BenchmarkPush/segmentSize:_256KB_maxInMemSegments:10_message:8KB
BenchmarkPush/segmentSize:_256KB_maxInMemSegments:10_message:8KB-12         	   76905	     18910 ns/op	     418 B/op	      10 allocs/op

BenchmarkPush/segmentSize:_128KB_maxInMemSegments:20_message:8KB
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:20_message:8KB-12         	   58221	     22258 ns/op	     465 B/op	      10 allocs/op

BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:16KB
BenchmarkPush/segmentSize:_128KB_maxInMemSegments:10_message:16KB-12        	   34053	     48635 ns/op	     566 B/op	      11 allocs/op
```
### push and pop operation
```
BenchmarkPushAndPop
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:8KB-12         	   22273	     45872 ns/op	   19512 B/op	      40 allocs/op

BenchmarkPushAndPop/segmentSize:_256KB_maxInMemSegments:10_message:8KB
BenchmarkPushAndPop/segmentSize:_256KB_maxInMemSegments:10_message:8KB-12         	   38874	     37169 ns/op	   19456 B/op	      40 allocs/op

BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:20_message:8KB
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:20_message:8KB-12         	   38048	     36274 ns/op	   19514 B/op	      40 allocs/op

BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:16KB
BenchmarkPushAndPop/segmentSize:_128KB_maxInMemSegments:10_message:16KB-12        	   19768	     63399 ns/op	   36893 B/op	      41 allocs/op
```
