# How to install and use

Minimum go version 1.12 is required.

1) Get the source code and compile
```
git clone https://github.com/zhoreeq/meshname.git
cd meshname
make
```
2) Run the daemon
```
./meshnamed
```
3) Optionally, set configuration flags
```
./meshnamed -listenaddr [::1]:53535 -debug
```
4) See the list of all available flags
```
./meshnamed -help
```

## Get meshname subdomain from an IPv6 address

```
./meshnamed -getname 200:f8b1:f974:967f:dd32:145d:1cc0:3679
aiaprmpzoslh7xjscrorzqbwpe
```

Use this subdomain with a .meshname TLD to configure DNS records 
on your authoritative server, (i.e. dnsmasq, bind or PopuraDNS).

## systemd unit

Look for `meshnamed.service` in the source directory for a systemd unit file.

## Configure dnsmasq as a primary DNS resolver with "meshname." support

`/etc/dnsmasq.conf`

    port=53
    domain-needed
    bogus-priv
    server=/meshname/::1#53535
    server=8.8.8.8

## Custom top level domains (TLD) and subnet filtering

meshnamed can be configured to resolve custom TLDs.
To run meshnamed for TLD `.newmesh` with addresses in `fd00::/8` 
set a flag `-networks newmesh=fd00::/8`.

By default, in addition to `.meshname` it also resolves `.ygg` for IPv6 addresses in 
`200::/7` subnet and `.cjd` for `fc00::/8`. 

Requests are filtered by subnet validation. Request is ignored if a decoded 
IPv6 address doesn't match the specified subnet for a TLD.
