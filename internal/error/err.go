package errors

var EthernetLenError = NewProtocolError("to ethernet buf len to min")
var IpLenError = NewProtocolError("to ip buf len to min")
var IcmpLenError = NewProtocolError("to icmp buf len to min")
var ErrRingEmpty = NewProtocolError("to icmp buf len to min")
var ErrRingFull = NewProtocolError("to icmp buf len to min")
var ErrRingWrite = NewProtocolError("to icmp buf len to min")
var ErrRingRead = NewProtocolError("to icmp buf len to min")
