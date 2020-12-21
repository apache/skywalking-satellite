#!/usr/bin/env bash

# ----------------------------------------------------------------------------
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------



export PROTO_HOME=protocol/all_protocol
export COLLECT_PROTOCOL_HOME=protocol/skywalking-data-collect-protocol
export SATELLITE_PROTOCOL_HOME=protocol/satellite-protocol
export GEN_CODE_PATH=protocol/gen-codes

export COLLECT_PROTOCOL_MODULE=skywalking/network

go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

# generate codes by merged proto files
rm -rf $GEN_CODE_PATH && rm -rf $PROTO_HOME
mkdir ${PROTO_HOME} && mkdir $GEN_CODE_PATH
cp -R ${COLLECT_PROTOCOL_HOME}/* ${PROTO_HOME} && cp -R ${SATELLITE_PROTOCOL_HOME}/* ${PROTO_HOME}
rm -rf ${PROTO_HOME}/*/*Compat.proto && rm -rf ${PROTO_HOME}/*/*compat.proto
protoc -I=${PROTO_HOME} --go_out=${GEN_CODE_PATH} --go-grpc_out=${GEN_CODE_PATH} ${PROTO_HOME}/*/*.proto
rm -rf ${PROTO_HOME}

# init  go modules
cd ${GEN_CODE_PATH}/${COLLECT_PROTOCOL_MODULE}||exit 1
go mod init ${COLLECT_PROTOCOL_MODULE}
go mod tidy
cd -|| exit 1







