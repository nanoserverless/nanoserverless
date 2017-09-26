# Build the function
I want to be able to build a function with a simple api call  
The /create endpoint take an template and a source code file  
This creates a docker image ready to run my function (in container and service mode)  

# Run the function in container mode
I want that function consume nothing by default (so no online service running)  
When I call that function, it runs a container, execute my code, return result and destroy container (`docker run --rm`)  
Headers and query params are passed in environment variables  

# Up the function
To avoid the creation/destroy overhead time, I want to be able to UP the function in a swarm service  
So in function image, I add https://github.com/msoap/shell2http binary  
This binary listen on TCP, when called, run the function and print to stdout  
Http Headers and Query Params are available in environment variables, and body is available in stdin  

# Run the function in service mode
The api call to the function is the same as container mode  
The gateway detects if a swarm service is available and forward request if so  
If no service UP, the function is run in container mode  

# Not implemented but in my mind
Scale the function - Abitlity to scale function  
Auto scale/up/down function - Load average on function can up/down/scale function  
