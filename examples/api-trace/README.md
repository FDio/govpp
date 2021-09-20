# API trace example

The example demonstrates how to use GoVPP API trace functionality. Connection object `core.Connection` contains
API tracer able to record API messages sent to and from VPP.

Access to the tracer is done via `Trace()`. It allows accessing several methods to manage collected entries:
* `Enable(<bool>)` either enables or disables the trace. Note that the trace is disabled by default and messages are not recorded while so.
* `GetRecords() []*api.Record` provide messages collected since the plugin was enabled or cleared.
* `GetRecordsForChannel(<channelID>) []*api.Record` provide messages collected on the given channel since the plugin was enabled or cleared.
* `Clear()` removes recorded messages.

A record is represented by `Record` type. It contains information about the message, its direction, time and channel ID. Following fields are available:
* `Message api.Message` returns recorded entry as GoVPP Message.
* `Timestamp time.Time` is the message timestamp.
* `IsReceived bool` is true if the message is a reply or details message, false otherwise.
* `ChannelID uint16` is the ID of channel processing the traced message. 

