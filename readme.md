# CrestronTcpBridge

## Motivation

2 years ago I bought a house where a crestron system is installed, which consists of a CP2E, two Cresnet modules and a TSW 751 display. I don't like the concept how Crestron designs the user interface and the slight possibilities Crestron offers to combine the equipment with other bridges and interfaces. Therefore I need a possibility to control my shutters, lights and plugs via another system. The Crestron controller offers a TCP server control, which I use to send commands from any client to the controller.  

## Design

#### Crestron

On the Crestron controller a side a TCP server is added. An incomming request on `$RX` is send to the `telnet-server` module. The module validates the access code and, in case of success, extracts the command ID. A `string-case-digital` module reacts on the change of the command ID and set the specified digital output to high. 

-----------------------------                  --------------------------- 
| Client                    |  TCP connection  | CP2E                    |
| [access code][command ID] | ---------------> | valdidate [access code] |
|                           |                  | act [command ID]        |
-----------------------------                  ---------------------------

In addition the `system-state-to-json` module is used to create a response on `$TX`. 

#### Client

The client consists of a service `crebrid` and a program `crebri`. A config file located in `/etc/crebrid/crebrid.cfg` defines where the crestron server is located, on which it will listen and what the access code looks like. The service is connected to the controller and checks the connection frequently. Command could be send via the `crebri` program. The program transmit the command to the service and service finally sends the command to the controller. As a response the program receive the information if the command was successfully send and the current state of the controlled item (e.g. plug off, lights on or shutter up). 
