# meshname

Special-use naming system for self-organized IPv6 mesh networks. 

## Motivation

Having a naming system is a common requirement for deploying preexisting 
decentralized applications. Protocols like e-mail, XMPP and ActivityPub require 
domain names for server to server communications.

Self-organized networks like CJDNS and Yggdrasil Network use public-key 
cryptography for IP address allocation. Every network node owns 
a globally unique IPv6 address. Binary form of that address can be encoded with 
base32 notation for deriving a globally unique name space managed by that node.

Since there is no need for a global authority or consensus, such a naming system 
will reliably work in any network split scenarios.

".meshname" is meant to be used by machines, not by humans. A human-readable 
naming system would require a lot more engineering effort. 

## How .meshname domains work

Each mesh node can manage its own unique name space in "meshname." zone. 
The name space is derived from its IPv6 address as follows:

1) IPv6 address is converted to its binary form of 16 bytes:

    IPv6Address('200:6fc8:9220:f400:5cc2:305a:4ac6:967e')

    b'\x02\x00o\xc8\x92 \xf4\x00\\\xc20ZJ\xc6\x96~'

2) The binary value is encoded to base32:

    AIAG7SESED2AAXGCGBNEVRUWPY======

3) Padding symbols "======" are removed from the end of the string.

The resulting name space managed by '200:6fc8:9220:f400:5cc2:305a:4ac6:967e'
is "aiag7sesed2aaxgcgbnevruwpy.meshname."

In order to resolve a domain in "xxx.meshname." space, the client derives IPv6 
address from the second level domain "xxx" and use it as authoritative DNS server
for that zone.

"xxx.meshname" name is itself managed by the DNS server derived from "xxx" and 
can point to any other IPv6 address.

## Resolving process explained

1) A client application makes a request to a resolver.
I.e. request AAAA record for "test.aiag7sesed2aaxgcgbnevruwpy.meshname.".

2) When a resolver detects "meshname." domain, it extracts the second level 
domain from it. In this example, "aiag7sesed2aaxgcgbnevruwpy.meshname.".

3) If the resolver is configured as an authoritative server for that 
domain, it sends back a response as a regular DNS server would do.

4) If it's not, the resolver derives IPv6 address of the corresponding 
authoritative DNS server from the second level domain.
For "aiag7sesed2aaxgcgbnevruwpy.meshname." the authoritative server is 
"200:6fc8:9220:f400:5cc2:305a:4ac6:967e".
The resolver then relays clients request to a derived server address and 
relays a response back to the client.

## Why not .ip6.arpa

There is a special domain for reverse DNS lookups, but it takes 72 characters to
store a single value. The same value in .meshname takes 35 characters.

"e.7.6.9.6.c.a.4.a.5.0.3.2.c.c.5.0.0.4.f.0.2.2.9.8.c.f.6.0.0.2.0.ip6.arpa" 
versus "aiag7sesed2aaxgcgbnevruwpy.meshname"

This saves twice amount of bandwidth and storage space. It is also arguably more 
aesthetically appealing, even though that's not a goal.
