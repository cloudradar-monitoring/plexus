# plexus

Plexus exposes a simple interface for remote control with the
[MeshCentral](https://www.meshcommander.com/meshcentral2) Server. It's intented
to be used for ad-hoc remote assistance without installing an agent
permanently.

Plexus is a proxy server that depends on a MeshCentral server.

## Installation

### Prerequisites

#### Create a user for Plexus

Plexus doesn't need root root privileges to function. Therefore, it is
recommended to create a new user account:

```bash
$ sudo useradd -d /var/lib/plexus -m -U -r -s /bin/false plexus
```

#### TLS certificate

Plexus is intended to provide remote assistance to users located outside your
local network. Using encryption is therefore crucial, and you should use a
publicly resolvable FQDN along with certificates trusted by all browser, for
example using Let's encrypt:

```bash
$ FQDN=plexus.example.com
$ certbot certonly -d $FQDN -n --agree-tos --standalone --register-unsafely-without-email
```

Change permissions that the Plexus user can access the keys and certificates:

```bash
$ chgrp -R plexus /etc/letsencrypt
$ chmod -R g=rx /etc/letsencrypt
```

#### Install the MeshCentral server.

**Unless otherwise noted, all commends should be executed with the Plexus user\***:

1. Install the latest Node.js TLS release:

   ```bash
   $ curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
   $ sudo apt-get install -y nodejs
   ```

   [Learn more](https://github.com/nodesource/distributions/blob/master/README.md#deb) about installing Node.js.

1. Change to the Plexus user via bash

   ```bash
   su - plexus -s /bin/bash
   ```

1. Install MeshCentral:

   ```bash
   $ mkdir ~/meshcentral
   $ cd ~/meshcentral
   $ npm install meshcentral
   ```

   It takes a while to download all dependencies.

1. Create a configuration for MeshCentral. MeshCentral will not be exposed to
   the network. It runs on localhost only because all requests go through the
   Plexus proxy.

   ```bash
   $ mkdir ~/meshcentral/meshcentral-data/
   $ cat <<EOF >~/meshcentral/meshcentral-data/config.json
   {
     "\$schema": "http://info.meshcentral.com/downloads/meshcentral-config-schema.json",
     "settings": {
       "LANonly": true,
       "WANonly": false,
       "agentPortTls": false,
       "exactports": true,
       "tlsoffload": true,
       "port": 5000,
       "portBind": "127.0.0.1",
       "redirPort": 5001,
       "redirPortBind": "127.0.0.1",
       "AllowLoginToken": true
     },
     "domains": {
       "control": {
         "title": "Plexus Control",
         "newAccounts": false
       }
     }
   }
   EOF
   ```

1. Configure MeshCental to use your certificates for TLS:

   ```bash
   ln -s /etc/letsencrypt/live/${FQDN}/fullchain.pem ~/meshcentral/meshcentral-data/webserver-cert-public.crt
   ln -s /etc/letsencrypt/live/${FQDN}/privkey.pem ~/meshcentral/meshcentral-data/webserver-cert-private.key
   ```

1. Finally, start MeshCentral using:

   ```bash
   $ cd ~/meshcentral
   $ node node_modules/meshcentral
   ```

   You should get an output like the following on your terminal.

   ```bash
   MeshCentral HTTP redirection server running on port 5001.
   Generating certificates, may take a few minutes...
   Generating root certificate...
   Generating HTTPS certificate...
   Generating MeshAgent certificate...
   Generating Intel AMT MPS certificate...
   MeshCentral v0.9.25, LAN mode.
   Server has no users, next new account will be site administrator.
   MeshCentral HTTP server running on port 5000.
   ```

   If the above command doesn't report errors, stop mehscentral with CTRL-C.

1. Create a user on the MeshCentral server, that Plexus will use for internal connections.

   ```bash
   $ PASSWD=$(openssl rand -hex 30)
   $ node node_modules/meshcentral --createaccount plexus --pass "${PASSWD}" --domain control
   ```

   MeshCentral should respond with:

   ```
   Done.
   ```

1. Now, let's run MeshCentral via Systemd. From the root account, create a
   service file.

   ```bash
   $ cat << EOF > /etc/systemd/system/meshcentral.service
   [Unit]
   Description=MeshCentral Server
   [Service]
   Type=simple
   LimitNOFILE=1000000
   ExecStart=/usr/bin/node /var/lib/plexus/meshcentral/node_modules/meshcentral
   WorkingDirectory=/var/lib/plexus/meshcentral
   Environment=NODE_ENV=production
   User=plexus
   Group=plexus
   Restart=always
   RestartSec=10
   [Install]
   WantedBy=multi-user.target
   EOF
   ```

1. Start and enable the MeshCentral service:

   ```bash
   $ systemctl enable --now meshcentral
   $ systemctl status meshcentral
   ```

   You should get a confirmation like

   ```
   ● meshcentral.service - MeshCentral Server
        Loaded: loaded (/etc/systemd/system/meshcentral.service; disabled; vendor preset: enabled)
        Active: active (running) since Tue 2021-09-14 14:21:45 UTC; 29ms ago
      Main PID: 7329 (node)
         Tasks: 6 (limit: 2354)
        Memory: 2.2M
           CPU: 10ms
        CGroup: /system.slice/meshcentral.service
                └─7329 /usr/bin/node /var/lib/meshcentral/node_modules/meshcentral --tlsoffload --exactports
   ```

Resources

- [MeshCentral Ctrl Documentation](https://info.meshcentral.com/downloads/MeshCentral2/MeshCtrlUsersGuide-0.0.1.pdf)
- [MeshCentral Install Guide](https://info.meshcentral.com/downloads/MeshCentral2/MeshCentral2InstallGuide-0.1.0.pdf)
- [MeshCentral User Guide](https://info.meshcentral.com/downloads/MeshCentral2/MeshCentral2UserGuide-0.2.9.pdf)

### Install Plexus

1. Create a log dir for Plexus

   ```bash
   $ mkdir /var/log/plexus/
   $ chown root:plexus /var/log/plexus/
   $ chmod g+rwx /var/log/plexus/
   ```

1. Install Plexus

   ```bash
   $ cd /tmp
   $ wget https://github.com/cloudradar-monitoring/plexus/releases/download/v0.0.4/plexus_0.0.4_linux_amd64.tar.gz
   $ tar -xzf plexus_0.0.4_linux_amd64.tar.gz -C /usr/local/bin plexus
   ```

1. Create a Plexus configuration file.

   ```bash
   $ mkdir /etc/plexus
   $ cat << EOF > /etc/plexus/plexus.conf
   # The TLS cert file
   PLEXUS_TLS_CERT_FILE=/etc/letsencrypt/live/${FQDN}/fullchain.pem
   # The TLS key file
   PLEXUS_TLS_KEY_FILE=/etc/letsencrypt/live/${FQDN}/privkey.pem
   # The URL of the MeshCentral server.
   PLEXUS_MESH_CENTRAL_URL=ws://localhost:5000
   # The MeshCentral account username.
   PLEXUS_MESH_CENTRAL_USER=plexus
   # The MeshCentral account password.
   PLEXUS_MESH_CENTRAL_PASS=${PASSWD}
   # The address plexus will listen on.
   PLEXUS_SERVER_ADDRESS=0.0.0.0:8080
   # The loglevel (one of: debug, info, warn, error)
   PLEXUS_LOG_LEVEL=info
   PLEXUS_LOG_FILE=/var/log/plexus/plexus.log
   EOF
   ```

1. Check if everything is configured correctly

   ```bash
   $ /usr/local/bin/plexus -c /etc/plexus/plexus.conf verify-config
   ```

   The command should output:

   ```
   Config: Ok!
   MeshCentral Server: Ok!
   TLS Certificate: Ok!
   ```

1. Start Plexus via Systemd

   ```bash
   $ cat << EOF > /etc/systemd/system/plexus.service
   [Unit]
   Description=Plexus Server
   [Service]
   Type=simple
   LimitNOFILE=1000000
   ExecStart=/usr/local/bin/plexus -c /etc/plexus/plexus.conf serve
   WorkingDirectory=/var/lib/plexus
   User=plexus
   Group=plexus
   Restart=always
   AmbientCapabilities=cap_net_bind_service
   RestartSec=10
   [Install]
   WantedBy=multi-user.target
   EOF
   ```

   ```bash
   $ systemctl enable --now plexus
   ```

1. Check that Plexus is running:

   ```bash
   $ systemctl status plexus
   ```

   It should output the following:

   ```
      ● plexus.service - Plexus Server
        Loaded: loaded (/etc/systemd/system/plexus.service; disabled; vendor preset: enabled)
        Active: active (running) since Thu 2021-09-16 17:41:43 UTC; 5s ago
      Main PID: 14307 (plexus)
         Tasks: 7 (limit: 2354)
        Memory: 4.6M
           CPU: 16ms
        CGroup: /system.slice/plexus.service
                └─14307 /usr/local/bin/plexus -c /etc/plexus/plexus.conf serve
   ```

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
   $ go run ./cmd/plexus -c plexus.config.development serve
   ```
