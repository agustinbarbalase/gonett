# gonett

gonett is a Go version for library [mininet](https://github.com/mininet/mininet/tree/master)

## Build

```bash
make build
```

This produces `bin/gonett`.

## Commands

All commands must run as root (or via sudo).

### Build sample topology

Creates h1, h2 hosts and s1 switch, wires them, assigns IPs, and brings links up.

```bash
sudo ./bin/gonett build
```

### List containers

```bash
sudo ./bin/gonett ls
```

### Exec in a container namespace

```bash
sudo ./bin/gonett exec h1 ping -c 3 10.0.0.2
```

### Attach interactive shell

```bash
sudo ./bin/gonett attach h1
```

### Remove a container by name

```bash
sudo ./bin/gonett rm h1
```

### Cleanup everything

Deletes all managed namespaces, bridges, veths.

```bash
sudo ./bin/gonett cleanup
```

## Notes

- Requires Linux with network namespace support.
- Uses `vishvananda/netlink` and `netns`; namespace operations are thread-bound, so internal code pins goroutines to OS threads where necessary.
