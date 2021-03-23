# go-home
### [This project is now on hold for architectural restructuring]
go-home is a home IoT server that allows devices to connect to hub and dynamically add their services to be able to be controlled from the hub

The go-home uses sqlite as persistent storage  

Device can connect to `/connect` and will be given an `id` to be saved. Next time this device can connect with said `id` to refresh the connection.  

Said device will provide a [.serv](https://github.com/IktaS/go-serv) service definition as the basis for calling their endpoint from the hub.  

Connection to `/connect` will be in the form of json:  
  - `name` as the device name
  - `hub-code` for authentication to the hub
  - `serv` for service definition
  - `algo` for the algo used to decompress `serv`
  
List of devices that's available will be able to be accessed in `/device`  

`/device` will follow a rest-like form.

You can access each device with `/device/[id]`, and their respective service and message from `/device/[id]/service` and `/device/[id]/message`.

And you can call a device service by hitting `/device/[id]/service/[service-name]?[service-params]` with `service-params` follows a URL query like input.

An example of an IoT device implementing this can be seen in [this esp32 example](https://github.com/IktaS/esp32-go-home-module-example)

If you're interested in developing or just have any question in general, feel free to open a discussion in this repo, or contact me on discord Ikta#8871
