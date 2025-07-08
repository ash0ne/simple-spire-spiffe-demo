# A Simple Spiffe Demo with Spire
This is a  simple demo project designed to help understand SPIFFE concepts through a clear, simple example using SPIRE.

## SPIFFE Overview

The core idea behind [SPIFFE](https://spiffe.io/) (Secure Production Identity Framework For Everyone) is to provide cloud and platform-agnostic workload identities.

Instead of relying on IPs, hostnames, or manual secrets, SPIFFE issues cryptographically verifiable identities to workloads — no matter where they run. This makes it really easy to manage inter process authnetication in a complex estate that has multiple services running in diffferent hosting setups.

## The Concept
The architecture is based on the idea that most modern software architectures have a core platform/domain as the trust boundary, that in-turn has nodes/hosts on top which host workloads/services. 

So, within each domain, there is a SPIFFE Server that acts as the trusted identity authority and each host in that domain runs a SPIFFE Agent that communicates with the SPIFFE Server.

Workloads running on those hosts get their own SPIFFE IDs — these identities are issued by the agent, verified by the server, and scoped under the host’s trust domain.

![alt text](SPIFFE.svg)

### Example

Imagine a system where:

**Trust Domain:** example.org — is your organization or platform boundary.
**Host:** is a server running a SPIFFE Agent, attested and registered with the SPIFFE Server.
**Workloads:** are individual applications or services running on the host, each assigned a SPIFFE ID like:

spiffe://example.org/agent1
spiffe://example.org/workloadA
spiffe://example.org/workloadB

Here, the workloads’ identities are parented by the host’s SPIFFE ID, allowing secure, auditable identity delegation.


![alt text](SPIFFE-Flow.svg)

----------------------------------------------------------------------------------------------------------------------

## How To Run This Project (SPIFFE in action)

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
8. Now create a workload with the agent as the parent (replacing the join token)

    ```
    docker exec -it simple-spire-spiffe-demo-spire-server-1  bin/spire-server entry create \
    -parentID spiffe://example.org/spire/agent/join_token/<join_token> \
    -spiffeID spiffe://example.org/workload1 \
    -selector unix:uid:0
    ```
9. Now to fetch the SVID, run the below command

    ```
    docker exec -it simple-spire-spiffe-demo-spire-agent-1 bin/spire-agent api fetch x509 -output json
    ```
