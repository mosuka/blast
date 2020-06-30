# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [v0.9.1]

- Update tests #139
- Update protocol buffers #135
- Update zap #134
- Update gRPC #133
- Update raft #132
- Update tests #131
- Upgrade Bleve to v1.0.9 #130
- Add test #129

## [v0.9.0]

- Implement CORS #128
- Delete the experimentally implemented feature for distributed search #127
- Add coverage to Makefile #114
- Docker compose #119
- Bump Bleve version to v0.8.1 #117


## [v0.8.1]

- Update go version and dependencies #109


## [v0.8.0]

- Add swagger specification experimentaly #107
- New CLI #82
- Split protobuf into components #84
- Change subcommands #85
- Update protobuf #86
- Change protobuf #87
- Change the cluster watching method #90
- Change cluster watch command for manager #92
- Change node state to enum from string #93
- Change node info structure #94
- Change protobuf for indexer and dispatcher #95
- Change server arguments #96
- Change index protobuf #97
- Use protobuf document #98
- Change node state to Node_SHUTDOWN in a error #99
- Fix a bug for waiting to receive an indexer cluster updates from the stream #100
- Migrate to grpc-gateway #105


## [v0.7.1] - 2019-07-18

- Add raft-badger #69
- Add raft-storage-type flag #73
- Add gRPC access logger #74
- Improve indexing performance #71
- Remove original document #72
- Rename config package to builtins #75


## [v0.7.0] - 2019-07-03

- Add GEO search example #65
- Migrate grpc-middleware #68


## [v0.6.1] - 2019-06-21

- Fix HTTP response into JSON format #64
- Update Dockerfile #62


## [v0.6.0] - 2019-06-19

- Add federated search #30
- Add cluster manager (#48)
- Add KVS HTTP handlers #46
- Update http logger #51
- Update logutils (#50)
- Remve KVS (#49)


## [v0.5.0] - 2019-03-22

- Support bulk update #41
- Support Badger #38
- Add index stats #37
- Add Wikipedia example #35
- Support cznicb and leveldb #34
- Add logging #33
- Add CHANGES.md #29
- Add error handling for server startup #28.
- Fixed some badger bugs #40
- Restructure store package #36
- Update examples #32
- update Makefile #31


## [v0.4.0] - 2019-03-14

- Code refactoring.
