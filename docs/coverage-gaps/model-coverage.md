# Model Coverage

- OpenAPI models: `30`
- Matched client models: `17`
- Missing client models: `13`
- Extra client models: `35`

## Matched Models

| OpenAPI Model | Client Model | Status | OpenAPI Properties | Covered | Missing | Extra |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| `AwsExternalId` | `AWSExternalID` | covered | 2 | 2 | 0 | 0 |
| `Contact` | `Contact` | covered | 3 | 3 | 0 | 0 |
| `Device` | `Device` | property gaps | 42 | 40 | 2 | 3 |
| `DevicePostureAttributes` | `DevicePostureAttributes` | covered | 2 | 2 | 0 | 0 |
| `DeviceRoutes` | `DeviceRoutes` | covered | 2 | 2 | 0 | 0 |
| `DnsConfiguration` | `DNSConfiguration` | covered | 7 | 7 | 0 | 0 |
| `DnsConfigurationPreferences` | `DNSConfigurationPreferences` | covered | 2 | 2 | 0 | 0 |
| `DnsConfigurationResolver` | `DNSConfigurationResolver` | covered | 2 | 2 | 0 | 0 |
| `DnsPreferences` | `DNSPreferences` | covered | 1 | 1 | 0 | 0 |
| `Key` | `Key` | covered | 21 | 21 | 0 | 0 |
| `KeyCapabilities` | `KeyCapabilities` | covered | 4 | 4 | 0 | 0 |
| `LogstreamEndpointConfiguration` | `SetLogstreamConfigurationRequest` | property gaps (heuristic) | 19 | 18 | 1 | 0 |
| `NetworkFlowLog` | `NetworkFlowLog` | covered | 32 | 32 | 0 | 0 |
| `PostureIntegration` | `PostureIntegration` | property gaps | 12 | 5 | 7 | 0 |
| `TailnetSettings` | `TailnetSettings` | covered | 11 | 11 | 0 | 0 |
| `User` | `User` | covered | 12 | 12 | 0 | 0 |
| `Webhook` | `Webhook` | covered | 8 | 8 | 0 | 0 |

## Missing OpenAPI Models

| OpenAPI Model | Properties |
| --- | ---: |
| `ConfigurationAuditLog` | 20 |
| `ConnectionCounts` | 7 |
| `DeviceInvite` | 14 |
| `DnsSearchPaths` | 1 |
| `Error` | 1 |
| `LogstreamEndpointPublishingStatus` | 12 |
| `ServiceHostInfo` | 3 |
| `SplitDns` | 0 |
| `UserInvite` | 7 |
| `VIPServiceApproval` | 2 |
| `VIPServiceInfo` | 5 |
| `VIPServiceInfoPut` | 5 |
| `subscriptions` | 0 |

## Extra Client Models

| Client Model | Source | Properties |
| --- | --- | ---: |
| `ACL` | `policyfile.go:59` | 60 |
| `ACLAttrConfig` | `policyfile.go:178` | 3 |
| `ACLAutoApprovers` | `policyfile.go:94` | 2 |
| `ACLDERPMap` | `policyfile.go:119` | 14 |
| `ACLDERPNode` | `policyfile.go:132` | 9 |
| `ACLDERPRegion` | `policyfile.go:124` | 13 |
| `ACLEntry` | `policyfile.go:99` | 7 |
| `ACLSSH` | `policyfile.go:144` | 7 |
| `ACLTest` | `policyfile.go:110` | 6 |
| `APIError` | `client.go:71` | 4 |
| `APIErrorData` | `client.go:78` | 2 |
| `ClientConnectivity` | `devices.go:61` | 11 |
| `ClientSupports` | `devices.go:52` | 6 |
| `Contacts` | `contacts.go:26` | 9 |
| `CreateFederatedIdentityRequest` | `keys.go:62` | 7 |
| `CreateKeyRequest` | `keys.go:30` | 6 |
| `CreateOAuthClientRequest` | `keys.go:37` | 3 |
| `CreatePostureIntegrationRequest` | `device_posture.go:40` | 5 |
| `CreateWebhookRequest` | `webhooks.go:75` | 3 |
| `DERPRegion` | `devices.go:47` | 2 |
| `DeviceKey` | `devices.go:349` | 1 |
| `DevicePostureAttributeRequest` | `devices.go:121` | 3 |
| `DevicePostureIdentity` | `devices.go:76` | 3 |
| `Distro` | `devices.go:70` | 3 |
| `Grant` | `policyfile.go:168` | 6 |
| `LogstreamConfiguration` | `logging.go:49` | 17 |
| `NodeAttrGrant` | `policyfile.go:154` | 6 |
| `NodeAttrGrantApp` | `policyfile.go:162` | 3 |
| `Service` | `services.go:17` | 6 |
| `SetFederatedIdentityRequest` | `keys.go:78` | 7 |
| `SetOAuthClientRequest` | `keys.go:50` | 3 |
| `TrafficStats` | `logging.go:179` | 7 |
| `UpdateContactRequest` | `contacts.go:42` | 1 |
| `UpdatePostureIntegrationRequest` | `device_posture.go:49` | 4 |
| `UpdateTailnetSettingsRequest` | `tailnet_settings.go:37` | 11 |
