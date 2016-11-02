# OnionWatch - nionyn

[![Build Status](https://drone.io/github.com/BrassHornCommunications/OnionWatch/status.png)](https://drone.io/github.com/BrassHornCommunications/OnionWatch/latest)

## How it works

Every hour we scan the Onionoo database for relays that users are monitoring. If a relay or Hidden Service goes offline we send an *(optionally GPG encrypted)* email to inform you.

This is a rudimentary replacement for the Tor Weather service.


### Subscribing to Alerts
* Visit https://onionwatch.email/manage/ and search for your relay / hidden service
* Click *Subscribe to alerts*
* Enter your email address *(and GPG key if you want to receive GPG encrypted email)*
* Check your email and follow the link

### Unsubscribing
* Visit https://onionwatch.email/unsubscribe/
* Enter your email address and the relay / hidden service alert you want to unsubscribe from
* Check your email and follow the link
* Click Unsubscribe *(or unsubscribe from all)*


## Improvements
* Support Twitter DMs for notification
* Report when Relay changes port / IP
* Report if anything else changes (bandwidth etc)
* SMS / Phonecall support (Twilio?)
* Distributed consensus / use DirPorts directly
* Active checks *(connect to ORPort / DirPort etc)*
* Move to ns1/ns2/ns3 brasshorncomms.uk and use DNSSEC, TLSA etc
* Offer tshirts ;)
