package errors

var EthernetLenError = NewProtocolError("to ethernet buf len to min")
var IpLenError = NewProtocolError("to ip buf len to min")
var IcmpLenError = NewProtocolError("to icmp buf len to min")
