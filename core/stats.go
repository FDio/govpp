package core

import (
	"path"
	"strings"
	"time"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/api"
)

var (
	RetryUpdateCount    = 10
	RetryUpdateDelay    = time.Millisecond * 10
	HealthCheckInterval = time.Second // default health check probe interval
)

const (
	SystemStatsPrefix               = "/sys/"
	SystemStats_VectorRate          = SystemStatsPrefix + "vector_rate"
	SystemStats_NumWorkerThreads    = SystemStatsPrefix + "num_worker_threads"
	SystemStats_VectorRatePerWorker = SystemStatsPrefix + "vector_rate_per_worker"
	SystemStats_InputRate           = SystemStatsPrefix + "input_rate"
	SystemStats_LastUpdate          = SystemStatsPrefix + "last_update"
	SystemStats_LastStatsClear      = SystemStatsPrefix + "last_stats_clear"
	SystemStats_Heartbeat           = SystemStatsPrefix + "heartbeat"

	NodeStatsPrefix    = "/sys/node/"
	NodeStats_Names    = NodeStatsPrefix + "names"
	NodeStats_Clocks   = NodeStatsPrefix + "clocks"
	NodeStats_Vectors  = NodeStatsPrefix + "vectors"
	NodeStats_Calls    = NodeStatsPrefix + "calls"
	NodeStats_Suspends = NodeStatsPrefix + "suspends"

	BufferStatsPrefix     = "/buffer-pools/"
	BufferStats_Cached    = "cached"
	BufferStats_Used      = "used"
	BufferStats_Available = "available"

	CounterStatsPrefix = "/err/"

	MemoryStatSegPrefix = "/mem/statseg"
	MemoryStatSegment   = "/mem/stat segment"
	MemoryMainHeap      = "/mem/main heap"
	MemoryStats_Total   = "total"
	MemoryStats_Used    = "used"

	InterfaceStatsPrefix         = "/if/"
	InterfaceStats_Names         = InterfaceStatsPrefix + "names"
	InterfaceStats_Drops         = InterfaceStatsPrefix + "drops"
	InterfaceStats_Punt          = InterfaceStatsPrefix + "punt"
	InterfaceStats_IP4           = InterfaceStatsPrefix + "ip4"
	InterfaceStats_IP6           = InterfaceStatsPrefix + "ip6"
	InterfaceStats_RxNoBuf       = InterfaceStatsPrefix + "rx-no-buf"
	InterfaceStats_RxMiss        = InterfaceStatsPrefix + "rx-miss"
	InterfaceStats_RxError       = InterfaceStatsPrefix + "rx-error"
	InterfaceStats_TxError       = InterfaceStatsPrefix + "tx-error"
	InterfaceStats_Mpls          = InterfaceStatsPrefix + "mpls"
	InterfaceStats_Rx            = InterfaceStatsPrefix + "rx"
	InterfaceStats_RxUnicast     = InterfaceStatsPrefix + "rx-unicast"
	InterfaceStats_RxMulticast   = InterfaceStatsPrefix + "rx-multicast"
	InterfaceStats_RxBroadcast   = InterfaceStatsPrefix + "rx-broadcast"
	InterfaceStats_Tx            = InterfaceStatsPrefix + "tx"
	InterfaceStats_TxUnicast     = InterfaceStatsPrefix + "tx-unicast"
	InterfaceStats_TxUnicastMiss = InterfaceStatsPrefix + "tx-unicast-miss"
	InterfaceStats_TxMulticast   = InterfaceStatsPrefix + "tx-multicast"
	InterfaceStats_TxBroadcast   = InterfaceStatsPrefix + "tx-broadcast"

	// TODO: network stats
	NetworkStatsPrefix     = "/net/"
	NetworkStats_RouteTo   = NetworkStatsPrefix + "route/to"
	NetworkStats_RouteVia  = NetworkStatsPrefix + "route/via"
	NetworkStats_MRoute    = NetworkStatsPrefix + "mroute"
	NetworkStats_Adjacency = NetworkStatsPrefix + "adjacency"
	NetworkStats_Punt      = NetworkStatsPrefix + "punt"
)

type StatsConnection struct {
	statsClient adapter.StatsAPI

	maxAttempts int           // interval for reconnect attempts
	recInterval time.Duration // maximum number of reconnect attempts

	connChan chan ConnectionEvent // connection event channel
	done     chan struct{}        // to terminate stats connection watcher

	errorStatsData *adapter.StatDir
	nodeStatsData  *adapter.StatDir
	ifaceStatsData *adapter.StatDir
	sysStatsData   *adapter.StatDir
	bufStatsData   *adapter.StatDir
	memStatsData   *adapter.StatDir
}

func newStatsConnection(stats adapter.StatsAPI, attempts int, interval time.Duration) *StatsConnection {
	if attempts == 0 {
		attempts = DefaultMaxReconnectAttempts
	}
	if interval == 0 {
		interval = DefaultReconnectInterval
	}

	return &StatsConnection{
		statsClient: stats,
		maxAttempts: attempts,
		recInterval: interval,
		connChan:    make(chan ConnectionEvent, NotificationChanBufSize),
		done:        make(chan struct{}),
	}
}

// ConnectStats connects to Stats API using specified adapter and returns a connection handle.
// This call blocks until it is either connected, or an error occurs.
// Only one connection attempt will be performed.
func ConnectStats(stats adapter.StatsAPI) (*StatsConnection, error) {
	log.Debug("Connecting to stats..")
	c := newStatsConnection(stats, DefaultMaxReconnectAttempts, DefaultReconnectInterval)

	if err := c.statsClient.Connect(); err != nil {
		return nil, err
	}
	log.Debugf("Connected to stats.")

	return c, nil
}

// AsyncConnectStats  connects to the VPP stats socket asynchronously and returns the connection
// handle with state channel. The call is non-blocking and the caller is expected to watch ConnectionEvent
// values from the channel and wait for connect/disconnect events. Connection loop tries to reconnect the
// socket in case the session was disconnected.
func AsyncConnectStats(stats adapter.StatsAPI, attempts int, interval time.Duration) (*StatsConnection, chan ConnectionEvent, error) {
	log.Debug("Connecting to stats asynchronously..")
	c := newStatsConnection(stats, attempts, interval)

	go c.connectLoop()

	return c, c.connChan, nil
}

func (c *StatsConnection) connectLoop() {
	log.Debug("Asynchronously connecting to stats..")
	var reconnectAttempts int

	// loop until connected
	for {
		if err := c.statsClient.Connect(); err == nil {
			c.sendStatsConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Connected})
			break
		} else if reconnectAttempts < c.maxAttempts {
			reconnectAttempts++
			log.Warnf("connecting stats failed (attempt %d/%d): %v", reconnectAttempts, c.maxAttempts, err)
			time.Sleep(c.recInterval)
		} else {
			c.sendStatsConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Failed, Error: err})
			return
		}
	}
	// start monitoring stats connection state
	go c.monitorSocket()
}

// Disconnect disconnects from Stats API and releases all connection-related resources.
func (c *StatsConnection) Disconnect() {
	if c == nil {
		return
	}
	if c.statsClient != nil {
		if err := c.statsClient.Disconnect(); err != nil {
			log.Debugf("disconnecting stats client failed: %v", err)
		}
	}
	close(c.done)
}

func (c *StatsConnection) monitorSocket() {
	defer close(c.connChan)

	var state, lastState ConnectionState
	ticker := time.NewTicker(HealthCheckInterval)

	for {
		select {
		case <-ticker.C:
			_, err := c.statsClient.ListStats(SystemStats_Heartbeat)
			state = Connected
			if err == adapter.ErrStatsDataBusy {
				state = NotResponding
			}
			if err == adapter.ErrStatsDisconnected {
				state = Disconnected
			}
			if err == adapter.ErrStatsAccessFailed {
				state = Failed
			}
			if state == lastState {
				continue
			}
			lastState = state
			c.sendStatsConnEvent(ConnectionEvent{Timestamp: time.Now(), State: state, Error: err})
		case <-c.done:
			log.Debugf("health check watcher closed")
			c.sendStatsConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Disconnected, Error: nil})
			return
		}
	}
}

func (c *StatsConnection) updateStats(statDir **adapter.StatDir, patterns ...string) error {
	if statDir == nil {
		panic("statDir must not nil")
	}
	try := func() error {
		if (*statDir) == nil {
			dir, err := c.statsClient.PrepareDir(patterns...)
			if err != nil {
				log.Debugln("preparing dir failed:", err)
				return err
			}
			*statDir = dir
		} else {
			if err := c.statsClient.UpdateDir(*statDir); err != nil {
				log.Debugln("updating dir failed:", err)
				*statDir = nil
				return err
			}
		}

		return nil
	}
	var err error
	for r := 0; r < RetryUpdateCount; r++ {
		if err = try(); err == nil {
			if r > 0 {
				log.Debugf("retry successfull (r=%d)", r)
			}
			return nil
		} else if err == adapter.ErrStatsDirStale || err == adapter.ErrStatsDataBusy {
			// retrying
			if r > 1 {
				log.Debugf("sleeping for %v before next try", RetryUpdateDelay)
				time.Sleep(RetryUpdateDelay)
			}
		} else {
			// error is not retryable
			break
		}
	}
	return err
}

// GetSystemStats retrieves VPP system stats.
func (c *StatsConnection) GetSystemStats(sysStats *api.SystemStats) (err error) {
	if err := c.updateStats(&c.sysStatsData, SystemStatsPrefix); err != nil {
		return err
	}

	for _, stat := range c.sysStatsData.Entries {
		var val uint64
		if s, ok := stat.Data.(adapter.ScalarStat); ok {
			val = uint64(s)
		}
		switch string(stat.Name) {
		case SystemStats_VectorRate:
			sysStats.VectorRate = val
		case SystemStats_NumWorkerThreads:
			sysStats.NumWorkerThreads = val
		case SystemStats_VectorRatePerWorker:
			var vals []uint64
			if ss, ok := stat.Data.(adapter.SimpleCounterStat); ok {
				vals = make([]uint64, len(ss))
				for w := range ss {
					if ss[w] == nil {
						continue
					}
					vals[w] = uint64(ss[w][0])
				}
			}
			sysStats.VectorRatePerWorker = vals
		case SystemStats_InputRate:
			sysStats.InputRate = val
		case SystemStats_LastUpdate:
			sysStats.LastUpdate = val
		case SystemStats_LastStatsClear:
			sysStats.LastStatsClear = val
		case SystemStats_Heartbeat:
			sysStats.Heartbeat = val
		}
	}

	return nil
}

// GetErrorStats retrieves VPP error stats.
func (c *StatsConnection) GetErrorStats(errorStats *api.ErrorStats) (err error) {
	if err := c.updateStats(&c.errorStatsData, CounterStatsPrefix); err != nil {
		return err
	}

	if errorStats.Errors == nil || len(errorStats.Errors) != len(c.errorStatsData.Entries) {
		errorStats.Errors = make([]api.ErrorCounter, len(c.errorStatsData.Entries))
		for i := 0; i < len(c.errorStatsData.Entries); i++ {
			errorStats.Errors[i].CounterName = string(c.errorStatsData.Entries[i].Name)
		}
	}

	for i, stat := range c.errorStatsData.Entries {
		if errStat, ok := stat.Data.(adapter.ErrorStat); ok {
			values := make([]uint64, len(errStat))
			for j, errStatW := range errStat {
				values[j] = uint64(errStatW)
			}
			errorStats.Errors[i].Values = values
		}
		if errStat, ok := stat.Data.(adapter.SimpleCounterStat); ok {
			values := make([]uint64, len(errStat))
			for j, errStatW := range errStat {
				for _, val := range errStatW {
					values[j] += uint64(val)
				}
			}
			errorStats.Errors[i].Values = values
		}
	}
	return nil
}

func (c *StatsConnection) GetNodeStats(nodeStats *api.NodeStats) (err error) {
	if err := c.updateStats(&c.nodeStatsData, NodeStatsPrefix); err != nil {
		return err
	}

	prepNodes := func(l int) {
		if nodeStats.Nodes == nil || len(nodeStats.Nodes) != l {
			nodeStats.Nodes = make([]api.NodeCounters, l)
			for i := 0; i < l; i++ {
				nodeStats.Nodes[i].NodeIndex = uint32(i)
			}
		}
	}
	perNode := func(stat adapter.StatEntry, fn func(*api.NodeCounters, uint64)) {
		if s, ok := stat.Data.(adapter.SimpleCounterStat); ok {
			prepNodes(len(s[0]))
			for i := range nodeStats.Nodes {
				val := adapter.ReduceSimpleCounterStatIndex(s, i)
				fn(&nodeStats.Nodes[i], val)
			}
		}
	}

	for _, stat := range c.nodeStatsData.Entries {
		switch string(stat.Name) {
		case NodeStats_Names:
			if stat, ok := stat.Data.(adapter.NameStat); ok {
				prepNodes(len(stat))
				for i, nc := range nodeStats.Nodes {
					if nc.NodeName != string(stat[i]) {
						nc.NodeName = string(stat[i])
						nodeStats.Nodes[i] = nc
					}
				}
			}
		case NodeStats_Clocks:
			perNode(stat, func(node *api.NodeCounters, val uint64) {
				node.Clocks = val
			})
		case NodeStats_Vectors:
			perNode(stat, func(node *api.NodeCounters, val uint64) {
				node.Vectors = val
			})
		case NodeStats_Calls:
			perNode(stat, func(node *api.NodeCounters, val uint64) {
				node.Calls = val
			})
		case NodeStats_Suspends:
			perNode(stat, func(node *api.NodeCounters, val uint64) {
				node.Suspends = val
			})
		}
	}

	return nil
}

// GetInterfaceStats retrieves VPP per interface stats.
func (c *StatsConnection) GetInterfaceStats(ifaceStats *api.InterfaceStats) (err error) {
	if err := c.updateStats(&c.ifaceStatsData, InterfaceStatsPrefix); err != nil {
		return err
	}

	prep := func(l int) {
		if ifaceStats.Interfaces == nil || len(ifaceStats.Interfaces) != l {
			ifaceStats.Interfaces = make([]api.InterfaceCounters, l)
			for i := 0; i < l; i++ {
				ifaceStats.Interfaces[i].InterfaceIndex = uint32(i)
			}
		}
	}
	perNode := func(stat adapter.StatEntry, fn func(*api.InterfaceCounters, uint64)) {
		if s, ok := stat.Data.(adapter.SimpleCounterStat); ok {
			prep(len(s[0]))
			for i := range ifaceStats.Interfaces {
				val := adapter.ReduceSimpleCounterStatIndex(s, i)
				fn(&ifaceStats.Interfaces[i], val)
			}
		}
	}
	perNodeComb := func(stat adapter.StatEntry, fn func(*api.InterfaceCounters, [2]uint64)) {
		if s, ok := stat.Data.(adapter.CombinedCounterStat); ok {
			prep(len(s[0]))
			for i := range ifaceStats.Interfaces {
				val := adapter.ReduceCombinedCounterStatIndex(s, i)
				fn(&ifaceStats.Interfaces[i], val)
			}
		}
	}

	for _, stat := range c.ifaceStatsData.Entries {
		switch string(stat.Name) {
		case InterfaceStats_Names:
			if stat, ok := stat.Data.(adapter.NameStat); ok {
				prep(len(stat))
				for i, nc := range ifaceStats.Interfaces {
					if nc.InterfaceName != string(stat[i]) {
						nc.InterfaceName = string(stat[i])
						ifaceStats.Interfaces[i] = nc
					}
				}
			}
		case InterfaceStats_Drops:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.Drops = val
			})
		case InterfaceStats_Punt:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.Punts = val
			})
		case InterfaceStats_IP4:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.IP4 = val
			})
		case InterfaceStats_IP6:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.IP6 = val
			})
		case InterfaceStats_RxNoBuf:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.RxNoBuf = val
			})
		case InterfaceStats_RxMiss:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.RxMiss = val
			})
		case InterfaceStats_RxError:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.RxErrors = val
			})
		case InterfaceStats_TxError:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.TxErrors = val
			})
		case InterfaceStats_Mpls:
			perNode(stat, func(iface *api.InterfaceCounters, val uint64) {
				iface.Mpls = val
			})
		case InterfaceStats_Rx:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.Rx.Packets = val[0]
				iface.Rx.Bytes = val[1]
			})
		case InterfaceStats_RxUnicast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.RxUnicast.Packets = val[0]
				iface.RxUnicast.Bytes = val[1]
			})
		case InterfaceStats_RxMulticast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.RxMulticast.Packets = val[0]
				iface.RxMulticast.Bytes = val[1]
			})
		case InterfaceStats_RxBroadcast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.RxBroadcast.Packets = val[0]
				iface.RxBroadcast.Bytes = val[1]
			})
		case InterfaceStats_Tx:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.Tx.Packets = val[0]
				iface.Tx.Bytes = val[1]
			})
		case InterfaceStats_TxUnicastMiss:
			// tx-unicast-miss was a spelling mistake in older versions
			//
			fallthrough
		case InterfaceStats_TxUnicast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.TxUnicast.Packets = val[0]
				iface.TxUnicast.Bytes = val[1]
			})
		case InterfaceStats_TxMulticast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.TxMulticast.Packets = val[0]
				iface.TxMulticast.Bytes = val[1]
			})
		case InterfaceStats_TxBroadcast:
			perNodeComb(stat, func(iface *api.InterfaceCounters, val [2]uint64) {
				iface.TxBroadcast.Packets = val[0]
				iface.TxBroadcast.Bytes = val[1]
			})
		}
	}

	return nil
}

// GetBufferStats retrieves VPP buffer pools stats.
func (c *StatsConnection) GetBufferStats(bufStats *api.BufferStats) (err error) {
	if err := c.updateStats(&c.bufStatsData, BufferStatsPrefix); err != nil {
		return err
	}

	if bufStats.Buffer == nil {
		bufStats.Buffer = make(map[string]api.BufferPool)
	}

	for _, stat := range c.bufStatsData.Entries {
		d, f := path.Split(string(stat.Name))
		d = strings.TrimSuffix(d, "/")

		name := strings.TrimPrefix(d, BufferStatsPrefix)
		b, ok := bufStats.Buffer[name]
		if !ok {
			b.PoolName = name
		}

		var val float64
		s, ok := stat.Data.(adapter.ScalarStat)
		if ok {
			val = float64(s)
		}
		switch f {
		case BufferStats_Cached:
			b.Cached = val
		case BufferStats_Used:
			b.Used = val
		case BufferStats_Available:
			b.Available = val
		}

		bufStats.Buffer[name] = b
	}

	return nil
}

func (c *StatsConnection) GetMemoryStats(memStats *api.MemoryStats) (err error) {
	if err := c.updateStats(&c.memStatsData, MemoryStatSegPrefix, MemoryStatSegment, MemoryMainHeap); err != nil {
		return err
	}
	convertStats := func(stats []adapter.Counter) api.MemoryCounters {
		memUsg := make([]adapter.Counter, 7)
		copy(memUsg, stats)
		return api.MemoryCounters{
			Total: uint64(memUsg[0]), Used: uint64(memUsg[1]), Free: uint64(memUsg[2]), UsedMMap: uint64(memUsg[3]),
			TotalAlloc: uint64(memUsg[4]), FreeChunks: uint64(memUsg[5]), Releasable: uint64(memUsg[6]),
		}
	}

	for _, stat := range c.memStatsData.Entries {
		if strings.Contains(string(stat.Name), MemoryStatSegPrefix) {
			_, f := path.Split(string(stat.Name))
			var val float64
			m, ok := stat.Data.(adapter.ScalarStat)
			if ok {
				val = float64(m)
			}
			switch f {
			case MemoryStats_Total:
				memStats.Total = val //nolint:staticcheck
			case MemoryStats_Used:
				memStats.Used = val //nolint:staticcheck
			}
		} else if string(stat.Name) == MemoryStatSegment {
			if perHeapStats, ok := stat.Data.(adapter.SimpleCounterStat); ok {
				if memStats.Stat == nil {
					memStats.Stat = make(map[int]api.MemoryCounters)
				}
				for heap, stats := range perHeapStats {
					memStats.Stat[heap] = convertStats(stats)
				}
			}
		} else if string(stat.Name) == MemoryMainHeap {
			if perHeapStats, ok := stat.Data.(adapter.SimpleCounterStat); ok {
				if memStats.Main == nil {
					memStats.Main = make(map[int]api.MemoryCounters)
				}
				for heap, stats := range perHeapStats {
					memStats.Main[heap] = convertStats(stats)
				}
			}
		}
	}
	return nil
}

func (c *StatsConnection) sendStatsConnEvent(event ConnectionEvent) {
	select {
	case c.connChan <- event:
	default:
		log.Warn("Stats connection state channel is full, discarding value.")
	}
}
