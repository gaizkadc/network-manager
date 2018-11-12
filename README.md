# Network-Manager
The Network-Manager is the component in charge of:
* costumer and application network management, which includes network creation, deletion, retrieval, and listing.
* member authorization to join these networks.
* DNS entry central management, which includes DNS entry creation, deletion, and listing.

The Network-Manager includes the following sub-components:
* Network-Manager Server
* Networking-Client
* DNS-Client

## Network-Manager Server
To launch the Network-Manager run:

`$ ./bin/network-manager run --ztaccesstoken <ztaccesstoken> --consoleLogging --debug`

## Networking-Client
Keep in mind that the System-Model must be running to execute these commands.

#### Add network:
`$ ./bin/networking-cli add --name <networkName> --orgid <organizationID> --consoleLogging --debug`

#### Delete network:
`$ ./bin/networking-cli delete --netid <networkID> --orgid <organizationID> --consoleLogging --debug`

#### Get network:
`$ ./bin/networking-cli get --netid <networkID> --orgid <organizationID> --consoleLogging --debug`

#### List networks:
`$ ./bin/networking-cli list --orgid <organizationID> --consoleLogging --debug`

#### Authorize member:
`$ ./bin/networking-cli authorize --orgid <organizationID> --netid <networkID> --memberid <memberID> --consoleLogging --debug`

## DNS-Client

Again, System-Model must be running to execute these commands.

#### Add entry:
`$ ./bin/dns-cli add --fqdn <FQDN> --ip <IP> --netid <networkID>  --consoleLogging --debug`

#### Delete entry:
`$ ./bin/dns-cli delete --fqdn <FQDN> --orgid <organizationID> --consoleLogging --debug`

#### List entries:
`$ ./bin/dns-cli list --orgid <organizationID> --consoleLogging --debug`

More options are available on all commands. Run `-h` or `--help` at any point in the command to see all available options.