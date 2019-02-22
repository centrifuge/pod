/*
Package p2p implements the Centrifuge wire protocol. Centrifuge wire protocol is used by p2p clients in the centrifuge
network to communicate about various protocol activities such as document consensus. This document is a specification of this wire protocol.

Background

In order to understand the design decisions behind the protocol one needs a quick background on libp2p and protobufs the tech stack used by the initial implementation of the protocol - `go-centrifuge`, though the wire protocol it self is independent
of this stack.

- Libp2p protocol multiplexing

Libp2p can  multiplex messages from multiple protocols in to the same `inet` stream. The protocols identifiers can be something like `/centrifuge/0.1` and `/http/v2`, one can register handlers for each of these protocols in the libp2p host and it will take care of routing the messages to the appropriate handler.

Further reading - https://github.com/libp2p/specs/blob/master/6-interfaces.md

- Protobufs

Centrifuge wire protocol uses google protobufs to encode the message envelopes used in the network. Its an efficient format for encoding/decoding as well as having good multi language support.

Further reading - https://developers.google.com/protocol-buffers/


The Protocol

1. Protocol Identifier

Centrifuge wire protocol uses libp2p protocol multiplexing as a way to differentiate protocol versions as well as the centrifuge ID that a particular message is targeted at in the centrifuge node.
i.e. at the start of the node protocol handlers are registered for each protocol version for each `centrifugeID` configured in the system. Given this requirement a centrifuge protocol identifier string must comply to the following format,
`/centrifuge/<version>/<centrifugeIDHex>`
Eg: `/centrifuge/0.0.1/0xf71876181dce`


2. Centrifuge Protocol Envelope

	+----------------------------------------------------+------------------------------+
	|Envelope length in varint   |predefined centrifuge  | message bytes                |
	|                            |message type           | (encoding depends on type)   |
	+----------------------------------------------------+------------------------------+


Once a message is forwarded to the stream handler registered for a particular protocol ID, The illustrated envelope needs to be decoded by the handler in order to process the message.

2.1 The message envelope is encoded as a protobuf.

	message P2PEnvelope {
	  // defines what type of message it is. (if we modify centrifuge-protobufs we can use type any with typeURL here)
	  MessageType type = 1;
	  // serialized (could be a protobuf) for the actual message
	  bytes body = 2;
	}

2.2 The encoded envelope is written to the stream using a `byte length delimited` encoding which allows stream decoder to iterate the number of bytes in the following stream that are relevant to the current message.  A Length varint of 2 means that the value is 2 byte length followed by the specified number of bytes of data.

2.3 Once the message has been decoded(unmarshalled) in to `MessageEnvelope` the handler(router) can identify the message type and forward to the relevant specific handler for the given message type. The message type in this case is also serves as a protocol multiplexer.

2.4 The actual message byte encoding depends on the message type as well, which the router can decide to decode or forward as is.
*/
package p2p
