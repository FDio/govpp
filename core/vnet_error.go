package core

import "fmt"

type VnetError int32

/*
	definitions from: vpp/src/vnet/api_errno.h
*/
const (
	_                                  VnetError = 0
	UNSPECIFIED                                  = -1
	INVALID_SW_IF_INDEX                          = -2
	NO_SUCH_FIB                                  = -3
	NO_SUCH_INNER_FIB                            = -4
	NO_SUCH_LABEL                                = -5
	NO_SUCH_ENTRY                                = -6
	INVALID_VALUE                                = -7
	INVALID_VALUE_2                              = -8
	UNIMPLEMENTED                                = -9
	INVALID_SW_IF_INDEX_2                        = -10
	SYSCALL_ERROR_1                              = -11
	SYSCALL_ERROR_2                              = -12
	SYSCALL_ERROR_3                              = -13
	SYSCALL_ERROR_4                              = -14
	SYSCALL_ERROR_5                              = -15
	SYSCALL_ERROR_6                              = -16
	SYSCALL_ERROR_7                              = -17
	SYSCALL_ERROR_8                              = -18
	SYSCALL_ERROR_9                              = -19
	SYSCALL_ERROR_10                             = -20
	FEATURE_DISABLED                             = -30
	INVALID_REGISTRATION                         = -31
	NEXT_HOP_NOT_IN_FIB                          = -50
	UNKNOWN_DESTINATION                          = -51
	PREFIX_MATCHES_NEXT_HOP                      = -52
	NEXT_HOP_NOT_FOUND_MP                        = -53
	NO_MATCHING_INTERFACE                        = -54
	INVALID_VLAN                                 = -55
	VLAN_ALREADY_EXISTS                          = -56
	INVALID_SRC_ADDRESS                          = -57
	INVALID_DST_ADDRESS                          = -58
	ADDRESS_LENGTH_MISMATCH                      = -59
	ADDRESS_NOT_FOUND_FOR_INTERFACE              = -60
	ADDRESS_NOT_LINK_LOCAL                       = -61
	IP6_NOT_ENABLED                              = -62
	IN_PROGRESS                                  = 10
	NO_SUCH_NODE                                 = -63
	NO_SUCH_NODE2                                = -64
	NO_SUCH_TABLE                                = -65
	NO_SUCH_TABLE2                               = -66
	NO_SUCH_TABLE3                               = -67
	SUBIF_ALREADY_EXISTS                         = -68
	SUBIF_CREATE_FAILED                          = -69
	INVALID_MEMORY_SIZE                          = -70
	INVALID_INTERFACE                            = -71
	INVALID_VLAN_TAG_COUNT                       = -72
	INVALID_ARGUMENT                             = -73
	UNEXPECTED_INTF_STATE                        = -74
	TUNNEL_EXIST                                 = -75
	INVALID_DECAP_NEXT                           = -76
	RESPONSE_NOT_READY                           = -77
	NOT_CONNECTED                                = -78
	IF_ALREADY_EXISTS                            = -79
	BOND_SLAVE_NOT_ALLOWED                       = -80
	VALUE_EXIST                                  = -81
	SAME_SRC_DST                                 = -82
	IP6_MULTICAST_ADDRESS_NOT_PRESENT            = -83
	SR_POLICY_NAME_NOT_PRESENT                   = -84
	NOT_RUNNING_AS_ROOT                          = -85
	ALREADY_CONNECTED                            = -86
	UNSUPPORTED_JNI_VERSION                      = -87
	FAILED_TO_ATTACH_TO_JAVA_THREAD              = -88
	INVALID_WORKER                               = -89
	LISP_DISABLED                                = -90
	CLASSIFY_TABLE_NOT_FOUND                     = -91
	INVALID_EID_TYPE                             = -92
	CANNOT_CREATE_PCAP_FILE                      = -93
	INCORRECT_ADJACENCY_TYPE                     = -94
	EXCEEDED_NUMBER_OF_RANGES_CAPACITY           = -95
	EXCEEDED_NUMBER_OF_PORTS_CAPACITY            = -96
	INVALID_ADDRESS_FAMILY                       = -97
	INVALID_SUB_SW_IF_INDEX                      = -98
	TABLE_TOO_BIG                                = -99
	CANNOT_ENABLE_DISABLE_FEATURE                = -100
	BFD_EEXIST                                   = -101
	BFD_ENOENT                                   = -102
	BFD_EINUSE                                   = -103
	BFD_NOTSUPP                                  = -104
	ADDRESS_IN_USE                               = -105
	ADDRESS_NOT_IN_USE                           = -106
	QUEUE_FULL                                   = -107
	UNKNOWN_URI_TYPE                             = -108
	URI_FIFO_CREATE_FAILED                       = -109
	LISP_RLOC_LOCAL                              = -110
	BFD_EAGAIN                                   = -111
	INVALID_GPE_MODE                             = -112
	LISP_GPE_ENTRIES_PRESENT                     = -113
	ADDRESS_FOUND_FOR_INTERFACE                  = -114
	SESSION_CONNECT                              = -115
	ENTRY_ALREADY_EXISTS                         = -116
	SVM_SEGMENT_CREATE_FAIL                      = -117
	APPLICATION_NOT_ATTACHED                     = -118
	BD_ALREADY_EXISTS                            = -119
	BD_IN_USE                                    = -120
	BD_NOT_MODIFIABLE                            = -121
	BD_ID_EXCEED_MAX                             = -122
	SUBIF_DOESNT_EXIST                           = -123
	L2_MACS_EVENT_CLINET_PRESENT                 = -124
	INVALID_QUEUE                                = -125
	UNSUPPORTED                                  = -126
	DUPLICATE_IF_ADDRESS                         = -127
	APP_INVALID_NS                               = -128
	APP_WRONG_NS_SECRET                          = -129
	APP_CONNECT_SCOPE                            = -130
	APP_ALREADY_ATTACHED                         = -131
	SESSION_REDIRECT                             = -132
	ILLEGAL_NAME                                 = -133
	NO_NAME_SERVERS                              = -134
	NAME_SERVER_NOT_FOUND                        = -135
	NAME_RESOLUTION_NOT_ENABLED                  = -136
	NAME_SERVER_FORMAT_ERROR                     = -137
	NAME_SERVER_NO_SUCH_NAME                     = -138
	NAME_SERVER_NO_ADDRESSES                     = -139
	NAME_SERVER_NEXT_SERVER                      = -140
	APP_CONNECT_FILTERED                         = -141
	ACL_IN_USE_INBOUND                           = -142
	ACL_IN_USE_OUTBOUND                          = -143
	INIT_FAILED                                  = -144
	NETLINK_ERROR                                = -145
)

func (e VnetError) Error() string {
	switch e {
	case UNSPECIFIED:
		return "Unspecified Error"
	case INVALID_SW_IF_INDEX:
		return "Invalid sw_if_index"
	case NO_SUCH_FIB:
		return "No such FIB / VRF"
	case NO_SUCH_INNER_FIB:
		return "No such inner FIB / VRF"
	case NO_SUCH_LABEL:
		return "No such label"
	case NO_SUCH_ENTRY:
		return "No such entry"
	case INVALID_VALUE:
		return "Invalid value"
	case INVALID_VALUE_2:
		return "Invalid value #2"
	case UNIMPLEMENTED:
		return "Unimplemented"
	case INVALID_SW_IF_INDEX_2:
		return "Invalid sw_if_index #2"
	case SYSCALL_ERROR_1:
		return "System call error #1"
	case SYSCALL_ERROR_2:
		return "System call error #2"
	case SYSCALL_ERROR_3:
		return "System call error #3"
	case SYSCALL_ERROR_4:
		return "System call error #4"
	case SYSCALL_ERROR_5:
		return "System call error #5"
	case SYSCALL_ERROR_6:
		return "System call error #6"
	case SYSCALL_ERROR_7:
		return "System call error #7"
	case SYSCALL_ERROR_8:
		return "System call error #8"
	case SYSCALL_ERROR_9:
		return "System call error #9"
	case SYSCALL_ERROR_10:
		return "System call error #10"
	case FEATURE_DISABLED:
		return "Feature disabled by configuration"
	case INVALID_REGISTRATION:
		return "Invalid registration"
	case NEXT_HOP_NOT_IN_FIB:
		return "Next hop not in FIB"
	case UNKNOWN_DESTINATION:
		return "Unknown destination"
	case PREFIX_MATCHES_NEXT_HOP:
		return "Prefix matches next hop"
	case NEXT_HOP_NOT_FOUND_MP:
		return "Next hop not found (multipath)"
	case NO_MATCHING_INTERFACE:
		return "No matching interface for probe"
	case INVALID_VLAN:
		return "Invalid VLAN"
	case VLAN_ALREADY_EXISTS:
		return "VLAN subif already exists"
	case INVALID_SRC_ADDRESS:
		return "Invalid src address"
	case INVALID_DST_ADDRESS:
		return "Invalid dst address"
	case ADDRESS_LENGTH_MISMATCH:
		return "Address length mismatch"
	case ADDRESS_NOT_FOUND_FOR_INTERFACE:
		return "Address not found for interface"
	case ADDRESS_NOT_LINK_LOCAL:
		return "Address not link-local"
	case IP6_NOT_ENABLED:
		return "ip6 not enabled"
	case IN_PROGRESS:
		return "Operation in progress"
	case NO_SUCH_NODE:
		return "No such graph node"
	case NO_SUCH_NODE2:
		return "No such graph node #2"
	case NO_SUCH_TABLE:
		return "No such table"
	case NO_SUCH_TABLE2:
		return "No such table #2"
	case NO_SUCH_TABLE3:
		return "No such table #3"
	case SUBIF_ALREADY_EXISTS:
		return "Subinterface already exists"
	case SUBIF_CREATE_FAILED:
		return "Subinterface creation failed"
	case INVALID_MEMORY_SIZE:
		return "Invalid memory size requested"
	case INVALID_INTERFACE:
		return "Invalid interface"
	case INVALID_VLAN_TAG_COUNT:
		return "Invalid number of tags for requested operation"
	case INVALID_ARGUMENT:
		return "Invalid argument"
	case UNEXPECTED_INTF_STATE:
		return "Unexpected interface state"
	case TUNNEL_EXIST:
		return "Tunnel already exists"
	case INVALID_DECAP_NEXT:
		return "Invalid decap-next"
	case RESPONSE_NOT_READY:
		return "Response not ready"
	case NOT_CONNECTED:
		return "Not connected to the data plane"
	case IF_ALREADY_EXISTS:
		return "Interface already exists"
	case BOND_SLAVE_NOT_ALLOWED:
		return "Operation not allowed on slave of BondEthernet"
	case VALUE_EXIST:
		return "Value already exists"
	case SAME_SRC_DST:
		return "Source and destination are the same"
	case IP6_MULTICAST_ADDRESS_NOT_PRESENT:
		return "IP6 multicast address required"
	case SR_POLICY_NAME_NOT_PRESENT:
		return "Segement routing policy name required"
	case NOT_RUNNING_AS_ROOT:
		return "Not running as root"
	case ALREADY_CONNECTED:
		return "Connection to the data plane already exists"
	case UNSUPPORTED_JNI_VERSION:
		return "Unsupported JNI version"
	case FAILED_TO_ATTACH_TO_JAVA_THREAD:
		return "Failed to attach to Java thread"
	case INVALID_WORKER:
		return "Invalid worker thread"
	case LISP_DISABLED:
		return "LISP is disabled"
	case CLASSIFY_TABLE_NOT_FOUND:
		return "Classify table not found"
	case INVALID_EID_TYPE:
		return "Unsupported LSIP EID type"
	case CANNOT_CREATE_PCAP_FILE:
		return "Cannot create pcap file"
	case INCORRECT_ADJACENCY_TYPE:
		return "Invalid adjacency type for this operation"
	case EXCEEDED_NUMBER_OF_RANGES_CAPACITY:
		return "Operation would exceed configured capacity of ranges"
	case EXCEEDED_NUMBER_OF_PORTS_CAPACITY:
		return "Operation would exceed capacity of number of ports"
	case INVALID_ADDRESS_FAMILY:
		return "Invalid address family"
	case INVALID_SUB_SW_IF_INDEX:
		return "Invalid sub-interface sw_if_index"
	case TABLE_TOO_BIG:
		return "Table too big"
	case CANNOT_ENABLE_DISABLE_FEATURE:
		return "Cannot enable/disable feature"
	case BFD_EEXIST:
		return "Duplicate BFD object"
	case BFD_ENOENT:
		return "No such BFD object"
	case BFD_EINUSE:
		return "BFD object in use"
	case BFD_NOTSUPP:
		return "BFD feature not supported"
	case ADDRESS_IN_USE:
		return "Address in use"
	case ADDRESS_NOT_IN_USE:
		return "Address not in use"
	case QUEUE_FULL:
		return "Queue full"
	case UNKNOWN_URI_TYPE:
		return "Unknown URI type"
	case URI_FIFO_CREATE_FAILED:
		return "URI FIFO segment create failed"
	case LISP_RLOC_LOCAL:
		return "RLOC address is local"
	case BFD_EAGAIN:
		return "BFD object cannot be manipulated at this time"
	case INVALID_GPE_MODE:
		return "Invalid GPE mode"
	case LISP_GPE_ENTRIES_PRESENT:
		return "LISP GPE entries are present"
	case ADDRESS_FOUND_FOR_INTERFACE:
		return "Address found for interface"
	case SESSION_CONNECT:
		return "Session failed to connect"
	case ENTRY_ALREADY_EXISTS:
		return "Entry already exists"
	case SVM_SEGMENT_CREATE_FAIL:
		return "svm segment create fail"
	case APPLICATION_NOT_ATTACHED:
		return "application not attached"
	case BD_ALREADY_EXISTS:
		return "Bridge domain already exists"
	case BD_IN_USE:
		return "Bridge domain has member interfaces"
	case BD_NOT_MODIFIABLE:
		return "Bridge domain 0 can't be deleted/modified"
	case BD_ID_EXCEED_MAX:
		return "Bridge domain ID exceed 16M limit"
	case SUBIF_DOESNT_EXIST:
		return "Subinterface doesn't exist"
	case L2_MACS_EVENT_CLINET_PRESENT:
		return "Client already exist for L2 MACs events"
	case INVALID_QUEUE:
		return "Invalid queue"
	case UNSUPPORTED:
		return "Unsupported"
	case DUPLICATE_IF_ADDRESS:
		return "Address already present on another interface"
	case APP_INVALID_NS:
		return "Invalid application namespace"
	case APP_WRONG_NS_SECRET:
		return "Wrong app namespace secret"
	case APP_CONNECT_SCOPE:
		return "Connect scope"
	case APP_ALREADY_ATTACHED:
		return "App already attached"
	case SESSION_REDIRECT:
		return "Redirect failed"
	case ILLEGAL_NAME:
		return "Illegal name"
	case NO_NAME_SERVERS:
		return "No name servers configured"
	case NAME_SERVER_NOT_FOUND:
		return "Name server not found"
	case NAME_RESOLUTION_NOT_ENABLED:
		return "Name resolution not enabled"
	case NAME_SERVER_FORMAT_ERROR:
		return "Server format error (bug!)"
	case NAME_SERVER_NO_SUCH_NAME:
		return "No such name"
	case NAME_SERVER_NO_ADDRESSES:
		return "No addresses available"
	case NAME_SERVER_NEXT_SERVER:
		return "Retry with new server"
	case APP_CONNECT_FILTERED:
		return "Connect was filtered"
	case ACL_IN_USE_INBOUND:
		return "Inbound ACL in use"
	case ACL_IN_USE_OUTBOUND:
		return "Outbound ACL in use"
	case INIT_FAILED:
		return "Initialization Failed"
	case NETLINK_ERROR:
		return "netlink error"
	default:
		return fmt.Sprintf("unknown VnetError: %d", e)
	}
}
