## Desciption
**Experimental project - don't use in production environment**  
A simple ngrok like application but on your own infrastructure.  
You can run `server` on own VPS server and then expose local running web application by `agent` via tunnel connection to the `server`

## Example
To run example project you have to at first set hostname in `hosts` file:  
At linux add: `proxy.local  localhost` to the `/etc/hosts` file.

Next step is just run `docker-compose up` command.

And now you can switch to your web browser and check `proxy.local` url. 
That should show **ghost** application which is not expose directly to the host machine but only via tunnel inside docker network.
