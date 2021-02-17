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

FROM bash

ARG DIST_NAME='skywalking-satellite-latest-bin'

COPY "$DIST_NAME.tgz" /

RUN set -ex; \
    tar -xzf "$DIST_NAME.tgz"; \
    rm -rf "$DIST_NAME.tgz"; \
    mkdir /skywalking-satellite; \
    mv "$DIST_NAME"/* /skywalking-satellite;

WORKDIR /skywalking-satellite

EXPOSE 12800 11800 1234

ENTRYPOINT ["/bin/sh","bin/startup.sh"]
