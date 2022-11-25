# Welcome to Testworld

`"CHAOS TAKES CONTROL. WELCOME TO THE NEW WORLD. WELCOME TO TESTWORLD"`

`Testworld` (loosely analogous to [Westworld](https://medium.com/@naveen101/westworld-an-introduction-cc7d29bfbe84) ;) ) is a simulation and test environment for centrifuge p2p network. 
Here you can create, run and test PODs with various behaviours to observe how they would behave and debug any problems encountered.

## Description

### Primary entities

#### Controller
The `Controller` holds the webhook `WebhookReceiver` used to receive job and document notifications, and keeps track of all the hosts
used in the tests. It exposes functions that are required during test runs to retrieve the hosts, create clients for them or create
other accounts.

#### Host
The `Host` represents one POD instance, that has its own set of services and storage.
Each host has one keyring pair which is used to create its main account/identity during bootstrap. Note that this main account
is an anonymous proxy.

#### Control unit
The `ControlUnit` holds the config and service context used for each host, and it's the main entrypoint for interacting
with any running services.

### Secondary entities

#### Client
The `Client` is used to interact with each host via the HTTP API.

#### Webhook Receiver

The `WebhookReceiver` is an HTTP server that receives all job and document notifications sent by a POD instance.