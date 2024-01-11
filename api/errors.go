package api

import (
	"fmt"
	"strconv"
)

// CompatibilityError is the error type usually returned by CheckCompatibility
// method of Channel. It contains list of all the compatible/incompatible messages.
type CompatibilityError struct {
	CompatibleMessages   []string
	IncompatibleMessages []string
}

func (c *CompatibilityError) Error() string {
	return fmt.Sprintf("%d/%d messages incompatible", len(c.IncompatibleMessages), len(c.CompatibleMessages)+len(c.IncompatibleMessages))
}

// RetvalToVPPApiError returns error for retval value.
// Retval 0 returns nil error.
func RetvalToVPPApiError(retval int32) error {
	if retval == 0 {
		return nil
	}
	return VPPApiError(retval)
}

// VPPApiError represents VPP's vnet API error that is usually
// returned as Retval field in replies from VPP binary API.
type VPPApiError int32

func (e VPPApiError) Error() string {
	errid := int64(e)
	var errstr string
	if s, ok := vppApiErrors[e]; ok {
		errstr = fmt.Sprintf("%s (%d)", s, errid)
	} else {
		errstr = strconv.FormatInt(errid, 10)
	}
	return fmt.Sprintf("VPPApiError: %s", errstr)
}

// definitions from: vpp/src/vnet/error.h
const (
	_                                  VPPApiError = 0
	UNSPECIFIED                        VPPApiError = -1
	INVALID_SW_IF_INDEX                VPPApiError = -2
	NO_SUCH_FIB                        VPPApiError = -3
	NO_SUCH_INNER_FIB                  VPPApiError = -4
	NO_SUCH_LABEL                      VPPApiError = -5
	NO_SUCH_ENTRY                      VPPApiError = -6
	INVALID_VALUE                      VPPApiError = -7
	INVALID_VALUE_2                    VPPApiError = -8
	UNIMPLEMENTED                      VPPApiError = -9
	INVALID_SW_IF_INDEX_2              VPPApiError = -10
	SYSCALL_ERROR_1                    VPPApiError = -11
	SYSCALL_ERROR_2                    VPPApiError = -12
	SYSCALL_ERROR_3                    VPPApiError = -13
	SYSCALL_ERROR_4                    VPPApiError = -14
	SYSCALL_ERROR_5                    VPPApiError = -15
	SYSCALL_ERROR_6                    VPPApiError = -16
	SYSCALL_ERROR_7                    VPPApiError = -17
	SYSCALL_ERROR_8                    VPPApiError = -18
	SYSCALL_ERROR_9                    VPPApiError = -19
	SYSCALL_ERROR_10                   VPPApiError = -20
	FEATURE_DISABLED                   VPPApiError = -30
	INVALID_REGISTRATION               VPPApiError = -31
	NEXT_HOP_NOT_IN_FIB                VPPApiError = -50
	UNKNOWN_DESTINATION                VPPApiError = -51
	PREFIX_MATCHES_NEXT_HOP            VPPApiError = -52
	NEXT_HOP_NOT_FOUND_MP              VPPApiError = -53
	NO_MATCHING_INTERFACE              VPPApiError = -54
	INVALID_VLAN                       VPPApiError = -55
	VLAN_ALREADY_EXISTS                VPPApiError = -56
	INVALID_SRC_ADDRESS                VPPApiError = -57
	INVALID_DST_ADDRESS                VPPApiError = -58
	ADDRESS_LENGTH_MISMATCH            VPPApiError = -59
	ADDRESS_NOT_FOUND_FOR_INTERFACE    VPPApiError = -60
	ADDRESS_NOT_DELETABLE              VPPApiError = -61
	IP6_NOT_ENABLED                    VPPApiError = -62
	NO_SUCH_NODE                       VPPApiError = -63
	NO_SUCH_NODE2                      VPPApiError = -64
	NO_SUCH_TABLE                      VPPApiError = -65
	NO_SUCH_TABLE2                     VPPApiError = -66
	NO_SUCH_TABLE3                     VPPApiError = -67
	SUBIF_ALREADY_EXISTS               VPPApiError = -68
	SUBIF_CREATE_FAILED                VPPApiError = -69
	INVALID_MEMORY_SIZE                VPPApiError = -70
	INVALID_INTERFACE                  VPPApiError = -71
	INVALID_VLAN_TAG_COUNT             VPPApiError = -72
	INVALID_ARGUMENT                   VPPApiError = -73
	UNEXPECTED_INTF_STATE              VPPApiError = -74
	TUNNEL_EXIST                       VPPApiError = -75
	INVALID_DECAP_NEXT                 VPPApiError = -76
	RESPONSE_NOT_READY                 VPPApiError = -77
	NOT_CONNECTED                      VPPApiError = -78
	IF_ALREADY_EXISTS                  VPPApiError = -79
	BOND_SLAVE_NOT_ALLOWED             VPPApiError = -80
	VALUE_EXIST                        VPPApiError = -81
	SAME_SRC_DST                       VPPApiError = -82
	IP6_MULTICAST_ADDRESS_NOT_PRESENT  VPPApiError = -83
	SR_POLICY_NAME_NOT_PRESENT         VPPApiError = -84
	NOT_RUNNING_AS_ROOT                VPPApiError = -85
	ALREADY_CONNECTED                  VPPApiError = -86
	UNSUPPORTED_JNI_VERSION            VPPApiError = -87
	FAILED_TO_ATTACH_TO_JAVA_THREAD    VPPApiError = -88
	INVALID_WORKER                     VPPApiError = -89
	LISP_DISABLED                      VPPApiError = -90
	CLASSIFY_TABLE_NOT_FOUND           VPPApiError = -91
	INVALID_EID_TYPE                   VPPApiError = -92
	CANNOT_CREATE_PCAP_FILE            VPPApiError = -93
	INCORRECT_ADJACENCY_TYPE           VPPApiError = -94
	EXCEEDED_NUMBER_OF_RANGES_CAPACITY VPPApiError = -95
	EXCEEDED_NUMBER_OF_PORTS_CAPACITY  VPPApiError = -96
	INVALID_ADDRESS_FAMILY             VPPApiError = -97
	INVALID_SUB_SW_IF_INDEX            VPPApiError = -98
	TABLE_TOO_BIG                      VPPApiError = -99
	CANNOT_ENABLE_DISABLE_FEATURE      VPPApiError = -100
	BFD_EEXIST                         VPPApiError = -101
	BFD_ENOENT                         VPPApiError = -102
	BFD_EINUSE                         VPPApiError = -103
	BFD_NOTSUPP                        VPPApiError = -104
	ADDRESS_IN_USE                     VPPApiError = -105
	ADDRESS_NOT_IN_USE                 VPPApiError = -106
	QUEUE_FULL                         VPPApiError = -107
	APP_UNSUPPORTED_CFG                VPPApiError = -108
	URI_FIFO_CREATE_FAILED             VPPApiError = -109
	LISP_RLOC_LOCAL                    VPPApiError = -110
	BFD_EAGAIN                         VPPApiError = -111
	INVALID_GPE_MODE                   VPPApiError = -112
	LISP_GPE_ENTRIES_PRESENT           VPPApiError = -113
	ADDRESS_FOUND_FOR_INTERFACE        VPPApiError = -114
	SESSION_CONNECT                    VPPApiError = -115
	ENTRY_ALREADY_EXISTS               VPPApiError = -116
	SVM_SEGMENT_CREATE_FAIL            VPPApiError = -117
	APPLICATION_NOT_ATTACHED           VPPApiError = -118
	BD_ALREADY_EXISTS                  VPPApiError = -119
	BD_IN_USE                          VPPApiError = -120
	BD_NOT_MODIFIABLE                  VPPApiError = -121
	BD_ID_EXCEED_MAX                   VPPApiError = -122
	SUBIF_DOESNT_EXIST                 VPPApiError = -123
	L2_MACS_EVENT_CLINET_PRESENT       VPPApiError = -124
	INVALID_QUEUE                      VPPApiError = -125
	UNSUPPORTED                        VPPApiError = -126
	DUPLICATE_IF_ADDRESS               VPPApiError = -127
	APP_INVALID_NS                     VPPApiError = -128
	APP_WRONG_NS_SECRET                VPPApiError = -129
	APP_CONNECT_SCOPE                  VPPApiError = -130
	APP_ALREADY_ATTACHED               VPPApiError = -131
	SESSION_REDIRECT                   VPPApiError = -132
	ILLEGAL_NAME                       VPPApiError = -133
	NO_NAME_SERVERS                    VPPApiError = -134
	NAME_SERVER_NOT_FOUND              VPPApiError = -135
	NAME_RESOLUTION_NOT_ENABLED        VPPApiError = -136
	NAME_SERVER_FORMAT_ERROR           VPPApiError = -137
	NAME_SERVER_NO_SUCH_NAME           VPPApiError = -138
	NAME_SERVER_NO_ADDRESSES           VPPApiError = -139
	NAME_SERVER_NEXT_SERVER            VPPApiError = -140
	APP_CONNECT_FILTERED               VPPApiError = -141
	ACL_IN_USE_INBOUND                 VPPApiError = -142
	ACL_IN_USE_OUTBOUND                VPPApiError = -143
	INIT_FAILED                        VPPApiError = -144
	NETLINK_ERROR                      VPPApiError = -145
	BIER_BSL_UNSUP                     VPPApiError = -146
	INSTANCE_IN_USE                    VPPApiError = -147
	INVALID_SESSION_ID                 VPPApiError = -148
	ACL_IN_USE_BY_LOOKUP_CONTEXT       VPPApiError = -149
	INVALID_VALUE_3                    VPPApiError = -150
	NON_ETHERNET                       VPPApiError = -151
	BD_ALREADY_HAS_BVI                 VPPApiError = -152
	INVALID_PROTOCOL                   VPPApiError = -153
	INVALID_ALGORITHM                  VPPApiError = -154
	RSRC_IN_USE                        VPPApiError = -155
	KEY_LENGTH                         VPPApiError = -156
	FIB_PATH_UNSUPPORTED_NH_PROTO      VPPApiError = -157
	API_ENDIAN_FAILED                  VPPApiError = -159
	NO_CHANGE                          VPPApiError = -160
	MISSING_CERT_KEY                   VPPApiError = -161
	LIMIT_EXCEEDED                     VPPApiError = -162
	IKE_NO_PORT                        VPPApiError = -163
	UDP_PORT_TAKEN                     VPPApiError = -164
	EAGAIN                             VPPApiError = -165
	INVALID_VALUE_4                    VPPApiError = -166
	BUSY                               VPPApiError = -167
	BUG                                VPPApiError = -168
	FEATURE_ALREADY_DISABLED           VPPApiError = -169
	FEATURE_ALREADY_ENABLED            VPPApiError = -170
	INVALID_PREFIX_LENGTH              VPPApiError = -171
)

var vppApiErrors = map[VPPApiError]string{
	UNSPECIFIED:                        "Unspecified Error",
	INVALID_SW_IF_INDEX:                "Invalid sw_if_index",
	NO_SUCH_FIB:                        "No such FIB / VRF",
	NO_SUCH_INNER_FIB:                  "No such inner FIB / VRF",
	NO_SUCH_LABEL:                      "No such label",
	NO_SUCH_ENTRY:                      "No such entry",
	INVALID_VALUE:                      "Invalid value",
	INVALID_VALUE_2:                    "Invalid value #2",
	UNIMPLEMENTED:                      "Unimplemented",
	INVALID_SW_IF_INDEX_2:              "Invalid sw_if_index #2",
	SYSCALL_ERROR_1:                    "System call error #1",
	SYSCALL_ERROR_2:                    "System call error #2",
	SYSCALL_ERROR_3:                    "System call error #3",
	SYSCALL_ERROR_4:                    "System call error #4",
	SYSCALL_ERROR_5:                    "System call error #5",
	SYSCALL_ERROR_6:                    "System call error #6",
	SYSCALL_ERROR_7:                    "System call error #7",
	SYSCALL_ERROR_8:                    "System call error #8",
	SYSCALL_ERROR_9:                    "System call error #9",
	SYSCALL_ERROR_10:                   "System call error #10",
	FEATURE_DISABLED:                   "Feature disabled by configuration",
	INVALID_REGISTRATION:               "Invalid registration",
	NEXT_HOP_NOT_IN_FIB:                "Next hop not in FIB",
	UNKNOWN_DESTINATION:                "Unknown destination",
	PREFIX_MATCHES_NEXT_HOP:            "Prefix matches next hop",
	NEXT_HOP_NOT_FOUND_MP:              "Next hop not found (multipath)",
	NO_MATCHING_INTERFACE:              "No matching interface for probe",
	INVALID_VLAN:                       "Invalid VLAN",
	VLAN_ALREADY_EXISTS:                "VLAN subif already exists",
	INVALID_SRC_ADDRESS:                "Invalid src address",
	INVALID_DST_ADDRESS:                "Invalid dst address",
	ADDRESS_LENGTH_MISMATCH:            "Address length mismatch",
	ADDRESS_NOT_FOUND_FOR_INTERFACE:    "Address not found for interface",
	ADDRESS_NOT_DELETABLE:              "Address not deletable",
	IP6_NOT_ENABLED:                    "ip6 not enabled",
	NO_SUCH_NODE:                       "No such graph node",
	NO_SUCH_NODE2:                      "No such graph node #2",
	NO_SUCH_TABLE:                      "No such table",
	NO_SUCH_TABLE2:                     "No such table #2",
	NO_SUCH_TABLE3:                     "No such table #3",
	SUBIF_ALREADY_EXISTS:               "Subinterface already exists",
	SUBIF_CREATE_FAILED:                "Subinterface creation failed",
	INVALID_MEMORY_SIZE:                "Invalid memory size requested",
	INVALID_INTERFACE:                  "Invalid interface",
	INVALID_VLAN_TAG_COUNT:             "Invalid number of tags for requested operation",
	INVALID_ARGUMENT:                   "Invalid argument",
	UNEXPECTED_INTF_STATE:              "Unexpected interface state",
	TUNNEL_EXIST:                       "Tunnel already exists",
	INVALID_DECAP_NEXT:                 "Invalid decap-next",
	RESPONSE_NOT_READY:                 "Response not ready",
	NOT_CONNECTED:                      "Not connected to the data plane",
	IF_ALREADY_EXISTS:                  "Interface already exists",
	BOND_SLAVE_NOT_ALLOWED:             "Operation not allowed on slave of BondEthernet",
	VALUE_EXIST:                        "Value already exists",
	SAME_SRC_DST:                       "Source and destination are the same",
	IP6_MULTICAST_ADDRESS_NOT_PRESENT:  "IP6 multicast address required",
	SR_POLICY_NAME_NOT_PRESENT:         "Segement routing policy name required",
	NOT_RUNNING_AS_ROOT:                "Not running as root",
	ALREADY_CONNECTED:                  "Connection to the data plane already exists",
	UNSUPPORTED_JNI_VERSION:            "Unsupported JNI version",
	FAILED_TO_ATTACH_TO_JAVA_THREAD:    "Failed to attach to Java thread",
	INVALID_WORKER:                     "Invalid worker thread",
	LISP_DISABLED:                      "LISP is disabled",
	CLASSIFY_TABLE_NOT_FOUND:           "Classify table not found",
	INVALID_EID_TYPE:                   "Unsupported LSIP EID type",
	CANNOT_CREATE_PCAP_FILE:            "Cannot create pcap file",
	INCORRECT_ADJACENCY_TYPE:           "Invalid adjacency type for this operation",
	EXCEEDED_NUMBER_OF_RANGES_CAPACITY: "Operation would exceed configured capacity of ranges",
	EXCEEDED_NUMBER_OF_PORTS_CAPACITY:  "Operation would exceed capacity of number of ports",
	INVALID_ADDRESS_FAMILY:             "Invalid address family",
	INVALID_SUB_SW_IF_INDEX:            "Invalid sub-interface sw_if_index",
	TABLE_TOO_BIG:                      "Table too big",
	CANNOT_ENABLE_DISABLE_FEATURE:      "Cannot enable/disable feature",
	BFD_EEXIST:                         "Duplicate BFD object",
	BFD_ENOENT:                         "No such BFD object",
	BFD_EINUSE:                         "BFD object in use",
	BFD_NOTSUPP:                        "BFD feature not supported",
	ADDRESS_IN_USE:                     "Address in use",
	ADDRESS_NOT_IN_USE:                 "Address not in use",
	QUEUE_FULL:                         "Queue full",
	APP_UNSUPPORTED_CFG:                "Unsupported application config",
	URI_FIFO_CREATE_FAILED:             "URI FIFO segment create failed",
	LISP_RLOC_LOCAL:                    "RLOC address is local",
	BFD_EAGAIN:                         "BFD object cannot be manipulated at this time",
	INVALID_GPE_MODE:                   "Invalid GPE mode",
	LISP_GPE_ENTRIES_PRESENT:           "LISP GPE entries are present",
	ADDRESS_FOUND_FOR_INTERFACE:        "Address found for interface",
	SESSION_CONNECT:                    "Session failed to connect",
	ENTRY_ALREADY_EXISTS:               "Entry already exists",
	SVM_SEGMENT_CREATE_FAIL:            "svm segment create fail",
	APPLICATION_NOT_ATTACHED:           "application not attached",
	BD_ALREADY_EXISTS:                  "Bridge domain already exists",
	BD_IN_USE:                          "Bridge domain has member interfaces",
	BD_NOT_MODIFIABLE:                  "Bridge domain 0 can't be deleted/modified",
	BD_ID_EXCEED_MAX:                   "Bridge domain ID exceed 16M limit",
	SUBIF_DOESNT_EXIST:                 "Subinterface doesn't exist",
	L2_MACS_EVENT_CLINET_PRESENT:       "Client already exist for L2 MACs events",
	INVALID_QUEUE:                      "Invalid queue",
	UNSUPPORTED:                        "Unsupported",
	DUPLICATE_IF_ADDRESS:               "Address already present on another interface",
	APP_INVALID_NS:                     "Invalid application namespace",
	APP_WRONG_NS_SECRET:                "Wrong app namespace secret",
	APP_CONNECT_SCOPE:                  "Connect scope",
	APP_ALREADY_ATTACHED:               "App already attached",
	SESSION_REDIRECT:                   "Redirect failed",
	ILLEGAL_NAME:                       "Illegal name",
	NO_NAME_SERVERS:                    "No name servers configured",
	NAME_SERVER_NOT_FOUND:              "Name server not found",
	NAME_RESOLUTION_NOT_ENABLED:        "Name resolution not enabled",
	NAME_SERVER_FORMAT_ERROR:           "Server format error (bug!)",
	NAME_SERVER_NO_SUCH_NAME:           "No such name",
	NAME_SERVER_NO_ADDRESSES:           "No addresses available",
	NAME_SERVER_NEXT_SERVER:            "Retry with new server",
	APP_CONNECT_FILTERED:               "Connect was filtered",
	ACL_IN_USE_INBOUND:                 "Inbound ACL in use",
	ACL_IN_USE_OUTBOUND:                "Outbound ACL in use",
	INIT_FAILED:                        "Initialization Failed",
	NETLINK_ERROR:                      "netlink error",
	BIER_BSL_UNSUP:                     "BIER bit-string-length unsupported",
	INSTANCE_IN_USE:                    "Instance in use",
	INVALID_SESSION_ID:                 "session ID out of range",
	ACL_IN_USE_BY_LOOKUP_CONTEXT:       "ACL in use by a lookup context",
	INVALID_VALUE_3:                    "Invalid value #3",
	NON_ETHERNET:                       "Interface is not an Ethernet interface",
	BD_ALREADY_HAS_BVI:                 "Bridge domain already has a BVI interface",
	INVALID_PROTOCOL:                   "Invalid Protocol",
	INVALID_ALGORITHM:                  "Invalid Algorithm",
	RSRC_IN_USE:                        "Resource In Use",
	KEY_LENGTH:                         "invalid Key Length",
	FIB_PATH_UNSUPPORTED_NH_PROTO:      "Unsupported FIB Path protocol",
	API_ENDIAN_FAILED:                  "Endian mismatch detected",
	NO_CHANGE:                          "No change in table",
	MISSING_CERT_KEY:                   "Missing certifcate or key",
	LIMIT_EXCEEDED:                     "limit exceeded",
	IKE_NO_PORT:                        "port not managed by IKE",
	UDP_PORT_TAKEN:                     "UDP port already taken",
	EAGAIN:                             "Retry stream call with cursor",
	INVALID_VALUE_4:                    "Invalid value #4",
	BUSY:                               "Busy",
	BUG:                                "Bug",
	FEATURE_ALREADY_DISABLED:           "Feature already disabled",
	FEATURE_ALREADY_ENABLED:            "Feature already enabled",
	INVALID_PREFIX_LENGTH:              "Invalid prefix length",
}
