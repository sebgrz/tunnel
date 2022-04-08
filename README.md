## Desciption
**Experimental project - don't use in production environment**  
A simple ngrok like application but on your own infrastructure.  
You can run `server` on own VPS server and then expose local running web application by `agent` via tunnel connection to the `server`

## Example
To run example project you have to at first set hostnames in `hosts` file:  
On linux add:   
```
proxy.local  localhost
```
to the `/etc/hosts` file.

Next step is just run `docker-compose up` command.

And now you can switch to your web browser and check `proxy.local` url. 
That should show **ghost** application which is not expose directly to the host machine but only via tunnel inside docker network.

If you have any tool to test websocket connection you can connect to the internal testing websocket server 
via the same hostname: `proxy.local`.

## SSL
Run `generate-self-signed-ssl.sh` script inside `/ssl` directory in order to generate self-signed certificate for `proxy.local` domain.  
Script base on this article [How to create an HTTPS certificate for localhost domains](https://gist.github.com/cecilemuller/9492b848eb8fe46d462abeb26656c4f8)  
These certificats are use by docker compose setup with nginx proxy on top, so the script should be run before `docker-compose up`. All of these give you ability to run example project over https/wss protocols.

> Important! You have to add RootCA.crt to your web browser certificate authority in order to not encounter NOT_VALID_CERTIFICATE error.
