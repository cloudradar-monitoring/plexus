# plexus

Plexus exposes a simple interface for remote control with the
[Meshcentral](https://www.meshcommander.com/meshcentral2) Server. It's intented
to be used for ad-hoc remote assistence without installing an agent
permanently.

## Installation

tbd

## Usage

To create a remote control session, follow these steps:

1. Create a new session:

   ```bash
   $ curl https://localhost:8080/session \
          -F id="helping-joe" \
          -F ttl=3600 \
          -F username=admin \
          -F password=foobaz
   ```

   - username / password is optional and will be asked when opening session
     inside the browser or deleting the session.

   - ttl is the time to live of the session in seconds

   Plexus will respond with the following:

   ```json
   {
     "ID": "helping-joe",
     "SessionURL": "https://localhost:8080/session/helping-joe",
     "AgentMSH": "https://localhost:8080/config/helping-joe:xddRGfuIOaqwBxVIWrnp",
     "AgentConfig": {
       "ServerID": "9DA17836FD0BA3193ED52896929FD021990EBA4234ED56A6057115B7C53D24F58284E34954619CAECC131FC8BE82B9EE",
       "MeshName": "plexus/helping-joe/bGVnI",
       "MeshType": 2,
       "MeshID": "mesh//i8bgwSqhUVS5ttAYX5VCqSR2dxPY@5iSLv5p1jFJG5TJKYV91PaMoTf0AHSj@Eqi",
       "MeshIDHex": "0x8BC6E0C12AA15154B9B6D0185F9542A924767713D8FB98922EFE69D631491B94C929857DD4F68CA137F40074A3F84AA2",
       "MeshServer": "wss://localhost:8080/agent/helping-joe:xddRGfuIOaqwBxVIWrnp"
     },
     "ExpiresAt": "2021-09-12T14:31:40.830231373+02:00"
   }
   ```

1. Create the `meshagent.msh` on the system that will be remote controlled:

   ```bash
   $ curl https://localhost:8080/config/helping-joe:xddRGfuIOaqwBxVIWrnp > meshagent.msh
   ```

   The url is present in the response of the create session request. See `.AgentMSH`.

1. Start the [MeshAgent](https://github.com/Ylianst/MeshAgent) in the directory
   where you've created the `meshagent.msh`. You can get the binary from
   [here](https://github.com/Ylianst/MeshCentral/tree/master/agents).

   ```bash
   $ ./meshagent
   ```

1. Open the `.SessionURL` from the create session response in your favorite browser.

1. Click 'connect' on the upper left and the remote control should've started.

## Development

1. Start the development MeshCentral server:

   ```bash
   $ (cd dev && npm install)
   $ (cd dev && npm start)
   ```

1. Start Plexus with the development config:

   ```bash
   $ go run ./cmd/plexus -c plexus.config.development
   ```
