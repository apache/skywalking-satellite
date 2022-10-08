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

FROM golang:1.18

ADD . /satellite_code
WORKDIR /satellite_code

RUN set -ex; \
    apt-get update; \
    apt-get install -y netcat; \
    make linux; \
    mkdir -p /skywalking-satellite/configs; \
    cp configs/satellite_config.yaml /skywalking-satellite/configs/satellite_config.yaml; \
    cp bin/skywalking-satellite-latest-linux-amd64 /skywalking-satellite/satellite;

WORKDIR /skywalking-satellite

ENTRYPOINT ["./satellite", "start"]