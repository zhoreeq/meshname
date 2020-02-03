# meshname

Minimum go version 1.12 is required.

1) Get the source code and compile
```
git clone https://github.com/zhoreeq/meshname.git
cd meshname
make
```
2) Generate the default config for your host
```
./meshnamed genconf 200:6fc8:9220:f400:5cc2:305a:4ac6:967e | tee /tmp/meshnamed.conf
```
3) Optionally, set the configuration with environment variables
```
export LISTEN_ADDR=[::1]:53535
export MESH_SUBNET=200::/7
```
4) Run the daemon
```
./meshnamed daemon /tmp/meshnamed.conf
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

Set environment varialbe to listen on all interfaces and a standard DNS server port

    export LISTEN_ADDR=[::]:53

Allow incoming connections to port 53/UDP in firewall settings.

