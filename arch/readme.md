#Arch
Arch is a part of composelab which defines the basic low-level achitecture of all services written with composelab and that is you are either a service or a servicelink.ServiceLinks define a remote connection which links one service to another, the good part of this style of architecture is we dont care which is the underline structure, method or protocol used to speak to that remote service but it has to meet the criteria that we can delegate a request to it and recieve a response while Services are the core structures which entails the behaviour of the service itself and its operations.

##Architecture
Within this package are defined to main struct `Service` and `ServiceLink` based on two defined interfaces `Services` and `Linkage`

##API

###Interfaces:
Interfaces defined within the [Arch] package for the composelab microservice architecture

-   **Services**
Services are the core interface and they define the set of functions needed by any struct to be defined or see as servable in the context of a service in composelab. These functions include:

  - **Dial() void**
  These function initiates the remote connection and sets of the service into motion

  - **Drop() void**
  These function drops the connection and stops all communication with the respective service

  - **Location() string**
   These returns the location format of a service which is generally in the format <ServiceName>@<address:port>

-   **Linkage**
This defines the functions required by all connectors for services. They provide the basic communication links and are used by services for its own operations and also as communication with master service

  - **GetPath() string**
  These return the format <address>:<port> detailing the address and port of the link is connected in

  - **GetAddress() string**
  These returns the specific address of this service link  

  - **GetPort() int**
  These returns the specific port of these service link

  - **Serve() void**
  These function is used to request a specific action from the servicelink to its service

  - **Dial() void**
  These function initializes the ServiceLink connection for request processing

  - **Drop() void**
  These function ends the ServiceLink connection

###Structs:
Structs defined within the [Arch] package for the composelab microservice architecture

-   **Service**
These struct defines specific attributes for the service built using these struct

-   **ServiceLink**
This struct defines specific attributes for services that meets the Linkage interface



[Arch]: http://github.com/influx6/composelab/arch
