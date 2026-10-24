# Model Property Coverage

- OpenAPI model properties: `259`
- Covered in matched client models: `172`
- Missing from matched or absent client models: `87`

## Missing OpenAPI Models

### `ConfigurationAuditLog`

| Property |
| --- |
| `action` |
| `actionDetails` |
| `actor.displayName` |
| `actor.id` |
| `actor.loginName` |
| `actor.tags` |
| `actor.type` |
| `deferredAt` |
| `error` |
| `eventGroupID` |
| `eventTime` |
| `new` |
| `old` |
| `origin` |
| `target.id` |
| `target.isEphemeral` |
| `target.name` |
| `target.property` |
| `target.type` |
| `type` |

### `ConnectionCounts`

| Property |
| --- |
| `dst` |
| `proto` |
| `rxBytes` |
| `rxPkts` |
| `src` |
| `txBytes` |
| `txPkts` |

### `DeviceInvite`

| Property |
| --- |
| `accepted` |
| `acceptedBy.id` |
| `acceptedBy.loginName` |
| `acceptedBy.profilePicUrl` |
| `allowExitNode` |
| `created` |
| `deviceId` |
| `email` |
| `id` |
| `inviteUrl` |
| `lastEmailSentAt` |
| `multiUse` |
| `sharerId` |
| `tailnetId` |

### `DnsSearchPaths`

| Property |
| --- |
| `searchPaths` |

### `Error`

| Property |
| --- |
| `message` |

### `LogstreamEndpointPublishingStatus`

| Property |
| --- |
| `lastActivity` |
| `lastError` |
| `maxBodySize` |
| `numBytesSent` |
| `numEntriesSent` |
| `numFailedRequests` |
| `numSpoofedEntries` |
| `numTotalRequests` |
| `rateBytesSent` |
| `rateEntriesSent` |
| `rateFailedRequests` |
| `rateTotalRequests` |

### `ServiceHostInfo`

| Property |
| --- |
| `approvalLevel` |
| `configured` |
| `stableNodeID` |

### `SplitDns`

No leaf properties were extracted for this schema.

### `UserInvite`

| Property |
| --- |
| `email` |
| `id` |
| `inviteUrl` |
| `inviterId` |
| `lastEmailSentAt` |
| `role` |
| `tailnetId` |

### `VIPServiceApproval`

| Property |
| --- |
| `approved` |
| `autoApproved` |

### `VIPServiceInfo`

| Property |
| --- |
| `addrs` |
| `comment` |
| `name` |
| `ports` |
| `tags` |

### `VIPServiceInfoPut`

| Property |
| --- |
| `addrs` |
| `comment` |
| `name` |
| `ports` |
| `tags` |

### `subscriptions`

No leaf properties were extracted for this schema.

## Matched Models With Property Gaps

### `Device` -> `Device`

- Source: `devices.go:82`
- Match type: `exact`
- OpenAPI properties: `42`
- Covered: `40`
- Missing: `2`
- Extra: `3`

| Missing Property |
| --- |
| `advertisedRoutes` |
| `multipleConnections` |

| Extra Property | Source |
| --- | --- |
| `AdvertisedRoutes` | `devices.go:110` |
| `clientConnectivity.derp` | `devices.go:63` |
| `postureIdentity.hardwareAddresses` | `devices.go:79` |

### `LogstreamEndpointConfiguration` -> `SetLogstreamConfigurationRequest`

- Source: `logging.go:70`
- Match type: `heuristic`
- OpenAPI properties: `19`
- Covered: `18`
- Missing: `1`
- Extra: `0`

| Missing Property |
| --- |
| `logType` |

### `PostureIntegration` -> `PostureIntegration`

- Source: `device_posture.go:31`
- Match type: `exact`
- OpenAPI properties: `12`
- Covered: `5`
- Missing: `7`
- Extra: `0`

| Missing Property |
| --- |
| `clientSecret` |
| `configUpdated` |
| `status.error` |
| `status.lastSync` |
| `status.matchedCount` |
| `status.possibleMatchedCount` |
| `status.providerHostCount` |

