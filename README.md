
# Network-Manager
The Network-Manager is the component in charge of:
* costumer and application network management, which includes network creation, deletion, retrieval, and listing.
* create and manage connections between applications
* member authorization to join these networks.
* DNS entry central management, which includes DNS entry creation, deletion, and listing.

The Network-Manager includes the following sub-components:
* Network-Manager Server
* Networking-Client
* DNS-Client

## Getting Started

To launch the Network-Manager Server run:

`$ ./bin/network-manager run --ztaccesstoken <ztaccesstoken> --consoleLogging --debug`
### Prerequisites

Detail any component that has to be installed to run this component.

* Component1 (e.g.: system-model)
* Component2 (e.g.: conductor)

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```

## User client interface

**Networking-Client**

Keep in mind that the System-Model must be running to execute these commands.

-  Add network:

`$ ./bin/networking-cli add --name <networkName> --orgid <organizationID> --consoleLogging --debug`

- Delete network:

`$ ./bin/networking-cli delete --netid <networkID> --orgid <organizationID> --consoleLogging --debug`

- Get network:

`$ ./bin/networking-cli get --netid <networkID> --orgid <organizationID> --consoleLogging --debug`

- List networks:

`$ ./bin/networking-cli list --orgid <organizationID> --consoleLogging --debug`

- Authorize member:

`$ ./bin/networking-cli authorize --orgid <organizationID> --netid <networkID> --memberid <memberID> --consoleLogging --debug`

**DNS-Client**

Again, System-Model must be running to execute these commands.

- Add entry:

`$ ./bin/dns-cli add --fqdn <FQDN> --ip <IP> --netid <networkID>  --consoleLogging --debug`

- Delete entry:

`$ ./bin/dns-cli delete --fqdn <FQDN> --orgid <organizationID> --consoleLogging --debug`

- List entries:

`$ ./bin/dns-cli list --orgid <organizationID> --consoleLogging --debug`

More options are available on all commands. Run `-h` or `--help` at any point in the command to see all available options.

Ignore this entry if it does not apply.

## Known Issues

Explain any relevant issues that may affect this repo.


## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/nalej/network-manager/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/network-manager/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.



