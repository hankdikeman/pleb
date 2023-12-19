https://docs.aws.amazon.com/sdk-for-go/api/service/s3/


API:

Generic bucket puts/gets. Shouldn't be used

    - GetContent
    - PutContent

Tools (e.g. model weights)
    
    - PutTools (Server)
    - GetTools (Client)

Inputs (e.g. training data)

    - PutInputs (Server)
    - GetInputs (Client)

Outputs (e.g. training misses, new weights)

    - PutOutputs (Client)
    - GetOutputs (Server)


Streaming gRPC APIs: https://jbrandhorst.com/post/grpc-binary-blob-stream/
Will need to be compressed on the wire as well
