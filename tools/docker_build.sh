#!/usr/bin/env bash

#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
DOCKER_DIR=$ROOT_DIR/docker
VERSION=$1
DIST_NAME=skywalking-satellite-$VERSION-bin
DIST_FILE=$ROOT_DIR/$DIST_NAME.tgz
DOCKER_DIST_FILE=$DOCKER_DIR/$DIST_NAME.tgz

if [ ! -f "$DIST_FILE" ]; then
  echo "$DIST_FILE is not exist, could not build the skywalking-satellite docker image."
  exit 1
fi

rm -rf "$DOCKER_DIST_FILE"
cp "$DIST_FILE" "$DOCKER_DIST_FILE"
docker build --build-arg DIST_NAME="$DIST_NAME" -t skywalking-satellite:"$VERSION" --no-cache "$DOCKER_DIR"

if [ $? -eq 0 ]; then
 echo "skywalking-satellite:$VERSION docker images build success!"
else
 echo "skywalking-satellite:$VERSION docker images build failure!"
 exit 1
fi
