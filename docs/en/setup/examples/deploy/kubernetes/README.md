# Deploy on Kubernetes

It could help you run the Satellite as a gateway in Kubernetes environment.

## Install

We recommend install the Satellite by `helm`, follow command below, it could start the latest release version of SkyWalking Backend, UI and Satellite.

```shell
export SKYWALKING_RELEASE_NAME=skywalking  # change the release name according to your scenario
export SKYWALKING_RELEASE_NAMESPACE=default  # change the namespace to where you want to install SkyWalking
export REPO=skywalking

helm repo add ${REPO} https://apache.jfrog.io/artifactory/skywalking-helm                                
helm install "${SKYWALKING_RELEASE_NAME}" ${REPO}/skywalking -n "${SKYWALKING_RELEASE_NAMESPACE}" \
  --set oap.image.tag=8.8.1 \
  --set oap.storageType=elasticsearch \
  --set ui.image.tag=8.8.1 \
  --set elasticsearch.imageTag=6.8.6 \
  --set satellite.enabled=true \
  --set satellite.image.tag=v0.4.0
```

## Change Address

After the Satellite and Backend started, need to change the address from agent/node. Then the satellite could load balance the request from agent/node to OAP backend.

Such as in Java Agent, you should change the property value in `collector.backend_service` forward to this: `skywalking-satellite.${SKYWALKING_RELEASE_NAMESPACE}:11800`.
