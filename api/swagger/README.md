# Swagger setup for the relayer node
The relayer node exposes some api routes in order to monitor the health of it:
- connection status between ethereum and elrond clients
- information for each half-bridge regarding the current status

In order to setup the swagger for interacting with those routes we will use `docker`

1. Edit [openapi.yaml](swagger/openapi.yaml), changing the host to the relayer address.
2. (Optionally) edit [docker-compose.yml](docker-compose.yml) ports.
3. Run docker
    
    ```docker-compose up -d```
4. Go to localhost:8082 and start interacting with the swagger. Default value

   (if you changed the port from docker-compose.yml, this has to be reflected also here)