# meshname

Special-use naming system for self-organized IPv6 mesh networks. 
Implements [Meshname protocol](https://github.com/zhoreeq/meshname/blob/master/protocol.md).

Minimum go version 1.12 is required.

1) Get the source code and compile
```
git clone https://github.com/zhoreeq/meshname.git
cd meshname
make
```
2) Generate the default config for your host
```
./meshnamed -genconf 200:6fc8:9220:f400:5cc2:305a:4ac6:967e -subdomain meshname | tee /tmp/meshnamed.conf
```
3) Run the daemon
```
./meshnamed -useconffile /tmp/meshnamed.conf
```
4) Optionally, set configuration flags
```
./meshnamed -listenaddr [::1]:53535 -debug -useconffile /tmp/meshnamed.conf
```
5) See list of all configuration flags
```
./meshnamed -help
```
Add new DNS records to configuration file and restart the daemon to apply settings.
A record can be of any valid string form parsed by [miekg/dns](https://godoc.org/github.com/miekg/dns#NewRR).

## systemd unit

Look for `meshnamed.service` in the source directory for a systemd unit file.

## Example configuration file

In this example, meshnamed is configured as authoritative for two domain zones:

    {
            "Domain":"aiag7sesed2aaxgcgbnevruwpy",
            "Records": [
                    "aiag7sesed2aaxgcgbnevruwpy.meshname. AAAA 200:6fc8:9220:f400:5cc2:305a:4ac6:967e",
                    "_xmpp-client._tcp.aiag7sesed2aaxgcgbnevruwpy.meshname. SRV 5 0 5222 xmpp.aiag7sesed2aaxgcgbnevruwpy.meshname",
                    "_xmpp-server._tcp.aiag7sesed2aaxgcgbnevruwpy.meshname. SRV 5 0 5269 xmpp.aiag7sesed2aaxgcgbnevruwpy.meshname",
                    "xmpp.aiag7sesed2aaxgcgbnevruwpy.meshname. AAAA 300:6fc8:9220:f400::1",
                    "forum.aiag7sesed2aaxgcgbnevruwpy.meshname. CNAME amag7sesed2aaaaaaaaaaaaaau.meshname."
            ]
    }
    {
            "Domain":"amag7sesed2aaaaaaaaaaaaaau",
            "Records":[
                    "amag7sesed2aaaaaaaaaaaaaau.meshname. AAAA 300:6fc8:9220:f400::5"
            ]
    }

## Configure dnsmasq as a primary DNS resolver with "meshname." support

`/etc/dnsmasq.conf`

    port=53
    domain-needed
    bogus-priv
    server=/meshname/::1#53535
    server=8.8.8.8

## Using meshnamed as a standalone DNS server

Set the flag to listen on all interfaces and a standard DNS server port

    ./meshnamed -listenaddr [::]:53 -useconffile /tmp/meshnamed.conf

Run as root and allow incoming connections to port 53/UDP in firewall settings.

## Custom top level domains (TLD) and subnet filtering

meshnamed can be configured to resolve custom TLDs.
To run meshnamed for TLD `.newmesh` with addresses in `fd00::/8` 
set a flag `-networks newmesh=fd00::/8`.

By default, in addition to `.meshname` it also resolves `.ygg` for IPv6 addresses in 
`200::/7` subnet and `.cjd` for `fc00::/8`. 

Requests are filtered by subnet validation. Request is ignored if a decoded 
IPv6 address doesn't match the specified subnet for a TLD.

## See also

[YggNS](https://github.com/russian-meshnet/YggNS/blob/master/README.md)
