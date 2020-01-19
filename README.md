# meshname

Special-use naming system for self-organized IPv6 mesh networks. 

## Motivation

Having a naming system is a common requirement for deploying preexisting 
decentralized applications. Protocols like e-mail, XMPP and ActivityPub require 
domain names for server to server communications.

Self-organized and trustless networks like CJDNS and Yggdrasil Network are 
using public-key cryptography for IP address allocation. Every network node owns 
a globally unique IPv6 address, and 16 bytes of that address can be 
translated to a globally unique domain name.

Since there is no need for a global authority or consensus, such a naming system 
will reliably work in any network split scenarios.

".mesh.arpa" is ment to be used by machines, not by humans. A human-readable 
naming system would require a lot more engineering effort. 

## How to resolve .mesh.arpa domains 

Every third level domain in ".mesh.arpa" space represents a single IPv6 address.

Domain "aicrxoqgun7siwm42akzfsox7m.mesh.arpa" is resolved as follows:

1) Append base32 padding "======" to the upper cased third level domain token;

    AICRXOQGUN7SIWM42AKZFSOX7M======

2) Decode base32 string to a binary IPv6 representation;

    b'\x02\x05\x1b\xba\x06\xa3\x7f$Y\x9c\xd0\x15\x92\xc9\xd7\xfb'

3) Convert the resulting 16 bytes to a IPv6 address structure.

    IPv6Address('205:1bba:6a3:7f24:599c:d015:92c9:d7fb')

If the server cannot translate a given domain name to IP address it should 
return empty response. 

Every additional subdomain, e.g. "mail.xxx.mesh.arpa, xmpp.xxx.mesh.arpa" 
resolves to the same IPv6 address as "xxx.mesh.arpa".

## Why not .ip6.arpa

There is a special domain for reverse DNS lookups, but it takes 72 characters to
store a single value. The same value in .mesh.arpa takes 36 characters.

"7.c.4.9.0.d.8.f.8.d.2.a.6.4.6.7.8.e.2.d.4.b.1.a.d.4.7.8.0.0.2.0.ip6.arpa" 
versus "aicrxoqgun7siwm42akzfsox7m.mesh.arpa"

This saves twice amount of bandwidth and storage space. It is also arguably more 
aesthetically appealing, even though that's not a goal.

## Why .arpa

".arpa" is a special domain reserved for Internet infrastructure. There is a 
similar special-use domain for home networks ".home.arpa" specified in RFC 8375. 
If ".mesh.arpa" will become widely used it could also be standardized, otherwise 
it won't break much.
