# Endpoint Coverage

| Tag | OpenAPI | Covered | Missing |
| --- | ---: | ---: | ---: |
| Contacts | 3 | 2 | 1 |
| DNS | 11 | 11 | 0 |
| DeviceInvites | 6 | 0 | 6 |
| DevicePosture | 5 | 5 | 0 |
| Devices | 15 | 13 | 2 |
| Keys | 5 | 4 | 1 |
| Logging | 8 | 5 | 3 |
| PolicyFile | 4 | 3 | 1 |
| Services | 7 | 0 | 7 |
| TailnetSettings | 2 | 2 | 0 |
| UserInvites | 5 | 0 | 5 |
| Users | 7 | 1 | 6 |
| Webhooks | 7 | 7 | 0 |

## Missing From The Client

| Method | Path | Operation ID | Tags | Summary |
| --- | --- | --- | --- | --- |
| POST | `/tailnet/{tailnet}/contacts/{contactType}/resend-verification-email` | `resendContactVerificationEmail` | Contacts | Resend verification email |
| POST | `/device-invites/-/accept` | `acceptDeviceInvite` | DeviceInvites | Accept a device invite |
| DELETE | `/device-invites/{deviceInviteId}` | `deleteDeviceInvite` | DeviceInvites | Delete a device invite |
| GET | `/device-invites/{deviceInviteId}` | `getDeviceInvite` | DeviceInvites | Get a device invite |
| POST | `/device-invites/{deviceInviteId}/resend` | `resendDeviceInvite` | DeviceInvites | Resend a device invite |
| GET | `/device/{deviceId}/device-invites` | `listDeviceInvites` | DeviceInvites | List device invites |
| POST | `/device/{deviceId}/device-invites` | `createDeviceInvites` | DeviceInvites | Create device invites |
| POST | `/device/{deviceId}/expire` | `expireDeviceKey` | Devices | Expire a device's key |
| PATCH | `/tailnet/{tailnet}/device-attributes` | `batchUpdateCustomDevicePostureAttributes` | Devices | Batch update custom device posture attributes |
| GET | `/tailnet/{tailnet}/keys` | `listTailnetKeys` | Keys | List tailnet keys |
| GET | `/tailnet/{tailnet}/logging/configuration` | `listConfigurationAuditLogs` | Logging | List configuration audit logs |
| GET | `/tailnet/{tailnet}/logging/network` | `listNetworkFlowLogs` | Logging | List network flow logs |
| GET | `/tailnet/{tailnet}/logging/{logType}/stream/status` | `getLogStreamingStatus` | Logging | Get log streaming status |
| POST | `/tailnet/{tailnet}/acl/preview` | `previewRuleMatches` | PolicyFile | Preview rule matches |
| GET | `/tailnet/{tailnet}/services` | `listServices` | Services | List all Services |
| DELETE | `/tailnet/{tailnet}/services/{serviceName}` | `deleteService` | Services | Delete a Service |
| GET | `/tailnet/{tailnet}/services/{serviceName}` | `getService` | Services | Get a Service |
| PUT | `/tailnet/{tailnet}/services/{serviceName}` | `updateService` | Services | Update a Service |
| GET | `/tailnet/{tailnet}/services/{serviceName}/device/{deviceId}/approved` | `getServiceDeviceApproval` | Services | Get approval status of Service on a device |
| POST | `/tailnet/{tailnet}/services/{serviceName}/device/{deviceId}/approved` | `updateServiceDeviceApproval` | Services | Update approval status of Service on a device |
| GET | `/tailnet/{tailnet}/services/{serviceName}/devices` | `listServiceHosts` | Services | List devices hosting a Service |
| GET | `/tailnet/{tailnet}/user-invites` | `listUserInvites` | UserInvites | List user invites |
| POST | `/tailnet/{tailnet}/user-invites` | `createUserInvites` | UserInvites | Create user invites |
| DELETE | `/user-invites/{userInviteId}` | `deleteUserInvite` | UserInvites | Delete a user invite |
| GET | `/user-invites/{userInviteId}` | `getUserInvite` | UserInvites | Get a user invite |
| POST | `/user-invites/{userInviteId}/resend` | `resendUserInvite` | UserInvites | Resend a user invite |
| GET | `/tailnet/{tailnet}/users` | `listUsers` | Users | List users |
| POST | `/users/{userId}/approve` | `approveUser` | Users | Approve a user |
| POST | `/users/{userId}/delete` | `deleteUser` | Users | Delete a user |
| POST | `/users/{userId}/restore` | `restoreUser` | Users | Restore a user |
| POST | `/users/{userId}/role` | `updateUserRole` | Users | Update user role |
| POST | `/users/{userId}/suspend` | `suspendUser` | Users | Suspend a user |

## Implemented But Missing From OpenAPI

| Method | Path | Client Method | Source |
| --- | --- | --- | --- |
| GET | `/api/v2/tailnet/{tailnet}/vip-services` | `ServicesResource.List` | `services.go:32` |
| DELETE | `/api/v2/tailnet/{tailnet}/vip-services/{}` | `ServicesResource.Delete` | `services.go:66` |
| GET | `/api/v2/tailnet/{tailnet}/vip-services/{}` | `ServicesResource.Get` | `services.go:46` |
| PUT | `/api/v2/tailnet/{tailnet}/vip-services/{}` | `ServicesResource.CreateOrUpdate` | `services.go:56` |
