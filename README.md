```bash
sysctl -w net.ipv6.ip_nonlocal_bind = 1
ip r add local fd80::/64 dev lo
```
