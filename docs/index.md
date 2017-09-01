---
title: "Serial Vault"
table_of_contents: False
---

# About Serial Vault

The Serial Vault is a web service that generates signed serial assertions for devices. 
It can be run in a data centre or on premises on a factory LAN.

The Serial Vault holds a list of approved models for a manufacturer, and the encrypted 
signing key(s) for the models. The service validates the model and logs if the serial number 
and device-key fingerprint have been previously used.

Security around the SerialVault will be concerned about avoiding key or whole-system 
duplication, but it is left to the ODM/OEM to ensure physical security of the machine (theft) 
and ensuring that request to its Rest API only originate from within a trusted LAN or VPN.

The hardware specification should be industrial grade for durability, rack mountable and high 
spec’d to reduce lag in response time as much as reasonably possible.

It stores model information and the encrypted signing keys in a database, and can optionally 
use a TPM module as a part of the encryption process.
