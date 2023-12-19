Layout:
    - PLEBS
        - Local worker node, belongs to a Pat
        - Probably local to the same process as its Pat
    - PATS
        - Local administrator for compute nodes with members
    - SENS
        - I don't know if there is an advantage to "Edge"-ish nodes or not
    - CAESAR
        - Administers Pats (Senators?) and manages service discovery

Pieces to assemble independently:

1. C -> S: assign jobs, no data xfer
2. S -> PA: upload data, retrieve
3. PA -> PL: split work


Service Discovery - FORA

- Start with a centralized KV-store, eventually need distributed and scalable
    (move towards ZK eventually?)

Persistent Storage

- NoSQL storage for small stuff - FABRICAE
    - S3 object IDs, client IDs, job IDs

- S3 object storage for big stuff - HORREA
    - S3 object storage, encoded datasets + models

- NoSQL storage for small data, S3 object storage for big stuff (datasets)
    S3 contains datasets, model weights (finalized)
    NoSQL stores S3 object IDs + other crap about job


Order of operations:

Microservice architecture
    PLEB        -   Client
    PATRICIAN   -   Edge
    SENATOR     -   Server? Edge?
    CAESAR      -   Server
    FORA        -   Server
    FABRICAE    -   AWS, Server FE
    HORREA      -   AWS, Server FE

1. dockerized processes (empty) for each of the above
2. service discovery - FORA
3. S3 frontend - HORREA
4. NoSQL frontend - FABRICAE
5. Job management - CAESAR
6. Job performance - PLEB
7. Middlemen? - PATRICIAN + SENATOR
