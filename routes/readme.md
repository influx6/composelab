#Composelab/Routes

#Example
Below is a very simple example of using these package to build interesting routes

    ```
      package routes

      import (
      	"testing"

      	. "github.com/franela/goblin"
      	"github.com/influx6/goutils"
      	"github.com/influx6/grids"
      )

    	app := NewRoutes("app") //=> creates a /app route

      //you can create routes with regular expressions using
      //the [Reggy] package route patterns
      // app/{tag:[regexp]} like below
      //this creates a /app/:id route
    	app.Branch(`{id:[\d+]}`)

      //creates a /app/log/realtime route
    	app.Branch("logs/realtime")


      //lets listen for the /app/log route
    	logs, err := app.Select("logs")
    	logs.Terminal().Only(func(req interface{}) {
    		_, ok := req.(*grids.GridPacket)
    		gob.Assert(ok).IsTrue()
    		done()
    	})

      //lets do the long-way for sending request
    	pack := grids.NewPacket()
    	pack.Set("Pathways", []string{"app", "logs"})
    	app.InSend("Request", pack)


      //lets listen for the /app/:id route
    	id, err := app.Select("id")
    	id.Terminal().Any(grids.ByPackets(func(f *grids.GridPacket){
        //lets get the param map
    		params, ok := f.Get("Params").(*goutils.Map)
        //do something
    	}))

      //lets do the easy way for doing requests
    	app.IssueRequestPath("app/4", "400")

    ```


##Secret API
Underneath the wrapped and nice looking API (God's grace of course) of this package, it actually uses the [Grids](http://github.com/influx6/grids) as the underline system for all route build up. Each *Routes composes the Grids struct and thereby provides the power of that package. Also *Routes uses the [Reggy] ClassicMatcher for each route parts,making it flexible to define routes as a combination of regular expressions

##Requests
In Composable/Routes, requests are just [Grids] GridPackets that contain a speific Meta data "Pathways". This contains the list of individual pathways to which these request must pass. The choice for this approach was for the reason for flexiblity and compatibility with the [Grids] API. Also with this any packet is a valid request which lends itself to a lot of possiblities. As a packet succesfully passes through each *Routes, each packet is given a meta map "Params" where the current match for the current route is added as a key:value pair

#API
The API for Composelab/Routes is very simple and easy.

####Interfaces
- TerminalType
This interface define the member functions for a *Routes terminal type

  - **Only(func(data interface{}))**
  This member function adds a listen to the Routes event system and will only fire the given callback when the request matches just the *Route owner only. ie. if a Route has a /app path unless the request is for /app path,then this function will not be called

  - **Any(func(data interface{}))**
  This member function adds a listen to the Routes event system and will only fire the given callback when the request matches the *Route owner wether strictly or unstrictly ie. if a Route has a /app path if the request falls under its path(such as a request with path /app/logs), as far as the request matches the *Route path partly because it flows through it then these callback will be called

- RouteTypes
This interface defines the basic method which make the API for routes and what they exactly do

  - **Select(string)**
  These function retrieves the*Routes object if it exits for the set of route path provided

  - **Divert(string)**
  These function defines the behaviour of sublinking from the current route,it creates a simple branch off these route

  - **Branch(string)**
  These function defines the behaviour of alternative linking i.e it allows sequenctial route behaviour where we naturally go down a path way based on priority of who gets the request first.When the first receivers fails to match the request it passes it to the next link *Routes and so on until the request is fullfilled or revoked

  - **Terminal() Terminal**
  These function returns a terminal struct that has two event binders `Only` and `Any`, these two exists to provide simple bindings for when in the case of `Any` you desire to listen to all valid route matching these part or for in the case of 'Only' for only route visitors specific for just these route

####Structs

- **Terminal**
Terminal represents a event struct for a route which contains an event handler for both the Any and Only event subscription and publishing

  - **NewTerm() *Terminal**
    This returns a new terminal for use by a route

      ```
          term := NewTerm();
      ```

  - **Terminal.Any(callback func(data interface{}))**
    This member function binds a callback to the terminal any event handler

      ```
          term.Any(func(event interface{}){
            //do something
          })

          term.AllEvents.Emit(200)

      ```

  - **Terminal.Only(callback func(data interface{}))**
    This member function binds a callback into the All event handler

      ```
          term.Only(func(event interface{}){
            //do something
          })

          term.OnlyEvents.Emit(200)
      ```

- **Routes**
This struct represent a standard route piece that receives request packet then validates and forwards it to routes along its paths. Underneath as said earlier in #SecretAPI ,routes underneath compose [Grids] which gives them all the power that the [Grids] API provides

  - **NewRoutes(string) *Routes**
    This function returns a new Route struct pointer ready for use

      ```
      	app := NewRoutes("app") //=> creates a /app route
      ```

  - **Routes.IssueRequest(grids.GridPacket) error**
    This takes a packet and sends it off as a request if it has the required 'Pathways' meta tag in the packet headers

      ```
        //grids.NewPacket is from the Grids package
        pack := grids.NewPacket()
        pack.Set("Pathways",[]string{"apps","logs"})

      	app := NewRoutes("app") //=> creates a /app route

        err := app.IssueRequest("app/logs",pack)

        if err != nil {
          //do something
        }

      ```

  - **Routes.IssueRequestPath(string,interface{}) grids.GridPacket**
    This takes a path, a data of any type to which it sets as the "Payload" meta on the packet and sends it off as a request if it has the required 'Pathways' meta tag in the packet headers

      ```

      	app := NewRoutes("app") //=> creates a /app route

        app.IssueRequestPath("app/logs","4000")

      ```

  - **Routes.IssueRequestPacket(string,grids.GridPacket)**
    This takes a path and a packet and sends it off as a request

      ```
        //grids.NewPacket is from the Grids package
        pack := grids.NewPacket()

      	app := NewRoutes("app") //=> creates a /app route

        app.IssueRequestPacket("app/logs",pack)
      ```

  - **Routes.Select(string) (*Route,error)**
    As defined above in the RouteType interface

      ```
      	app := NewRoutes("app") //=> creates a /app route
      ```

  - **Routes.Branch(string)**
    As defined above in the RouteType interface

      ```
        //creates a app/:id route where id is must be an number
      	app.Branch(`{id:[\d+]}`)

        //create a /app/log/realtime route
      	app.Branch("logs/realtime")
      ```

  - **Routes.Divert(*Routes) *Routes**
    As defined above in the RouteType interface

      ```
      	app := NewRoutes("app") //=> creates a /app route
        sub := NewRoutes("subs") //-> creates a /subs route

        _ = app.Divert(sub) //returns sub for changing effect
      ```

  - **Routes.Terminal(*Routes) *Terminal**
    As defined above in the RouteType interface

      ```
      	app := NewRoutes("app") //=> creates a /app route

        app.Terminal().Only(func(data interface{}){ ... })
        app.Terminal().Any(func(data interface{}){ ... })
      ```


[Grids]: http://github.com/influx6/grids
[Reggy]: http://github.com/influx6/reggy
