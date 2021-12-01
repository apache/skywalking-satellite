# Enable/Disable Channel

Different channels mean that different protocols can be transparently transmitted to upstream services(OAP).

## Config

In the Satellite configuration, a channel is represented under the configured `pipes`. By default, we open all channels and process all known protocols.

You could **delete** the channel if you don't want to receive and transmit in satellite.

After restart the satellite service, then the channel what you delete is disable.