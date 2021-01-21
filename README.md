# ComplexLB

An http layer-7 load balancer written in go.

![alt text](https://user-images.githubusercontent.com/19676877/105277379-50cd5c80-5b71-11eb-89da-9386d8627af6.png)

ComplexLB looks to provide the standard features that other http load balancing software provide,
while taking advantage of Go's powerful concurrent constructs and semantically inscribed software engineering practices. 

It also performs active cleaning and passive recovery for unhealthy backends.
# Rationale

## Why a software load balancer and not a hardware loadbalancer ?

![alt text](https://user-images.githubusercontent.com/19676877/105276269-fc28e200-5b6e-11eb-9476-3eeb3e8cf9a3.png)
Hardware load balancers suffer from many limitations that software load balancers do not.
For e.g :
- Need for private rack-and-stack hardware.
- Complex configuration.
- Redundancy often requires another load balancer per existing load balancer.
- Scalibility is harder and more expensive since demand is variable and to accomodate more lbs need to be purchased in times of peak demand.
And in regular times the extra hardware will stay idle.
And the list goes on.
Meanwhile, software load balancers:
- Can be easily installed on virtual machines.
- Regardless of net traffic they can autoscale in real time, so supply & demand can achieve cost balance easier.
- Idle load balancing servers can be easily repurposed or 'deallocated'

## Why layer 7 ?

Doing load balancing at this layer allows us to handle request at the http level. Which allows us to multiplex our requests
with far more richer and complex data. For example rather than to use an obscure numeric flag on the a low level protocol packet(which is be very fast).
We can multiplex data on the contents of an http request for example header data. 


# Usage
```bash
Usage:
  -servers string
        Load balanced servers, use commas to separate
  -port int
        Port to serve (default 3030)
  -algo string
        Type of scheduling algorithm.
```
Example:

To add followings as load balanced backends
- http://localhost:3031
- http://localhost:3032
- http://localhost:3033
- http://localhost:3034
```bash
complex-lb.exe --servers=http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034 --algo=cycle
```
## Docker Compose

Running docker-compose will create a container running the previous command along side 3 images that will simply act as the adjacent
servers to the load balancer and will return 'Hello world' and the machine id.
```bash
docker-compose up --build
```



# To-do
- Add more scheduling algorithms.
- Interface to multiplex.
