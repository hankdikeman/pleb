This is a project to allow a local filesystem mount to read/write over the network and access a centralized data repository. The client API will be serviceable by several different underlying storage frameworks which are opaque to the client.

Microservice architecture
    PLEB        -   Client process, running on the host system. Hooks VFS -> server application
    SENATOR     -   Low level client management, serves client reqs directly
    CAESAR      -   High level client management, e.g. authentication
    FORA        -   Distributed KV Store, e.g. wrapper around ZK
    IUDEX       -   Concurrency Manager
    FABRICAE    -   Storage Frontend, FileAttrs, etc.
    HORREA      -   Raw Storage Frontend

development roadmap

1. empty docker containers for above (except Pleb)
2. Horrea bulk storage frontend
3. Senator, client management APIs
4. Pleb framework, basic FUSE interface
5. Caesar, client connections
6. Pleb FUSE framework, finish interface
7. TBD
