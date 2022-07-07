# Stats Client Example

This example demonstrates how to retrieve statistics from VPP using [the new Stats API](https://github.com/FDio/vpp/blob/master/src/vpp/stats/stats.md).

## Requirements

The following requirements are required to run this example:

- install VPP **18.10+** (VPP 19.04+ for statsclient)
- enable stats in VPP

To enable stats add following section to you VPP config:

  ```sh
  statseg {
  	default
  	per-node-counters on
  }
  ```
  > The [default socket](https://wiki.fd.io/view/VPP/Command-line_Arguments#.22statseg.22_parameters) is located at `/run/vpp/stats.sock`.

## Running example

First build the example: `go build go.fd.io/govpp/examples/stats-api`.

### Higher-level access to stats

Use commands following commands to retrieve stats that are aggregated and
processed into logical structures from [api package](../../api).

- `system` to retrieve system statistics
- `nodes` to retrieve per node statistics
- `interfaces` to retrieve per interface statistics
- `errors` to retrieve error statistics (you can use patterns to filter the errors)

#### System stats

Following command will retrieve system stats.
```
$ ./stats-api system
System stats: &{VectorRate:0 InputRate:0 LastUpdate:32560 LastStatsClear:0 Heartbeat:3255}
```

#### Node stats

Following command will retrieve per node stats.
```
$ ./stats-api nodes
Listing node stats..
...
 - {NodeIndex:554 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:555 Clocks:189609 Vectors:15 Calls:15 Suspends:0}
 - {NodeIndex:556 Clocks:2281847 Vectors:0 Calls:0 Suspends:21}
 - {NodeIndex:557 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:558 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:559 Clocks:7094 Vectors:0 Calls:1 Suspends:1}
 - {NodeIndex:560 Clocks:88159323916601 Vectors:0 Calls:14066116 Suspends:0}
 - {NodeIndex:561 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:562 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:563 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:564 Clocks:447894125 Vectors:0 Calls:0 Suspends:32395}
 - {NodeIndex:565 Clocks:1099655497824612 Vectors:0 Calls:40 Suspends:117}
 - {NodeIndex:566 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:567 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:568 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:569 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:570 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:571 Clocks:0 Vectors:0 Calls:0 Suspends:0}
 - {NodeIndex:572 Clocks:0 Vectors:0 Calls:0 Suspends:0}
Listed 573 node counters
```

#### Interface stats

Following command will retrieve per interface stats.
```
$ ./stats-api interfaces
Listing interface stats..
 - {InterfaceIndex:0 RxPackets:0 RxBytes:0 RxErrors:0 TxPackets:0 TxBytes:0 TxErrors:0 RxUnicast:[0 0] RxMulticast:[0 0] RxBroadcast:[0 0] TxUnicastMiss:[0 0] TxMulticast:[0 0] TxBroadcast:[0 0] Drops:0 Punts:0 IP4:0 IP6:0 RxNoBuf:0 RxMiss:0}
 - {InterfaceIndex:1 RxPackets:0 RxBytes:0 RxErrors:0 TxPackets:0 TxBytes:0 TxErrors:0 RxUnicast:[0 0] RxMulticast:[0 0] RxBroadcast:[0 0] TxUnicastMiss:[0 0] TxMulticast:[0 0] TxBroadcast:[0 0] Drops:5 Punts:0 IP4:0 IP6:0 RxNoBuf:0 RxMiss:0}
 - {InterfaceIndex:2 RxPackets:0 RxBytes:0 RxErrors:0 TxPackets:0 TxBytes:0 TxErrors:0 RxUnicast:[0 0] RxMulticast:[0 0] RxBroadcast:[0 0] TxUnicastMiss:[0 0] TxMulticast:[0 0] TxBroadcast:[0 0] Drops:0 Punts:0 IP4:0 IP6:0 RxNoBuf:0 RxMiss:0}
 - {InterfaceIndex:3 RxPackets:0 RxBytes:0 RxErrors:0 TxPackets:0 TxBytes:0 TxErrors:0 RxUnicast:[0 0] RxMulticast:[0 0] RxBroadcast:[0 0] TxUnicastMiss:[0 0] TxMulticast:[0 0] TxBroadcast:[0 0] Drops:0 Punts:0 IP4:0 IP6:0 RxNoBuf:0 RxMiss:0}
Listed 4 interface counters
```

#### Error stats

Following command will retrieve error stats.
Use flag `-all` to include stats with zero values.
```
$ ./stats-api errors ip
Listing error stats.. ip
 - {ip4-input/ip4 spoofed local-address packet drops 15}
Listed 1 (825) error counters
```

### Low-level access to stats

Use commands `ls` and `dump` to list and dump statistics in their raw format
from [adapter package](../../adapter).
Optionally, patterns can be used to filter the results.

#### List stats

Following command will list stats matching patterns `/sys/` and `/if/`.
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

#### Dump stats

Following command will dump stats with their types and actual values.
Use flag `-all` to include stats with zero values.
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
