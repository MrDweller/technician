# technician
A technician system subscribes for work tasks via a messaging broker. Once the technician decides to take a work, the system notifies a work handler system.  

This system uses an [event handling libary](https://github.com/MrDweller/event-handler), that is able to publish and subscribe for events to rabbitMQ, whilst also having the publishers register the subsciption details as a service in the Arrowhead framework. 

## Requirements

* **golang 1.22**, other versions should work.
* Arrowhead framework

## Setup

Create an `.env` file,

```

ADDRESS=<address>
PORT=<port>

DOMAIN_ADDRESS=<address that will be registered to the service registry>
DOMAIN_PORT=<port that will be registered to the service registry>>

SYSTEM_NAME=<system name>

SERVICE_REGISTRY_ADDRESS=<service registry address>
SERVICE_REGISTRY_PORT=<service registry port>
SERVICE_REGISTRY_IMPLEMENTATION=<service registry implementation>

CERT_FILE_PATH=<path to cert .pem file>
KEY_FILE_PATH=<path to key .pem file>
TRUSTSTORE_FILE_PATH=<path to truststore .pem file>
AUTHENTICATION_INFO=<authentication info>

EVENT_HANDLING_SYSTEM_TYPE=<what type of event handling>
WORK_HANDLER_TYPE=<what type of work handler>
EXTERNAL_ENDPOINT_URL=<url to external system for notifying of event, necessary only if `EVENT_HANDLING_SYSTEM_TYPE="USER_INTERACTIVE_EVENT_HANDLING"`>

EVENT_HANDLER_IMPLEMENTATION=<event handler implementation used>
```

`EVENT_HANDLING_SYSTEM_TYPE` can currently be either `"USER_INTERACTIVE_EVENT_HANDLING"` or `"DIRECT_EVENT_HANDLING"`.
`WORK_HANDLER_TYPE` can currently only be `"EXTERNAL_WORK_HANDLER"`.
`EVENT_HANDLER_IMPLEMENTATION`, the only implementation currently is `"rabbitmq-3.12.12"`.

## Start
Start the technician system by running, 

```
go run . 
```

This will start a command line interface for using the technician system.

## Commands

* `subscribe <event>`, this will subscribe to all <event> services that the technician is authorized to received, defined by Arrowhead.
* `unsubscribe <event>`, this will unsubscribe from all <event> services that the technician is currently subscribing on.
* `exit` unregisters the technician system from Arrowhead and stops the system
