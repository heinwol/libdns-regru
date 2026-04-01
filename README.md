DEVELOPER INSTRUCTIONS:
=======================

This repo is a template for developers to use when creating new [libdns](https://github.com/libdns/libdns) provider implementations.

Be sure to update:

- The package name
- The Go module name in go.mod
- The latest `libdns/libdns` version in go.mod
- All comments and documentation, including README below and godocs
- License (must be compatible with Apache/MIT)
- All "TODO:"s is in the code
- All methods that currently do nothing

**Please be sure to conform to the semantics described at the [libdns godoc](https://github.com/libdns/libdns).**

_Remove this section from the readme before publishing._

---

\<PROVIDER NAME\> for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/TODO:PROVIDER_NAME)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for \<PROVIDER\>, allowing you to manage DNS records.

TODO: Show how to configure and use. Explain any caveats.

- reg.ru uses TTL configuration **per zone**, not per record.
  - The mechanism chosen for handling AppendRecords/SetRecords TTL is the following:
    1. SOA data is cached.
    2. Whenever an `Append/Set` request is to be sent, all the input records are checked for any ttl modifications.
       - If ttl across all records is the same, the changes are propagated to the provider as a separate `UpdateSOA` request (happening after records modifications).
       - If ttl fields differ, the **minimum** of them is selected and written into the input records. Modified versions are returned from the function (if changes were successful, of course).
    3. SOA.MinimumTTL is **never** modified, you should manually call `UpdateSOA` whenever you want to change it.
- reg.ru does not support transactional changes. If a libdns request fails in the middle, **no cleanup is performed** (maybe we'll deal with it later).
