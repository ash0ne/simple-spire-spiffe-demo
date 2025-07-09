# A Simple Spiffe Demo with Spire

This is a simple demo project designed to help understand SPIFFE concepts through a clear, simple example using SPIRE.

## SPIFFE Overview

The core idea behind [SPIFFE](https://spiffe.io/) (Secure Production Identity Framework For Everyone) is to provide cloud and platform-agnostic workload identities.

Instead of relying on IPs, hostnames, or manual secrets, SPIFFE issues cryptographically verifiable identities to workloads — no matter where they run. This makes it really easy to manage inter process authentication in a complex estate that has multiple services running in different hosting setups.

## The Concept

The architecture is based on the idea that most modern software architectures have a core platform/domain as the trust boundary, that in-turn has nodes/hosts which have workloads and services hosted.

So, within each domain, there is a SPIFFE Server that acts as the trusted identity authority and each host in that domain runs a SPIFFE Agent that communicates with the SPIFFE Server.

Workloads running on those hosts get their own SPIFFE IDs — these identities are issued by the agent, verified by the server, and scoped under the host’s trust domain.

![alt text](SPIFFE.svg)

### Example

Imagine a system where:

- **Trust Domain:** example.org — is your organisation or platform boundary.
- **Host:** is a server running a SPIFFE Agent, attested and registered with the SPIFFE Server.
- **Workloads:** are individual applications or services running on the host, each assigned a SPIFFE ID like:

  - spiffe://example.org/agent1 - For host
  - spiffe://example.org/server - For Workload A
  - spiffe://example.org/client - For another Workload B that needs to call A

Here, the workloads’ identities are parented by the host’s(agent's) SPIFFE ID, allowing secure, auditable identity delegation.

![alt text](SPIFFE-Flow.svg)

⚠️ **Note:** Although this project is built to fully run on Docker, it uses unix uid based SPIFFE workload registrations. This is intentional to make it compatible with many operating systems and because this is intended to only run as a quick demo. Ideally when using SPIFFE in containerised environments you'd want to use Docker WorkloadAttestor instead of Unix based workload attestation - And maybe also go with JWT based SVIDs instead of cert based SVIDs.

---

## How To Run This Project (SPIFFE in action)

Make sure to run all the below commands from the root of the project directory.

1. Let's spin up the postgres server that is needed as the datastore for the SPIFFE server.

   ```
   docker compose up postgres -d
   ```

2. Spin up the SPIFFE server. SPIRE is a production-ready implementation of the SPIFFE.

   ```
   docker compose up spire-server -d
   ```

3. Extract the pem from server.

   ```
   docker exec -it simple-spire-spiffe-demo-spire-server-1 /opt/spire/bin/spire-server bundle show -format pem > ./spire/agent/bundle.pem
   ```

4. Generate a join token to attach agent to the server

   ```
   docker exec -it simple-spire-spiffe-demo-spire-server-1 /opt/spire/bin/spire-server token generate -spiffeID spiffe://example.org/agent1
   ```

5. Take the UUID token returned by the above command and replace it with the `<join_token>` in `docker-compose.yml`

6. Start the agent

   ```
   docker compose up spire-agent -d
   ```

7. Run the below command and you must see the agent's ID registered.

   ```
   docker exec -it simple-spire-spiffe-demo-spire-server-1 bin/spire-server entry show
   ```

8. Now create a workload-server with the agent as the parent (replacing the join token)

   ```
   docker exec -it simple-spire-spiffe-demo-spire-server-1  bin/spire-server entry create \
   -parentID spiffe://example.org/spire/agent/join_token/<join_token> \
   -spiffeID spiffe://example.org/server \
   -selector unix:uid:0
   ```

9. Now just run a quick test to see if you are able to fetch the SVID from within the agent container using the below command

   ```
   docker exec -it simple-spire-spiffe-demo-spire-agent-1 bin/spire-agent api fetch x509 -output json
   ```

10. Now create a workload-client with the agent as the parent (replacing the join token)

    ```
    docker exec -it simple-spire-spiffe-demo-spire-server-1  bin/spire-server entry create \
    -parentID spiffe://example.org/spire/agent/join_token/<join_token> \
    -spiffeID spiffe://example.org/client \
    -selector unix:uid:1000
    ```

11. Start the workload-server by running

    ```
    docker compose up workload-server -d
    ```
    if this is successful, you'll see a log like so

    ```
    Creating X509Source (SPIFFE Workload API client)...
    Configuring TLS with mTLS...
    Starting HTTPS server on :8443...
    ```
12. Finally, start the workload-client by running

    ```
    docker compose up workload-client -d
    ```
    The client makes an mTLS call to the server and immediately exits, but you'll know it has worked and you'll see a log like below

    Server log:
    ```
    [Server] VerifyPeerCertificate called with 1 certs
    [Server] Received request
    [Server] TLS handshake complete. Peer certificates count: 1
    [Server] Client SPIFFE ID from cert: spiffe://example.org/client
    [Server] Response written successfully
    ```

    Client log:
    ```
    [Client] Creating X509Source (SPIFFE Workload API client)...
    [Client] Configuring TLS with mTLS...
    [Client] Sending GET request to https://workload-server:8443
    [Client] VerifyPeerCertificate called with 1 certs
    [Client] GetClientCertificate called
    [Client] Returning client certificate
    [Client] Reading response body...
    [Client] Server response: Hello, client with SPIFFE ID: spiffe://example.org/client
    [Client] Closing X509Source...
    ```
