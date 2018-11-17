# Stats API Example

This example demonstrates how to retrieve statistics from VPP using [the new Stats API](https://github.com/FDio/vpp/blob/master/src/vpp/stats/stats.md).

## Requirements

The following requirements are required to run this example:

- install **VPP 18.10+**
- enable stats in VPP:

  ```sh
  statseg {
  	default
  }
  ``` 
  > The [default socket](https://wiki.fd.io/view/VPP/Command-line_Arguments#.22statseg.22_parameters) is located at `/run/vpp/stats.sock`.
- run the VPP, ideally with some traffic

## Running example

First build the example: `go build git.fd.io/govpp.git/examples/cmd/stats-api`. 

Use commands `ls` and `dump` to list and dump statistics. Optionally, patterns can be used to filter the results.

### List stats matching patterns `/sys/` and `/if/`
```
$ ./stats-api ls /sys/ /if/
Listing stats.. /sys/ /if/
 - /sys/vector_rate
 - /sys/input_rate
 - /sys/last_update
 - /sys/last_stats_clear
 - /sys/heartbeat
 - /sys/node/clocks
 - /sys/node/vectors
 - /sys/node/calls
 - /sys/node/suspends
 - /if/drops
 - /if/punt
 - /if/ip4
 - /if/ip6
 - /if/rx-no-buf
 - /if/rx-miss
 - /if/rx-error
 - /if/tx-error
 - /if/rx
 - /if/rx-unicast
 - /if/rx-multicast
 - /if/rx-broadcast
 - /if/tx
 - /if/tx-unicast-miss
 - /if/tx-multicast
 - /if/tx-broadcast
Listed 25 stats
```

### Dump all stats with their types and values
```
$ ./stats-api dump
Dumping stats..
 - /sys/last_update                       ScalarIndex 10408
 - /sys/heartbeat                         ScalarIndex 1041
 - /err/ip4-icmp-error/unknown type        ErrorIndex 5
 - /net/route/to                CombinedCounterVector [[{Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:0 Bytes:0} {Packets:5 Bytes:420}]]
 - /if/drops                      SimpleCounterVector [[0 5 5]]
Dumped 5 (2798) stats
```
