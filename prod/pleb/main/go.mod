module github.com/pleb/prod/pleb/main

go 1.21.4

require (
	bazil.org/fuse v0.0.0-20230120002735-62a210ff1fd5
	github.com/caarlos0/env/v10 v10.0.0
	github.com/pleb/prod/common/bootstrap v0.0.0-00010101000000-000000000000
	github.com/pleb/prod/common/config v0.0.0-00010101000000-000000000000
	github.com/pleb/prod/senator/pb v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.60.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.16.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/pleb/prod/senator/pb => ../../senator/pb

replace github.com/pleb/prod/common/config => ../../common/config

replace github.com/pleb/prod/common/bootstrap => ../../common/bootstrap
