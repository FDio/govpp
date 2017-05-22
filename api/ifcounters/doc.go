// Package ifcounters provides the helper API for decoding VnetInterfaceCounters binary API message
// that contains binary-encoded statistics data into the Go structs that are better consumable by the Go code.
//
// VPP provides two types of interface counters that can be encoded inside of a single VnetInterfaceCounters
// message: simple and combined. For both of them, ifcounters API provides a separate decode function:
// DecodeCounters or DecodeCombinedCounters. The functions return a slice of simple or combined counters respectively:
//
//	notifMsg := <-notifChan:
//	notif := notifMsg.(*interfaces.VnetInterfaceCounters)
//
//	if notif.IsCombined == 0 {
//		// simple counter
//		counters, err := ifcounters.DecodeCounters(ifcounters.VnetInterfaceCounters(*notif))
//		if err != nil {
//			fmt.Println("Error:", err)
//		} else {
//			fmt.Printf("%+v\n", counters)
//		}
//	} else {
//		// combined counter
//		counters, err := ifcounters.DecodeCombinedCounters(ifcounters.VnetInterfaceCounters(*notif))
//		if err != nil {
//			fmt.Println("Error:", err)
//		} else {
//			fmt.Printf("%+v\n", counters)
//		}
//	}
//
package ifcounters
