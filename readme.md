#Composelab
Composelab is the combines the idea of microservices with the real-time distributed web apps where which is a single independent service that still walk together for the creation of a full flegde service

##Install

    go get github.com/influx6/composelab

Then
    
    go install github.com/influx6/composelab

##Architecture
Composelab architecture is based on the idea that there are majorly two types of services a master directory which tells other services which services are available within the network which means we can also localize some services to be within a specific region(real-life geographical region or virtual) or within a closed loop of services. Composelab tries as much as possible not to redefined the standard web architecture but to use it to its advantage likes restful uris and api’s principles and the idea that services can and should tell others about what they provide and what is the conditions of their serving of a request such like https(http + ssl) only or needing an encryprion key. Services are identified by their “id” string.
    
    WebService Example:
     Our example service network has a set of services within its scope which includes:
        
        - Model Manager (/models)
        - Attachment Service (/attachments)
        - View Service (/views)


        Client  can generally directly request response from any of these services but each services can also forward its data to another services as an input to be processed and returned as a response to the initial request


##API
