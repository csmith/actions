# Actions, in go

This is a monorepo containing actions for use in GitHub or compatible
workflows, all written in Go and distributed as containers.

Shipping binaries reduces dependencies (no node!), and shipping the
actions as containers means they can bundle recent versions of any
necessary tools, instead of relying on a base image that has 20+GB
of out-of-date tools. 

## Disclaimer

These are a work-in-progress, undocumented, and currently reference
a private docker image. They may be useful as a reference but aren't
usable by anyone but me as-is.
