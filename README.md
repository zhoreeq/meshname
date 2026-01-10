# meshname

<img src="https://raw.githubusercontent.com/zhoreeq/meshname/master/img/logo-medium.png">

A universal naming system for all IPv6-based mesh networks, including CJDNS and Yggdrasil. 
Implements the [Meshname protocol](https://github.com/zhoreeq/meshname/blob/master/protocol.md).

## F.A.Q.

- Q: *Is it like a decentralized DNS thing?*
- A: Yeah, sort of. With it you can host your own meshname domains and resolve domains of others.

- Q: *Meshname domains are ugly.*
- A: Yes, if you want decentralization, you either have ugly names or a blockchain. Meshname has ugly names, but it works at least!

## How to use meshname domains?

Use a full-featured DNS server with the meshname protocol support, i.e. [PopuraDNS](https://github.com/popura-network/PopuraDNS).

For a standalone .meshname stub resolver see `USAGE.md`

## Alternative implementations

[Mario DNS](https://notabug.org/acetone/mario-dns) by acetone, a C++ implementation with a web interface.

[Ruby gem](https://rubygems.org/gems/meshname) by marek22k, [source](https://codeberg.org/mark22k/meshname/).

## See also

[YggNS](https://github.com/russian-meshnet/YggNS/blob/master/README.md)
