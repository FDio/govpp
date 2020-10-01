# Multi-VPP example

This example shows how to use GoVPP client to connect, configure and read stats from multiple VPPs simultaneously.

# Requirements

* VPP 19.08 or newer (required for stats client)
* VPP stats enabled

The example requires two running VPP instances. VPPs can be simply started in the same machine with different startup configs. Note that the default path to binary API and stats sockets are `/run/vpp/api.sock` or `/run/vpp/stats.sock` respectively. The example always uses the default path if none is set. It means that at least one VPP must have the following fields redefined:

```
socksvr {
  socket-name /run/custom-vpp-path/api.sock
}

statseg {
  socket-name /run/custom-vpp-path/stats.sock
}
```

And the custom path must be provided to the example. Four flags are available:
```
-api-sock-1 string - Path to binary API socket file of the VPP1 (default "/run/vpp/api.sock")
-api-sock-2 string - Path to binary API socket file of the VPP2 (default "/run/vpp/api.sock")
-stats-sock-1 string - Path to stats socket file of the VPP1 (default "/run/vpp/stats.sock")
-stats-sock-2 string - Path to stats socket file of the VPP2 (default "/run/vpp/stats.sock")
```
Let's say the VPP1 uses the default config, and the config above belongs to the VPP2. In that case, use the following command:
```
sudo ./multi-vpp -api-sock-2=/run/custom-vpp-path/api.sock -stats-sock-2=/run/custom-vpp-path/stats.sock
```

# Running the example

The example consists of the following steps:
* connects to both VPPs binary API socket and stats socket
* configures example interfaces with IP addresses
* dumps interface data via the binary API
* dumps interface data via socket client
* in case there are no errors, cleans up VPPs in order to be able running the example in a loop











