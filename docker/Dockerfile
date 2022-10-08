# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.18 as build

ARG VERSION="latest"

ENV CGO_ENABLED=0

WORKDIR /src

COPY . .

RUN VERSION=$VERSION make linux
RUN mv /src/bin/skywalking-satellite-${VERSION}-linux-amd64 /src/bin/skywalking-satellite

FROM scratch

VOLUME /skywalking/configs

WORKDIR /skywalking

COPY --from=build /src/bin/skywalking-satellite /
COPY --from=build /src/configs /skywalking/configs

EXPOSE 12800 11800 1234

CMD ["/skywalking-satellite", "start", "--config", "/skywalking/configs/satellite_config.yaml"]
