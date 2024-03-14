This repository contains a client library for LaunchDarkly's REST API. This client was automatically
generated from our [OpenAPI specification](https://app.launchdarkly.com/api/v2/openapi.json) using a [code generation library](https://github.com/launchdarkly/ld-openapi). View our [sample code](#getting-started) for example usage.

This REST API is for custom integrations, data export, or automating your feature flag workflows. *DO NOT* use this client library to include feature flags in your web or mobile application. To integrate feature flags with your application, read the [SDK documentation](https://docs.launchdarkly.com/sdk).

This client library is only compatible with the latest version of our REST API, version `20220603`. Previous versions of this client library, prior to version 10.0.0, are only compatible with earlier versions of our REST API. When you create an access token, you can set the REST API version associated with the token. By default, API requests you send using the token will use the specified API version. To learn more, read [Versioning](https://apidocs.launchdarkly.com/#section/Overview/Versioning).
# Go API client for ldapi

# Overview

## Authentication

All REST API resources are authenticated with either [personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens), or session cookies. Other authentication mechanisms are not supported. You can manage personal access tokens on your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.

LaunchDarkly also has SDK keys, mobile keys, and client-side IDs that are used by our server-side SDKs, mobile SDKs, and JavaScript-based SDKs, respectively. **These keys cannot be used to access our REST API**. These keys are environment-specific, and can only perform read-only operations such as fetching feature flag settings.

| Auth mechanism                                                                                  | Allowed resources                                                                                     | Use cases                                          |
| ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| [Personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens) | Can be customized on a per-token basis                                                                | Building scripts, custom integrations, data export. |
| SDK keys                                                                                        | Can only access read-only resources specific to server-side SDKs. Restricted to a single environment. | Server-side SDKs                     |
| Mobile keys                                                                                     | Can only access read-only resources specific to mobile SDKs, and only for flags marked available to mobile keys. Restricted to a single environment.           | Mobile SDKs                                        |
| Client-side ID                                                                                  | Can only access read-only resources specific to JavaScript-based client-side SDKs, and only for flags marked available to client-side. Restricted to a single environment.           | Client-side JavaScript                             |

> #### Keep your access tokens and SDK keys private
>
> Access tokens should _never_ be exposed in untrusted contexts. Never put an access token in client-side JavaScript, or embed it in a mobile application. LaunchDarkly has special mobile keys that you can embed in mobile apps. If you accidentally expose an access token or SDK key, you can reset it from your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.
>
> The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.

### Authentication using request header

The preferred way to authenticate with the API is by adding an `Authorization` header containing your access token to your requests. The value of the `Authorization` header must be your access token.

Manage personal access tokens from the [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.

### Authentication using session cookie

For testing purposes, you can make API calls directly from your web browser. If you are logged in to the LaunchDarkly application, the API will use your existing session to authenticate calls.

If you have a [role](https://docs.launchdarkly.com/home/team/built-in-roles) other than Admin, or have a [custom role](https://docs.launchdarkly.com/home/team/custom-roles) defined, you may not have permission to perform some API calls. You will receive a `401` response code in that case.

> ### Modifying the Origin header causes an error
>
> LaunchDarkly validates that the Origin header for any API request authenticated by a session cookie matches the expected Origin header. The expected Origin header is `https://app.launchdarkly.com`.
>
> If the Origin header does not match what's expected, LaunchDarkly returns an error. This error can prevent the LaunchDarkly app from working correctly.
>
> Any browser extension that intentionally changes the Origin header can cause this problem. For example, the `Allow-Control-Allow-Origin: *` Chrome extension changes the Origin header to `http://evil.com` and causes the app to fail.
>
> To prevent this error, do not modify your Origin header.
>
> LaunchDarkly does not require origin matching when authenticating with an access token, so this issue does not affect normal API usage.

## Representations

All resources expect and return JSON response bodies. Error responses also send a JSON body. To learn more about the error format of the API, read [Errors](/#section/Overview/Errors).

In practice this means that you always get a response with a `Content-Type` header set to `application/json`.

In addition, request bodies for `PATCH`, `POST`, and `PUT` requests must be encoded as JSON with a `Content-Type` header set to `application/json`.

### Summary and detailed representations

When you fetch a list of resources, the response includes only the most important attributes of each resource. This is a _summary representation_ of the resource. When you fetch an individual resource, such as a single feature flag, you receive a _detailed representation_ of the resource.

The best way to find a detailed representation is to follow links. Every summary representation includes a link to its detailed representation.

### Expanding responses

Sometimes the detailed representation of a resource does not include all of the attributes of the resource by default. If this is the case, the request method will clearly document this and describe which attributes you can include in an expanded response.

To include the additional attributes, append the `expand` request parameter to your request and add a comma-separated list of the attributes to include. For example, when you append `?expand=members,roles` to the [Get team](/tag/Teams#operation/getTeam) endpoint, the expanded response includes both of these attributes.

### Links and addressability

The best way to navigate the API is by following links. These are attributes in representations that link to other resources. The API always uses the same format for links:

- Links to other resources within the API are encapsulated in a `_links` object
- If the resource has a corresponding link to HTML content on the site, it is stored in a special `_site` link

Each link has two attributes:

- An `href`, which contains the URL
- A `type`, which describes the content type

For example, a feature resource might return the following:

```json
{
  \"_links\": {
    \"parent\": {
      \"href\": \"/api/features\",
      \"type\": \"application/json\"
    },
    \"self\": {
      \"href\": \"/api/features/sort.order\",
      \"type\": \"application/json\"
    }
  },
  \"_site\": {
    \"href\": \"/features/sort.order\",
    \"type\": \"text/html\"
  }
}
```

From this, you can navigate to the parent collection of features by following the `parent` link, or navigate to the site page for the feature by following the `_site` link.

Collections are always represented as a JSON object with an `items` attribute containing an array of representations. Like all other representations, collections have `_links` defined at the top level.

Paginated collections include `first`, `last`, `next`, and `prev` links containing a URL with the respective set of elements in the collection.

## Updates

Resources that accept partial updates use the `PATCH` verb. Most resources support the [JSON patch](/reference#updates-using-json-patch) format. Some resources also support the [JSON merge patch](/reference#updates-using-json-merge-patch) format, and some resources support the [semantic patch](/reference#updates-using-semantic-patch) format, which is a way to specify the modifications to perform as a set of executable instructions. Each resource supports optional [comments](/reference#updates-with-comments) that you can submit with updates. Comments appear in outgoing webhooks, the audit log, and other integrations.

When a resource supports both JSON patch and semantic patch, we document both in the request method. However, the specific request body fields and descriptions included in our documentation only match one type of patch or the other.

### Updates using JSON patch

[JSON patch](https://datatracker.ietf.org/doc/html/rfc6902) is a way to specify the modifications to perform on a resource. JSON patch uses paths and a limited set of operations to describe how to transform the current state of the resource into a new state. JSON patch documents are always arrays, where each element contains an operation, a path to the field to update, and the new value.

For example, in this feature flag representation:

```json
{
    \"name\": \"New recommendations engine\",
    \"key\": \"engine.enable\",
    \"description\": \"This is the description\",
    ...
}
```
You can change the feature flag's description with the following patch document:

```json
[{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"This is the new description\" }]
```

You can specify multiple modifications to perform in a single request. You can also test that certain preconditions are met before applying the patch:

```json
[
  { \"op\": \"test\", \"path\": \"/version\", \"value\": 10 },
  { \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }
]
```

The above patch request tests whether the feature flag's `version` is `10`, and if so, changes the feature flag's description.

Attributes that are not editable, such as a resource's `_links`, have names that start with an underscore.

### Updates using JSON merge patch

[JSON merge patch](https://datatracker.ietf.org/doc/html/rfc7386) is another format for specifying the modifications to perform on a resource. JSON merge patch is less expressive than JSON patch. However, in many cases it is simpler to construct a merge patch document. For example, you can change a feature flag's description with the following merge patch document:

```json
{
  \"description\": \"New flag description\"
}
```

### Updates using semantic patch

Some resources support the semantic patch format. A semantic patch is a way to specify the modifications to perform on a resource as a set of executable instructions.

Semantic patch allows you to be explicit about intent using precise, custom instructions. In many cases, you can define semantic patch instructions independently of the current state of the resource. This can be useful when defining a change that may be applied at a future date.

To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header.

Here's how:

```
Content-Type: application/json; domain-model=launchdarkly.semanticpatch
```

If you call a semantic patch resource without this header, you will receive a `400` response because your semantic patch will be interpreted as a JSON patch.

The body of a semantic patch request takes the following properties:

* `comment` (string): (Optional) A description of the update.
* `environmentKey` (string): (Required for some resources only) The environment key.
* `instructions` (array): (Required) A list of actions the update should perform. Each action in the list must be an object with a `kind` property that indicates the instruction. If the instruction requires parameters, you must include those parameters as additional fields in the object. The documentation for each resource that supports semantic patch includes the available instructions and any additional parameters.

For example:

```json
{
  \"comment\": \"optional comment\",
  \"instructions\": [ {\"kind\": \"turnFlagOn\"} ]
}
```

If any instruction in the patch encounters an error, the endpoint returns an error and will not change the resource. In general, each instruction silently does nothing if the resource is already in the state you request.

### Updates with comments

You can submit optional comments with `PATCH` changes.

To submit a comment along with a JSON patch document, use the following format:

```json
{
  \"comment\": \"This is a comment string\",
  \"patch\": [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }]
}
```

To submit a comment along with a JSON merge patch document, use the following format:

```json
{
  \"comment\": \"This is a comment string\",
  \"merge\": { \"description\": \"New flag description\" }
}
```

To submit a comment along with a semantic patch, use the following format:

```json
{
  \"comment\": \"This is a comment string\",
  \"instructions\": [ {\"kind\": \"turnFlagOn\"} ]
}
```

## Errors

The API always returns errors in a common format. Here's an example:

```json
{
  \"code\": \"invalid_request\",
  \"message\": \"A feature with that key already exists\",
  \"id\": \"30ce6058-87da-11e4-b116-123b93f75cba\"
}
```

The `code` indicates the general class of error. The `message` is a human-readable explanation of what went wrong. The `id` is a unique identifier. Use it when you're working with LaunchDarkly Support to debug a problem with a specific API call.

### HTTP status error response codes

| Code | Definition        | Description                                                                                       | Possible Solution                                                |
| ---- | ----------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| 400  | Invalid request       | The request cannot be understood.                                    | Ensure JSON syntax in request body is correct.                   |
| 401  | Invalid access token      | Requestor is unauthorized or does not have permission for this API call.                                                | Ensure your API access token is valid and has the appropriate permissions.                                     |
| 403  | Forbidden         | Requestor does not have access to this resource.                                                | Ensure that the account member or access token has proper permissions set. |
| 404  | Invalid resource identifier | The requested resource is not valid. | Ensure that the resource is correctly identified by ID or key. |
| 405  | Method not allowed | The request method is not allowed on this resource. | Ensure that the HTTP verb is correct. |
| 409  | Conflict          | The API request can not be completed because it conflicts with a concurrent API request. | Retry your request.                                              |
| 422  | Unprocessable entity | The API request can not be completed because the update description can not be understood. | Ensure that the request body is correct for the type of patch you are using, either JSON patch or semantic patch.
| 429  | Too many requests | Read [Rate limiting](/#section/Overview/Rate-limiting).                                               | Wait and try again later.                                        |

## CORS

The LaunchDarkly API supports Cross Origin Resource Sharing (CORS) for AJAX requests from any origin. If an `Origin` header is given in a request, it will be echoed as an explicitly allowed origin. Otherwise the request returns a wildcard, `Access-Control-Allow-Origin: *`. For more information on CORS, read the [CORS W3C Recommendation](http://www.w3.org/TR/cors). Example CORS headers might look like:

```http
Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization
Access-Control-Allow-Methods: OPTIONS, GET, DELETE, PATCH
Access-Control-Allow-Origin: *
Access-Control-Max-Age: 300
```

You can make authenticated CORS calls just as you would make same-origin calls, using either [token or session-based authentication](/#section/Overview/Authentication). If you are using session authentication, you should set the `withCredentials` property for your `xhr` request to `true`. You should never expose your access tokens to untrusted entities.

## Rate limiting

We use several rate limiting strategies to ensure the availability of our APIs. Rate-limited calls to our APIs return a `429` status code. Calls to our APIs include headers indicating the current rate limit status. The specific headers returned depend on the API route being called. The limits differ based on the route, authentication mechanism, and other factors. Routes that are not rate limited may not contain any of the headers described below.

> ### Rate limiting and SDKs
>
> LaunchDarkly SDKs are never rate limited and do not use the API endpoints defined here. LaunchDarkly uses a different set of approaches, including streaming/server-sent events and a global CDN, to ensure availability to the routes used by LaunchDarkly SDKs.

### Global rate limits

Authenticated requests are subject to a global limit. This is the maximum number of calls that your account can make to the API per ten seconds. All service and personal access tokens on the account share this limit, so exceeding the limit with one access token will impact other tokens. Calls that are subject to global rate limits may return the headers below:

| Header name                    | Description                                                                      |
| ------------------------------ | -------------------------------------------------------------------------------- |
| `X-Ratelimit-Global-Remaining` | The maximum number of requests the account is permitted to make per ten seconds. |
| `X-Ratelimit-Reset`            | The time at which the current rate limit window resets in epoch milliseconds.    |

We do not publicly document the specific number of calls that can be made globally. This limit may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limit.

### Route-level rate limits

Some authenticated routes have custom rate limits. These also reset every ten seconds. Any service or personal access tokens hitting the same route share this limit, so exceeding the limit with one access token may impact other tokens. Calls that are subject to route-level rate limits return the headers below:

| Header name                   | Description                                                                                           |
| ----------------------------- | ----------------------------------------------------------------------------------------------------- |
| `X-Ratelimit-Route-Remaining` | The maximum number of requests to the current route the account is permitted to make per ten seconds. |
| `X-Ratelimit-Reset`           | The time at which the current rate limit window resets in epoch milliseconds.                         |

A _route_ represents a specific URL pattern and verb. For example, the [Delete environment](/tag/Environments#operation/deleteEnvironment) endpoint is considered a single route, and each call to delete an environment counts against your route-level rate limit for that route.

We do not publicly document the specific number of calls that an account can make to each endpoint per ten seconds. These limits may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limits.

### IP-based rate limiting

We also employ IP-based rate limiting on some API routes. If you hit an IP-based rate limit, your API response will include a `Retry-After` header indicating how long to wait before re-trying the call. Clients must wait at least `Retry-After` seconds before making additional calls to our API, and should employ jitter and backoff strategies to avoid triggering rate limits again.

## OpenAPI (Swagger) and client libraries

We have a [complete OpenAPI (Swagger) specification](https://app.launchdarkly.com/api/v2/openapi.json) for our API.

We auto-generate multiple client libraries based on our OpenAPI specification. To learn more, visit the [collection of client libraries on GitHub](https://github.com/search?q=topic%3Alaunchdarkly-api+org%3Alaunchdarkly&type=Repositories). You can also use this specification to generate client libraries to interact with our REST API in your language of choice.

Our OpenAPI specification is supported by several API-based tools such as Postman and Insomnia. In many cases, you can directly import our specification to explore our APIs.

## Method overriding

Some firewalls and HTTP clients restrict the use of verbs other than `GET` and `POST`. In those environments, our API endpoints that use `DELETE`, `PATCH`, and `PUT` verbs are inaccessible.

To avoid this issue, our API supports the `X-HTTP-Method-Override` header, allowing clients to \"tunnel\" `DELETE`, `PATCH`, and `PUT` requests using a `POST` request.

For example, to call a `PATCH` endpoint using a `POST` request, you can include `X-HTTP-Method-Override:PATCH` as a header.

## Beta resources

We sometimes release new API resources in **beta** status before we release them with general availability.

Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.

We try to promote resources into general availability as quickly as possible. This happens after sufficient testing and when we're satisfied that we no longer need to make backwards-incompatible changes.

We mark beta resources with a \"Beta\" callout in our documentation, pictured below:

> ### This feature is in beta
>
> To use this feature, pass in a header including the `LD-API-Version` key with value set to `beta`. Use this header with each call. To learn more, read [Beta resources](/#section/Overview/Beta-resources).
>
> Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.

### Using beta resources

To use a beta resource, you must include a header in the request. If you call a beta resource without this header, you receive a `403` response.

Use this header:

```
LD-API-Version: beta
```

## Federal environments

The version of LaunchDarkly that is available on domains controlled by the United States government is different from the version of LaunchDarkly available to the general public. If you are an employee or contractor for a United States federal agency and use LaunchDarkly in your work, you likely use the federal instance of LaunchDarkly.

If you are working in the federal instance of LaunchDarkly, the base URI for each request is `https://app.launchdarkly.us`. In the \"Try it\" sandbox for each request, click the request path to view the complete resource path for the federal environment.

To learn more, read [LaunchDarkly in federal environments](https://docs.launchdarkly.com/home/advanced/federal).

## Versioning

We try hard to keep our REST API backwards compatible, but we occasionally have to make backwards-incompatible changes in the process of shipping new features. These breaking changes can cause unexpected behavior if you don't prepare for them accordingly.

Updates to our REST API include support for the latest features in LaunchDarkly. We also release a new version of our REST API every time we make a breaking change. We provide simultaneous support for multiple API versions so you can migrate from your current API version to a new version at your own pace.

### Setting the API version per request

You can set the API version on a specific request by sending an `LD-API-Version` header, as shown in the example below:

```
LD-API-Version: 20220603
```

The header value is the version number of the API version you would like to request. The number for each version corresponds to the date the version was released in `yyyymmdd` format. In the example above the version `20220603` corresponds to June 03, 2022.

### Setting the API version per access token

When you create an access token, you must specify a specific version of the API to use. This ensures that integrations using this token cannot be broken by version changes.

Tokens created before versioning was released have their version set to `20160426`, which is the version of the API that existed before the current versioning scheme, so that they continue working the same way they did before versioning.

If you would like to upgrade your integration to use a new API version, you can explicitly set the header described above.

> ### Best practice: Set the header for every client or integration
>
> We recommend that you set the API version header explicitly in any client or integration you build.
>
> Only rely on the access token API version during manual testing.

### API version changelog

|<div style=\"width:75px\">Version</div> | Changes | End of life (EOL)
|---|---|---|
| `20220603` | <ul><li>Changed the [list projects](/tag/Projects#operation/getProjects) return value:<ul><li>Response is now paginated with a default limit of `20`.</li><li>Added support for filter and sort.</li><li>The project `environments` field is now expandable. This field is omitted by default.</li></ul></li><li>Changed the [get project](/tag/Projects#operation/getProject) return value:<ul><li>The `environments` field is now expandable. This field is omitted by default.</li></ul></li></ul> | Current |
| `20210729` | <ul><li>Changed the [create approval request](/tag/Approvals#operation/postApprovalRequest) return value. It now returns HTTP Status Code `201` instead of `200`.</li><li> Changed the [get users](/tag/Users#operation/getUser) return value. It now returns a user record, not a user. </li><li>Added additional optional fields to environment, segments, flags, members, and segments, including the ability to create Big Segments. </li><li> Added default values for flag variations when new environments are created. </li><li>Added filtering and pagination for getting flags and members, including `limit`, `number`, `filter`, and `sort` query parameters. </li><li>Added endpoints for expiring user targets for flags and segments, scheduled changes, access tokens, Relay Proxy configuration, integrations and subscriptions, and approvals. </li></ul> | 2023-06-03 |
| `20191212` | <ul><li>[List feature flags](/tag/Feature-flags#operation/getFeatureFlags) now defaults to sending summaries of feature flag configurations, equivalent to setting the query parameter `summary=true`. Summaries omit flag targeting rules and individual user targets from the payload. </li><li> Added endpoints for flags, flag status, projects, environments, audit logs, members, users, custom roles, segments, usage, streams, events, and data export. </li></ul> | 2022-07-29 |
| `20160426` | <ul><li>Initial versioning of API. Tokens created before versioning have their version set to this.</li></ul> | 2020-12-12 |


## Overview
This API client was generated by the [OpenAPI Generator](https://openapi-generator.tech) project.  By using the [OpenAPI-spec](https://www.openapis.org/) from a remote server, you can easily generate an API client.

- API version: 2.0
- Package version: 14
- Build package: org.openapitools.codegen.languages.GoClientCodegen
For more information, please visit [https://support.launchdarkly.com](https://support.launchdarkly.com)

## Installation

Install the following dependencies:

```shell
go get github.com/stretchr/testify/assert
go get golang.org/x/oauth2
go get golang.org/x/net/context
```

Put the package under your project folder and add the following in import:

```golang
import ldapi "github.com/launchdarkly/api-client-go"
```

To use a proxy, set the environment variable `HTTP_PROXY`:

```golang
os.Setenv("HTTP_PROXY", "http://proxy_name:proxy_port")
```

## Configuration of Server URL

Default configuration comes with `Servers` field that contains server objects as defined in the OpenAPI specification.

### Select Server Configuration

For using other server than the one defined on index 0 set context value `sw.ContextServerIndex` of type `int`.

```golang
ctx := context.WithValue(context.Background(), ldapi.ContextServerIndex, 1)
```

### Templated Server URL

Templated server URL is formatted using default variables from configuration or from context value `sw.ContextServerVariables` of type `map[string]string`.

```golang
ctx := context.WithValue(context.Background(), ldapi.ContextServerVariables, map[string]string{
	"basePath": "v2",
})
```

Note, enum values are always validated and all unused variables are silently ignored.

### URLs Configuration per Operation

Each operation can use different server URL defined using `OperationServers` map in the `Configuration`.
An operation is uniquely identified by `"{classname}Service.{nickname}"` string.
Similar rules for overriding default operation server index and variables applies by using `sw.ContextOperationServerIndices` and `sw.ContextOperationServerVariables` context maps.

```
ctx := context.WithValue(context.Background(), ldapi.ContextOperationServerIndices, map[string]int{
	"{classname}Service.{nickname}": 2,
})
ctx = context.WithValue(context.Background(), ldapi.ContextOperationServerVariables, map[string]map[string]string{
	"{classname}Service.{nickname}": {
		"port": "8443",
	},
})
```

## Documentation for API Endpoints

All URIs are relative to *https://app.launchdarkly.com*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AccessTokensApi* | [**DeleteToken**](docs/AccessTokensApi.md#deletetoken) | **Delete** /api/v2/tokens/{id} | Delete access token
*AccessTokensApi* | [**GetToken**](docs/AccessTokensApi.md#gettoken) | **Get** /api/v2/tokens/{id} | Get access token
*AccessTokensApi* | [**GetTokens**](docs/AccessTokensApi.md#gettokens) | **Get** /api/v2/tokens | List access tokens
*AccessTokensApi* | [**PatchToken**](docs/AccessTokensApi.md#patchtoken) | **Patch** /api/v2/tokens/{id} | Patch access token
*AccessTokensApi* | [**PostToken**](docs/AccessTokensApi.md#posttoken) | **Post** /api/v2/tokens | Create access token
*AccessTokensApi* | [**ResetToken**](docs/AccessTokensApi.md#resettoken) | **Post** /api/v2/tokens/{id}/reset | Reset access token
*AccountMembersApi* | [**DeleteMember**](docs/AccountMembersApi.md#deletemember) | **Delete** /api/v2/members/{id} | Delete account member
*AccountMembersApi* | [**GetMember**](docs/AccountMembersApi.md#getmember) | **Get** /api/v2/members/{id} | Get account member
*AccountMembersApi* | [**GetMembers**](docs/AccountMembersApi.md#getmembers) | **Get** /api/v2/members | List account members
*AccountMembersApi* | [**PatchMember**](docs/AccountMembersApi.md#patchmember) | **Patch** /api/v2/members/{id} | Modify an account member
*AccountMembersApi* | [**PostMemberTeams**](docs/AccountMembersApi.md#postmemberteams) | **Post** /api/v2/members/{id}/teams | Add a member to teams
*AccountMembersApi* | [**PostMembers**](docs/AccountMembersApi.md#postmembers) | **Post** /api/v2/members | Invite new members
*AccountMembersBetaApi* | [**PatchMembers**](docs/AccountMembersBetaApi.md#patchmembers) | **Patch** /api/v2/members | Modify account members
*AccountUsageBetaApi* | [**GetEvaluationsUsage**](docs/AccountUsageBetaApi.md#getevaluationsusage) | **Get** /api/v2/usage/evaluations/{projectKey}/{environmentKey}/{featureFlagKey} | Get evaluations usage
*AccountUsageBetaApi* | [**GetEventsUsage**](docs/AccountUsageBetaApi.md#geteventsusage) | **Get** /api/v2/usage/events/{type} | Get events usage
*AccountUsageBetaApi* | [**GetExperimentationKeysUsage**](docs/AccountUsageBetaApi.md#getexperimentationkeysusage) | **Get** /api/v2/usage/experimentation-keys | Get experimentation keys usage
*AccountUsageBetaApi* | [**GetExperimentationUnitsUsage**](docs/AccountUsageBetaApi.md#getexperimentationunitsusage) | **Get** /api/v2/usage/experimentation-units | Get experimentation units usage
*AccountUsageBetaApi* | [**GetMauSdksByType**](docs/AccountUsageBetaApi.md#getmausdksbytype) | **Get** /api/v2/usage/mau/sdks | Get MAU SDKs by type
*AccountUsageBetaApi* | [**GetMauUsage**](docs/AccountUsageBetaApi.md#getmauusage) | **Get** /api/v2/usage/mau | Get MAU usage
*AccountUsageBetaApi* | [**GetMauUsageByCategory**](docs/AccountUsageBetaApi.md#getmauusagebycategory) | **Get** /api/v2/usage/mau/bycategory | Get MAU usage by category
*AccountUsageBetaApi* | [**GetStreamUsage**](docs/AccountUsageBetaApi.md#getstreamusage) | **Get** /api/v2/usage/streams/{source} | Get stream usage
*AccountUsageBetaApi* | [**GetStreamUsageBySdkVersion**](docs/AccountUsageBetaApi.md#getstreamusagebysdkversion) | **Get** /api/v2/usage/streams/{source}/bysdkversion | Get stream usage by SDK version
*AccountUsageBetaApi* | [**GetStreamUsageSdkversion**](docs/AccountUsageBetaApi.md#getstreamusagesdkversion) | **Get** /api/v2/usage/streams/{source}/sdkversions | Get stream usage SDK versions
*ApplicationsBetaApi* | [**CreateApplication**](docs/ApplicationsBetaApi.md#createapplication) | **Post** /api/v2/applications | Post application
*ApplicationsBetaApi* | [**DeleteApplication**](docs/ApplicationsBetaApi.md#deleteapplication) | **Delete** /api/v2/applications/{applicationKey} | Delete application
*ApplicationsBetaApi* | [**DeleteApplicationVersion**](docs/ApplicationsBetaApi.md#deleteapplicationversion) | **Delete** /api/v2/applications/{applicationKey}/versions/{versionKey} | Delete application version
*ApplicationsBetaApi* | [**GetApplication**](docs/ApplicationsBetaApi.md#getapplication) | **Get** /api/v2/applications/{applicationKey} | Get application by key
*ApplicationsBetaApi* | [**GetApplicationVersions**](docs/ApplicationsBetaApi.md#getapplicationversions) | **Get** /api/v2/applications/{applicationKey}/versions | Get application versions by application key
*ApplicationsBetaApi* | [**GetApplications**](docs/ApplicationsBetaApi.md#getapplications) | **Get** /api/v2/applications | Get applications
*ApplicationsBetaApi* | [**PatchApplication**](docs/ApplicationsBetaApi.md#patchapplication) | **Patch** /api/v2/applications/{applicationKey} | Update application
*ApplicationsBetaApi* | [**PatchApplicationVersion**](docs/ApplicationsBetaApi.md#patchapplicationversion) | **Patch** /api/v2/applications/{applicationKey}/versions/{versionKey} | Update application version
*ApplicationsBetaApi* | [**PostApplicationVersion**](docs/ApplicationsBetaApi.md#postapplicationversion) | **Post** /api/v2/applications/{applicationKey}/versions | Post application version
*ApprovalsApi* | [**DeleteApprovalRequest**](docs/ApprovalsApi.md#deleteapprovalrequest) | **Delete** /api/v2/approval-requests/{id} | Delete approval request
*ApprovalsApi* | [**DeleteApprovalRequestForFlag**](docs/ApprovalsApi.md#deleteapprovalrequestforflag) | **Delete** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests/{id} | Delete approval request for a flag
*ApprovalsApi* | [**GetApprovalForFlag**](docs/ApprovalsApi.md#getapprovalforflag) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests/{id} | Get approval request for a flag
*ApprovalsApi* | [**GetApprovalRequest**](docs/ApprovalsApi.md#getapprovalrequest) | **Get** /api/v2/approval-requests/{id} | Get approval request
*ApprovalsApi* | [**GetApprovalRequests**](docs/ApprovalsApi.md#getapprovalrequests) | **Get** /api/v2/approval-requests | List approval requests
*ApprovalsApi* | [**GetApprovalsForFlag**](docs/ApprovalsApi.md#getapprovalsforflag) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests | List approval requests for a flag
*ApprovalsApi* | [**PostApprovalRequest**](docs/ApprovalsApi.md#postapprovalrequest) | **Post** /api/v2/approval-requests | Create approval request
*ApprovalsApi* | [**PostApprovalRequestApply**](docs/ApprovalsApi.md#postapprovalrequestapply) | **Post** /api/v2/approval-requests/{id}/apply | Apply approval request
*ApprovalsApi* | [**PostApprovalRequestApplyForFlag**](docs/ApprovalsApi.md#postapprovalrequestapplyforflag) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests/{id}/apply | Apply approval request for a flag
*ApprovalsApi* | [**PostApprovalRequestForFlag**](docs/ApprovalsApi.md#postapprovalrequestforflag) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests | Create approval request for a flag
*ApprovalsApi* | [**PostApprovalRequestReview**](docs/ApprovalsApi.md#postapprovalrequestreview) | **Post** /api/v2/approval-requests/{id}/reviews | Review approval request
*ApprovalsApi* | [**PostApprovalRequestReviewForFlag**](docs/ApprovalsApi.md#postapprovalrequestreviewforflag) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests/{id}/reviews | Review approval request for a flag
*ApprovalsApi* | [**PostFlagCopyConfigApprovalRequest**](docs/ApprovalsApi.md#postflagcopyconfigapprovalrequest) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/approval-requests-flag-copy | Create approval request to copy flag configurations across environments
*AuditLogApi* | [**GetAuditLogEntries**](docs/AuditLogApi.md#getauditlogentries) | **Get** /api/v2/auditlog | List audit log entries
*AuditLogApi* | [**GetAuditLogEntry**](docs/AuditLogApi.md#getauditlogentry) | **Get** /api/v2/auditlog/{id} | Get audit log entry
*CodeReferencesApi* | [**DeleteBranches**](docs/CodeReferencesApi.md#deletebranches) | **Post** /api/v2/code-refs/repositories/{repo}/branch-delete-tasks | Delete branches
*CodeReferencesApi* | [**DeleteRepository**](docs/CodeReferencesApi.md#deleterepository) | **Delete** /api/v2/code-refs/repositories/{repo} | Delete repository
*CodeReferencesApi* | [**GetBranch**](docs/CodeReferencesApi.md#getbranch) | **Get** /api/v2/code-refs/repositories/{repo}/branches/{branch} | Get branch
*CodeReferencesApi* | [**GetBranches**](docs/CodeReferencesApi.md#getbranches) | **Get** /api/v2/code-refs/repositories/{repo}/branches | List branches
*CodeReferencesApi* | [**GetExtinctions**](docs/CodeReferencesApi.md#getextinctions) | **Get** /api/v2/code-refs/extinctions | List extinctions
*CodeReferencesApi* | [**GetRepositories**](docs/CodeReferencesApi.md#getrepositories) | **Get** /api/v2/code-refs/repositories | List repositories
*CodeReferencesApi* | [**GetRepository**](docs/CodeReferencesApi.md#getrepository) | **Get** /api/v2/code-refs/repositories/{repo} | Get repository
*CodeReferencesApi* | [**GetRootStatistic**](docs/CodeReferencesApi.md#getrootstatistic) | **Get** /api/v2/code-refs/statistics | Get links to code reference repositories for each project
*CodeReferencesApi* | [**GetStatistics**](docs/CodeReferencesApi.md#getstatistics) | **Get** /api/v2/code-refs/statistics/{projectKey} | Get code references statistics for flags
*CodeReferencesApi* | [**PatchRepository**](docs/CodeReferencesApi.md#patchrepository) | **Patch** /api/v2/code-refs/repositories/{repo} | Update repository
*CodeReferencesApi* | [**PostExtinction**](docs/CodeReferencesApi.md#postextinction) | **Post** /api/v2/code-refs/repositories/{repo}/branches/{branch}/extinction-events | Create extinction
*CodeReferencesApi* | [**PostRepository**](docs/CodeReferencesApi.md#postrepository) | **Post** /api/v2/code-refs/repositories | Create repository
*CodeReferencesApi* | [**PutBranch**](docs/CodeReferencesApi.md#putbranch) | **Put** /api/v2/code-refs/repositories/{repo}/branches/{branch} | Upsert branch
*ContextSettingsApi* | [**PutContextFlagSetting**](docs/ContextSettingsApi.md#putcontextflagsetting) | **Put** /api/v2/projects/{projectKey}/environments/{environmentKey}/contexts/{contextKind}/{contextKey}/flags/{featureFlagKey} | Update flag settings for context
*ContextsApi* | [**DeleteContextInstances**](docs/ContextsApi.md#deletecontextinstances) | **Delete** /api/v2/projects/{projectKey}/environments/{environmentKey}/context-instances/{id} | Delete context instances
*ContextsApi* | [**EvaluateContextInstance**](docs/ContextsApi.md#evaluatecontextinstance) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/flags/evaluate | Evaluate flags for context instance
*ContextsApi* | [**GetContextAttributeNames**](docs/ContextsApi.md#getcontextattributenames) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/context-attributes | Get context attribute names
*ContextsApi* | [**GetContextAttributeValues**](docs/ContextsApi.md#getcontextattributevalues) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/context-attributes/{attributeName} | Get context attribute values
*ContextsApi* | [**GetContextInstances**](docs/ContextsApi.md#getcontextinstances) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/context-instances/{id} | Get context instances
*ContextsApi* | [**GetContextKindsByProjectKey**](docs/ContextsApi.md#getcontextkindsbyprojectkey) | **Get** /api/v2/projects/{projectKey}/context-kinds | Get context kinds
*ContextsApi* | [**GetContexts**](docs/ContextsApi.md#getcontexts) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/contexts/{kind}/{key} | Get contexts
*ContextsApi* | [**PutContextKind**](docs/ContextsApi.md#putcontextkind) | **Put** /api/v2/projects/{projectKey}/context-kinds/{key} | Create or update context kind
*ContextsApi* | [**SearchContextInstances**](docs/ContextsApi.md#searchcontextinstances) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/context-instances/search | Search for context instances
*ContextsApi* | [**SearchContexts**](docs/ContextsApi.md#searchcontexts) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/contexts/search | Search for contexts
*CustomRolesApi* | [**DeleteCustomRole**](docs/CustomRolesApi.md#deletecustomrole) | **Delete** /api/v2/roles/{customRoleKey} | Delete custom role
*CustomRolesApi* | [**GetCustomRole**](docs/CustomRolesApi.md#getcustomrole) | **Get** /api/v2/roles/{customRoleKey} | Get custom role
*CustomRolesApi* | [**GetCustomRoles**](docs/CustomRolesApi.md#getcustomroles) | **Get** /api/v2/roles | List custom roles
*CustomRolesApi* | [**PatchCustomRole**](docs/CustomRolesApi.md#patchcustomrole) | **Patch** /api/v2/roles/{customRoleKey} | Update custom role
*CustomRolesApi* | [**PostCustomRole**](docs/CustomRolesApi.md#postcustomrole) | **Post** /api/v2/roles | Create custom role
*DataExportDestinationsApi* | [**DeleteDestination**](docs/DataExportDestinationsApi.md#deletedestination) | **Delete** /api/v2/destinations/{projectKey}/{environmentKey}/{id} | Delete Data Export destination
*DataExportDestinationsApi* | [**GetDestination**](docs/DataExportDestinationsApi.md#getdestination) | **Get** /api/v2/destinations/{projectKey}/{environmentKey}/{id} | Get destination
*DataExportDestinationsApi* | [**GetDestinations**](docs/DataExportDestinationsApi.md#getdestinations) | **Get** /api/v2/destinations | List destinations
*DataExportDestinationsApi* | [**PatchDestination**](docs/DataExportDestinationsApi.md#patchdestination) | **Patch** /api/v2/destinations/{projectKey}/{environmentKey}/{id} | Update Data Export destination
*DataExportDestinationsApi* | [**PostDestination**](docs/DataExportDestinationsApi.md#postdestination) | **Post** /api/v2/destinations/{projectKey}/{environmentKey} | Create Data Export destination
*EnvironmentsApi* | [**DeleteEnvironment**](docs/EnvironmentsApi.md#deleteenvironment) | **Delete** /api/v2/projects/{projectKey}/environments/{environmentKey} | Delete environment
*EnvironmentsApi* | [**GetEnvironment**](docs/EnvironmentsApi.md#getenvironment) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey} | Get environment
*EnvironmentsApi* | [**GetEnvironmentsByProject**](docs/EnvironmentsApi.md#getenvironmentsbyproject) | **Get** /api/v2/projects/{projectKey}/environments | List environments
*EnvironmentsApi* | [**PatchEnvironment**](docs/EnvironmentsApi.md#patchenvironment) | **Patch** /api/v2/projects/{projectKey}/environments/{environmentKey} | Update environment
*EnvironmentsApi* | [**PostEnvironment**](docs/EnvironmentsApi.md#postenvironment) | **Post** /api/v2/projects/{projectKey}/environments | Create environment
*EnvironmentsApi* | [**ResetEnvironmentMobileKey**](docs/EnvironmentsApi.md#resetenvironmentmobilekey) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/mobileKey | Reset environment mobile SDK key
*EnvironmentsApi* | [**ResetEnvironmentSDKKey**](docs/EnvironmentsApi.md#resetenvironmentsdkkey) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/apiKey | Reset environment SDK key
*ExperimentsBetaApi* | [**CreateExperiment**](docs/ExperimentsBetaApi.md#createexperiment) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments | Create experiment
*ExperimentsBetaApi* | [**CreateIteration**](docs/ExperimentsBetaApi.md#createiteration) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments/{experimentKey}/iterations | Create iteration
*ExperimentsBetaApi* | [**GetExperiment**](docs/ExperimentsBetaApi.md#getexperiment) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments/{experimentKey} | Get experiment
*ExperimentsBetaApi* | [**GetExperimentResults**](docs/ExperimentsBetaApi.md#getexperimentresults) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments/{experimentKey}/metrics/{metricKey}/results | Get experiment results
*ExperimentsBetaApi* | [**GetExperimentResultsForMetricGroup**](docs/ExperimentsBetaApi.md#getexperimentresultsformetricgroup) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments/{experimentKey}/metric-groups/{metricGroupKey}/results | Get experiment results for metric group
*ExperimentsBetaApi* | [**GetExperimentationSettings**](docs/ExperimentsBetaApi.md#getexperimentationsettings) | **Get** /api/v2/projects/{projectKey}/experimentation-settings | Get experimentation settings
*ExperimentsBetaApi* | [**GetExperiments**](docs/ExperimentsBetaApi.md#getexperiments) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments | Get experiments
*ExperimentsBetaApi* | [**GetLegacyExperimentResults**](docs/ExperimentsBetaApi.md#getlegacyexperimentresults) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/experiments/{environmentKey}/{metricKey} | Get legacy experiment results (deprecated)
*ExperimentsBetaApi* | [**PatchExperiment**](docs/ExperimentsBetaApi.md#patchexperiment) | **Patch** /api/v2/projects/{projectKey}/environments/{environmentKey}/experiments/{experimentKey} | Patch experiment
*ExperimentsBetaApi* | [**PutExperimentationSettings**](docs/ExperimentsBetaApi.md#putexperimentationsettings) | **Put** /api/v2/projects/{projectKey}/experimentation-settings | Update experimentation settings
*ExperimentsBetaApi* | [**ResetExperiment**](docs/ExperimentsBetaApi.md#resetexperiment) | **Delete** /api/v2/flags/{projectKey}/{featureFlagKey}/experiments/{environmentKey}/{metricKey}/results | Reset experiment results
*FeatureFlagsApi* | [**CopyFeatureFlag**](docs/FeatureFlagsApi.md#copyfeatureflag) | **Post** /api/v2/flags/{projectKey}/{featureFlagKey}/copy | Copy feature flag
*FeatureFlagsApi* | [**DeleteFeatureFlag**](docs/FeatureFlagsApi.md#deletefeatureflag) | **Delete** /api/v2/flags/{projectKey}/{featureFlagKey} | Delete feature flag
*FeatureFlagsApi* | [**GetExpiringContextTargets**](docs/FeatureFlagsApi.md#getexpiringcontexttargets) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/expiring-targets/{environmentKey} | Get expiring context targets for feature flag
*FeatureFlagsApi* | [**GetExpiringUserTargets**](docs/FeatureFlagsApi.md#getexpiringusertargets) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/expiring-user-targets/{environmentKey} | Get expiring user targets for feature flag
*FeatureFlagsApi* | [**GetFeatureFlag**](docs/FeatureFlagsApi.md#getfeatureflag) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey} | Get feature flag
*FeatureFlagsApi* | [**GetFeatureFlagStatus**](docs/FeatureFlagsApi.md#getfeatureflagstatus) | **Get** /api/v2/flag-statuses/{projectKey}/{environmentKey}/{featureFlagKey} | Get feature flag status
*FeatureFlagsApi* | [**GetFeatureFlagStatusAcrossEnvironments**](docs/FeatureFlagsApi.md#getfeatureflagstatusacrossenvironments) | **Get** /api/v2/flag-status/{projectKey}/{featureFlagKey} | Get flag status across environments
*FeatureFlagsApi* | [**GetFeatureFlagStatuses**](docs/FeatureFlagsApi.md#getfeatureflagstatuses) | **Get** /api/v2/flag-statuses/{projectKey}/{environmentKey} | List feature flag statuses
*FeatureFlagsApi* | [**GetFeatureFlags**](docs/FeatureFlagsApi.md#getfeatureflags) | **Get** /api/v2/flags/{projectKey} | List feature flags
*FeatureFlagsApi* | [**PatchExpiringTargets**](docs/FeatureFlagsApi.md#patchexpiringtargets) | **Patch** /api/v2/flags/{projectKey}/{featureFlagKey}/expiring-targets/{environmentKey} | Update expiring context targets on feature flag
*FeatureFlagsApi* | [**PatchExpiringUserTargets**](docs/FeatureFlagsApi.md#patchexpiringusertargets) | **Patch** /api/v2/flags/{projectKey}/{featureFlagKey}/expiring-user-targets/{environmentKey} | Update expiring user targets on feature flag
*FeatureFlagsApi* | [**PatchFeatureFlag**](docs/FeatureFlagsApi.md#patchfeatureflag) | **Patch** /api/v2/flags/{projectKey}/{featureFlagKey} | Update feature flag
*FeatureFlagsApi* | [**PostFeatureFlag**](docs/FeatureFlagsApi.md#postfeatureflag) | **Post** /api/v2/flags/{projectKey} | Create a feature flag
*FeatureFlagsBetaApi* | [**GetDependentFlags**](docs/FeatureFlagsBetaApi.md#getdependentflags) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/dependent-flags | List dependent feature flags
*FeatureFlagsBetaApi* | [**GetDependentFlagsByEnv**](docs/FeatureFlagsBetaApi.md#getdependentflagsbyenv) | **Get** /api/v2/flags/{projectKey}/{environmentKey}/{featureFlagKey}/dependent-flags | List dependent feature flags by environment
*FeatureFlagsBetaApi* | [**PostMigrationSafetyIssues**](docs/FeatureFlagsBetaApi.md#postmigrationsafetyissues) | **Post** /api/v2/projects/{projectKey}/flags/{flagKey}/environments/{environmentKey}/migration-safety-issues | Get migration safety issues
*FlagLinksBetaApi* | [**CreateFlagLink**](docs/FlagLinksBetaApi.md#createflaglink) | **Post** /api/v2/flag-links/projects/{projectKey}/flags/{featureFlagKey} | Create flag link
*FlagLinksBetaApi* | [**DeleteFlagLink**](docs/FlagLinksBetaApi.md#deleteflaglink) | **Delete** /api/v2/flag-links/projects/{projectKey}/flags/{featureFlagKey}/{id} | Delete flag link
*FlagLinksBetaApi* | [**GetFlagLinks**](docs/FlagLinksBetaApi.md#getflaglinks) | **Get** /api/v2/flag-links/projects/{projectKey}/flags/{featureFlagKey} | List flag links
*FlagLinksBetaApi* | [**UpdateFlagLink**](docs/FlagLinksBetaApi.md#updateflaglink) | **Patch** /api/v2/flag-links/projects/{projectKey}/flags/{featureFlagKey}/{id} | Update flag link
*FlagTriggersApi* | [**CreateTriggerWorkflow**](docs/FlagTriggersApi.md#createtriggerworkflow) | **Post** /api/v2/flags/{projectKey}/{featureFlagKey}/triggers/{environmentKey} | Create flag trigger
*FlagTriggersApi* | [**DeleteTriggerWorkflow**](docs/FlagTriggersApi.md#deletetriggerworkflow) | **Delete** /api/v2/flags/{projectKey}/{featureFlagKey}/triggers/{environmentKey}/{id} | Delete flag trigger
*FlagTriggersApi* | [**GetTriggerWorkflowById**](docs/FlagTriggersApi.md#gettriggerworkflowbyid) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/triggers/{environmentKey}/{id} | Get flag trigger by ID
*FlagTriggersApi* | [**GetTriggerWorkflows**](docs/FlagTriggersApi.md#gettriggerworkflows) | **Get** /api/v2/flags/{projectKey}/{featureFlagKey}/triggers/{environmentKey} | List flag triggers
*FlagTriggersApi* | [**PatchTriggerWorkflow**](docs/FlagTriggersApi.md#patchtriggerworkflow) | **Patch** /api/v2/flags/{projectKey}/{featureFlagKey}/triggers/{environmentKey}/{id} | Update flag trigger
*FollowFlagsApi* | [**DeleteFlagFollowers**](docs/FollowFlagsApi.md#deleteflagfollowers) | **Delete** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/followers/{memberId} | Remove a member as a follower of a flag in a project and environment
*FollowFlagsApi* | [**GetFlagFollowers**](docs/FollowFlagsApi.md#getflagfollowers) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/followers | Get followers of a flag in a project and environment
*FollowFlagsApi* | [**GetFollowersByProjEnv**](docs/FollowFlagsApi.md#getfollowersbyprojenv) | **Get** /api/v2/projects/{projectKey}/environments/{environmentKey}/followers | Get followers of all flags in a given project and environment
*FollowFlagsApi* | [**PutFlagFollowers**](docs/FollowFlagsApi.md#putflagfollowers) | **Put** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/followers/{memberId} | Add a member as a follower of a flag in a project and environment
*IntegrationAuditLogSubscriptionsApi* | [**CreateSubscription**](docs/IntegrationAuditLogSubscriptionsApi.md#createsubscription) | **Post** /api/v2/integrations/{integrationKey} | Create audit log subscription
*IntegrationAuditLogSubscriptionsApi* | [**DeleteSubscription**](docs/IntegrationAuditLogSubscriptionsApi.md#deletesubscription) | **Delete** /api/v2/integrations/{integrationKey}/{id} | Delete audit log subscription
*IntegrationAuditLogSubscriptionsApi* | [**GetSubscriptionByID**](docs/IntegrationAuditLogSubscriptionsApi.md#getsubscriptionbyid) | **Get** /api/v2/integrations/{integrationKey}/{id} | Get audit log subscription by ID
*IntegrationAuditLogSubscriptionsApi* | [**GetSubscriptions**](docs/IntegrationAuditLogSubscriptionsApi.md#getsubscriptions) | **Get** /api/v2/integrations/{integrationKey} | Get audit log subscriptions by integration
*IntegrationAuditLogSubscriptionsApi* | [**UpdateSubscription**](docs/IntegrationAuditLogSubscriptionsApi.md#updatesubscription) | **Patch** /api/v2/integrations/{integrationKey}/{id} | Update audit log subscription
*IntegrationDeliveryConfigurationsBetaApi* | [**CreateIntegrationDeliveryConfiguration**](docs/IntegrationDeliveryConfigurationsBetaApi.md#createintegrationdeliveryconfiguration) | **Post** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey}/{integrationKey} | Create delivery configuration
*IntegrationDeliveryConfigurationsBetaApi* | [**DeleteIntegrationDeliveryConfiguration**](docs/IntegrationDeliveryConfigurationsBetaApi.md#deleteintegrationdeliveryconfiguration) | **Delete** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey}/{integrationKey}/{id} | Delete delivery configuration
*IntegrationDeliveryConfigurationsBetaApi* | [**GetIntegrationDeliveryConfigurationByEnvironment**](docs/IntegrationDeliveryConfigurationsBetaApi.md#getintegrationdeliveryconfigurationbyenvironment) | **Get** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey} | Get delivery configurations by environment
*IntegrationDeliveryConfigurationsBetaApi* | [**GetIntegrationDeliveryConfigurationById**](docs/IntegrationDeliveryConfigurationsBetaApi.md#getintegrationdeliveryconfigurationbyid) | **Get** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey}/{integrationKey}/{id} | Get delivery configuration by ID
*IntegrationDeliveryConfigurationsBetaApi* | [**GetIntegrationDeliveryConfigurations**](docs/IntegrationDeliveryConfigurationsBetaApi.md#getintegrationdeliveryconfigurations) | **Get** /api/v2/integration-capabilities/featureStore | List all delivery configurations
*IntegrationDeliveryConfigurationsBetaApi* | [**PatchIntegrationDeliveryConfiguration**](docs/IntegrationDeliveryConfigurationsBetaApi.md#patchintegrationdeliveryconfiguration) | **Patch** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey}/{integrationKey}/{id} | Update delivery configuration
*IntegrationDeliveryConfigurationsBetaApi* | [**ValidateIntegrationDeliveryConfiguration**](docs/IntegrationDeliveryConfigurationsBetaApi.md#validateintegrationdeliveryconfiguration) | **Post** /api/v2/integration-capabilities/featureStore/{projectKey}/{environmentKey}/{integrationKey}/{id}/validate | Validate delivery configuration
*MetricsApi* | [**DeleteMetric**](docs/MetricsApi.md#deletemetric) | **Delete** /api/v2/metrics/{projectKey}/{metricKey} | Delete metric
*MetricsApi* | [**GetMetric**](docs/MetricsApi.md#getmetric) | **Get** /api/v2/metrics/{projectKey}/{metricKey} | Get metric
*MetricsApi* | [**GetMetrics**](docs/MetricsApi.md#getmetrics) | **Get** /api/v2/metrics/{projectKey} | List metrics
*MetricsApi* | [**PatchMetric**](docs/MetricsApi.md#patchmetric) | **Patch** /api/v2/metrics/{projectKey}/{metricKey} | Update metric
*MetricsApi* | [**PostMetric**](docs/MetricsApi.md#postmetric) | **Post** /api/v2/metrics/{projectKey} | Create metric
*MetricsBetaApi* | [**CreateMetricGroup**](docs/MetricsBetaApi.md#createmetricgroup) | **Post** /api/v2/projects/{projectKey}/metric-groups | Create metric group
*MetricsBetaApi* | [**DeleteMetricGroup**](docs/MetricsBetaApi.md#deletemetricgroup) | **Delete** /api/v2/projects/{projectKey}/metric-groups/{metricGroupKey} | Delete metric group
*MetricsBetaApi* | [**GetMetricGroup**](docs/MetricsBetaApi.md#getmetricgroup) | **Get** /api/v2/projects/{projectKey}/metric-groups/{metricGroupKey} | Get metric group
*MetricsBetaApi* | [**GetMetricGroups**](docs/MetricsBetaApi.md#getmetricgroups) | **Get** /api/v2/projects/{projectKey}/metric-groups | List metric groups
*MetricsBetaApi* | [**PatchMetricGroup**](docs/MetricsBetaApi.md#patchmetricgroup) | **Patch** /api/v2/projects/{projectKey}/metric-groups/{metricGroupKey} | Patch metric group
*OAuth2ClientsApi* | [**CreateOAuth2Client**](docs/OAuth2ClientsApi.md#createoauth2client) | **Post** /api/v2/oauth/clients | Create a LaunchDarkly OAuth 2.0 client
*OAuth2ClientsApi* | [**DeleteOAuthClient**](docs/OAuth2ClientsApi.md#deleteoauthclient) | **Delete** /api/v2/oauth/clients/{clientId} | Delete OAuth 2.0 client
*OAuth2ClientsApi* | [**GetOAuthClientById**](docs/OAuth2ClientsApi.md#getoauthclientbyid) | **Get** /api/v2/oauth/clients/{clientId} | Get client by ID
*OAuth2ClientsApi* | [**GetOAuthClients**](docs/OAuth2ClientsApi.md#getoauthclients) | **Get** /api/v2/oauth/clients | Get clients
*OAuth2ClientsApi* | [**PatchOAuthClient**](docs/OAuth2ClientsApi.md#patchoauthclient) | **Patch** /api/v2/oauth/clients/{clientId} | Patch client by ID
*OtherApi* | [**GetIps**](docs/OtherApi.md#getips) | **Get** /api/v2/public-ip-list | Gets the public IP list
*OtherApi* | [**GetOpenapiSpec**](docs/OtherApi.md#getopenapispec) | **Get** /api/v2/openapi.json | Gets the OpenAPI spec in json
*OtherApi* | [**GetRoot**](docs/OtherApi.md#getroot) | **Get** /api/v2 | Root resource
*OtherApi* | [**GetVersions**](docs/OtherApi.md#getversions) | **Get** /api/v2/versions | Get version information
*ProjectsApi* | [**DeleteProject**](docs/ProjectsApi.md#deleteproject) | **Delete** /api/v2/projects/{projectKey} | Delete project
*ProjectsApi* | [**GetFlagDefaultsByProject**](docs/ProjectsApi.md#getflagdefaultsbyproject) | **Get** /api/v2/projects/{projectKey}/flag-defaults | Get flag defaults for project
*ProjectsApi* | [**GetProject**](docs/ProjectsApi.md#getproject) | **Get** /api/v2/projects/{projectKey} | Get project
*ProjectsApi* | [**GetProjects**](docs/ProjectsApi.md#getprojects) | **Get** /api/v2/projects | List projects
*ProjectsApi* | [**PatchFlagDefaultsByProject**](docs/ProjectsApi.md#patchflagdefaultsbyproject) | **Patch** /api/v2/projects/{projectKey}/flag-defaults | Update flag default for project
*ProjectsApi* | [**PatchProject**](docs/ProjectsApi.md#patchproject) | **Patch** /api/v2/projects/{projectKey} | Update project
*ProjectsApi* | [**PostProject**](docs/ProjectsApi.md#postproject) | **Post** /api/v2/projects | Create project
*ProjectsApi* | [**PutFlagDefaultsByProject**](docs/ProjectsApi.md#putflagdefaultsbyproject) | **Put** /api/v2/projects/{projectKey}/flag-defaults | Create or update flag defaults for project
*RelayProxyConfigurationsApi* | [**DeleteRelayAutoConfig**](docs/RelayProxyConfigurationsApi.md#deleterelayautoconfig) | **Delete** /api/v2/account/relay-auto-configs/{id} | Delete Relay Proxy config by ID
*RelayProxyConfigurationsApi* | [**GetRelayProxyConfig**](docs/RelayProxyConfigurationsApi.md#getrelayproxyconfig) | **Get** /api/v2/account/relay-auto-configs/{id} | Get Relay Proxy config
*RelayProxyConfigurationsApi* | [**GetRelayProxyConfigs**](docs/RelayProxyConfigurationsApi.md#getrelayproxyconfigs) | **Get** /api/v2/account/relay-auto-configs | List Relay Proxy configs
*RelayProxyConfigurationsApi* | [**PatchRelayAutoConfig**](docs/RelayProxyConfigurationsApi.md#patchrelayautoconfig) | **Patch** /api/v2/account/relay-auto-configs/{id} | Update a Relay Proxy config
*RelayProxyConfigurationsApi* | [**PostRelayAutoConfig**](docs/RelayProxyConfigurationsApi.md#postrelayautoconfig) | **Post** /api/v2/account/relay-auto-configs | Create a new Relay Proxy config
*RelayProxyConfigurationsApi* | [**ResetRelayAutoConfig**](docs/RelayProxyConfigurationsApi.md#resetrelayautoconfig) | **Post** /api/v2/account/relay-auto-configs/{id}/reset | Reset Relay Proxy configuration key
*ReleasePipelinesBetaApi* | [**DeleteReleasePipeline**](docs/ReleasePipelinesBetaApi.md#deletereleasepipeline) | **Delete** /api/v2/projects/{projectKey}/release-pipelines/{pipelineKey} | Delete release pipeline
*ReleasePipelinesBetaApi* | [**GetAllReleasePipelines**](docs/ReleasePipelinesBetaApi.md#getallreleasepipelines) | **Get** /api/v2/projects/{projectKey}/release-pipelines | Get all release pipelines
*ReleasePipelinesBetaApi* | [**GetReleasePipelineByKey**](docs/ReleasePipelinesBetaApi.md#getreleasepipelinebykey) | **Get** /api/v2/projects/{projectKey}/release-pipelines/{pipelineKey} | Get release pipeline by key
*ReleasePipelinesBetaApi* | [**PatchReleasePipeline**](docs/ReleasePipelinesBetaApi.md#patchreleasepipeline) | **Patch** /api/v2/projects/{projectKey}/release-pipelines/{pipelineKey} | Update a release pipeline
*ReleasePipelinesBetaApi* | [**PostReleasePipeline**](docs/ReleasePipelinesBetaApi.md#postreleasepipeline) | **Post** /api/v2/projects/{projectKey}/release-pipelines | Create a release pipeline
*ReleasesBetaApi* | [**GetReleaseByFlagKey**](docs/ReleasesBetaApi.md#getreleasebyflagkey) | **Get** /api/v2/flags/{projectKey}/{flagKey}/release | Get release for flag
*ReleasesBetaApi* | [**PatchReleaseByFlagKey**](docs/ReleasesBetaApi.md#patchreleasebyflagkey) | **Patch** /api/v2/flags/{projectKey}/{flagKey}/release | Patch release for flag
*ScheduledChangesApi* | [**DeleteFlagConfigScheduledChanges**](docs/ScheduledChangesApi.md#deleteflagconfigscheduledchanges) | **Delete** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/scheduled-changes/{id} | Delete scheduled changes workflow
*ScheduledChangesApi* | [**GetFeatureFlagScheduledChange**](docs/ScheduledChangesApi.md#getfeatureflagscheduledchange) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/scheduled-changes/{id} | Get a scheduled change
*ScheduledChangesApi* | [**GetFlagConfigScheduledChanges**](docs/ScheduledChangesApi.md#getflagconfigscheduledchanges) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/scheduled-changes | List scheduled changes
*ScheduledChangesApi* | [**PatchFlagConfigScheduledChange**](docs/ScheduledChangesApi.md#patchflagconfigscheduledchange) | **Patch** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/scheduled-changes/{id} | Update scheduled changes workflow
*ScheduledChangesApi* | [**PostFlagConfigScheduledChanges**](docs/ScheduledChangesApi.md#postflagconfigscheduledchanges) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/scheduled-changes | Create scheduled changes workflow
*SegmentsApi* | [**DeleteSegment**](docs/SegmentsApi.md#deletesegment) | **Delete** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey} | Delete segment
*SegmentsApi* | [**GetContextInstanceSegmentsMembershipByEnv**](docs/SegmentsApi.md#getcontextinstancesegmentsmembershipbyenv) | **Post** /api/v2/projects/{projectKey}/environments/{environmentKey}/segments/evaluate | List segment memberships for context instance
*SegmentsApi* | [**GetExpiringTargetsForSegment**](docs/SegmentsApi.md#getexpiringtargetsforsegment) | **Get** /api/v2/segments/{projectKey}/{segmentKey}/expiring-targets/{environmentKey} | Get expiring targets for segment
*SegmentsApi* | [**GetExpiringUserTargetsForSegment**](docs/SegmentsApi.md#getexpiringusertargetsforsegment) | **Get** /api/v2/segments/{projectKey}/{segmentKey}/expiring-user-targets/{environmentKey} | Get expiring user targets for segment
*SegmentsApi* | [**GetSegment**](docs/SegmentsApi.md#getsegment) | **Get** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey} | Get segment
*SegmentsApi* | [**GetSegmentMembershipForContext**](docs/SegmentsApi.md#getsegmentmembershipforcontext) | **Get** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/contexts/{contextKey} | Get Big Segment membership for context
*SegmentsApi* | [**GetSegmentMembershipForUser**](docs/SegmentsApi.md#getsegmentmembershipforuser) | **Get** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/users/{userKey} | Get Big Segment membership for user
*SegmentsApi* | [**GetSegments**](docs/SegmentsApi.md#getsegments) | **Get** /api/v2/segments/{projectKey}/{environmentKey} | List segments
*SegmentsApi* | [**PatchExpiringTargetsForSegment**](docs/SegmentsApi.md#patchexpiringtargetsforsegment) | **Patch** /api/v2/segments/{projectKey}/{segmentKey}/expiring-targets/{environmentKey} | Update expiring targets for segment
*SegmentsApi* | [**PatchExpiringUserTargetsForSegment**](docs/SegmentsApi.md#patchexpiringusertargetsforsegment) | **Patch** /api/v2/segments/{projectKey}/{segmentKey}/expiring-user-targets/{environmentKey} | Update expiring user targets for segment
*SegmentsApi* | [**PatchSegment**](docs/SegmentsApi.md#patchsegment) | **Patch** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey} | Patch segment
*SegmentsApi* | [**PostSegment**](docs/SegmentsApi.md#postsegment) | **Post** /api/v2/segments/{projectKey}/{environmentKey} | Create segment
*SegmentsApi* | [**UpdateBigSegmentContextTargets**](docs/SegmentsApi.md#updatebigsegmentcontexttargets) | **Post** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/contexts | Update context targets on a Big Segment
*SegmentsApi* | [**UpdateBigSegmentTargets**](docs/SegmentsApi.md#updatebigsegmenttargets) | **Post** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/users | Update user context targets on a Big Segment
*SegmentsBetaApi* | [**CreateBigSegmentExport**](docs/SegmentsBetaApi.md#createbigsegmentexport) | **Post** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/exports | Create Big Segment export
*SegmentsBetaApi* | [**CreateBigSegmentImport**](docs/SegmentsBetaApi.md#createbigsegmentimport) | **Post** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/imports | Create Big Segment import
*SegmentsBetaApi* | [**GetBigSegmentExport**](docs/SegmentsBetaApi.md#getbigsegmentexport) | **Get** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/exports/{exportID} | Get Big Segment export
*SegmentsBetaApi* | [**GetBigSegmentImport**](docs/SegmentsBetaApi.md#getbigsegmentimport) | **Get** /api/v2/segments/{projectKey}/{environmentKey}/{segmentKey}/imports/{importID} | Get Big Segment import
*TagsApi* | [**GetTags**](docs/TagsApi.md#gettags) | **Get** /api/v2/tags | List tags
*TeamsApi* | [**DeleteTeam**](docs/TeamsApi.md#deleteteam) | **Delete** /api/v2/teams/{teamKey} | Delete team
*TeamsApi* | [**GetTeam**](docs/TeamsApi.md#getteam) | **Get** /api/v2/teams/{teamKey} | Get team
*TeamsApi* | [**GetTeamMaintainers**](docs/TeamsApi.md#getteammaintainers) | **Get** /api/v2/teams/{teamKey}/maintainers | Get team maintainers
*TeamsApi* | [**GetTeamRoles**](docs/TeamsApi.md#getteamroles) | **Get** /api/v2/teams/{teamKey}/roles | Get team custom roles
*TeamsApi* | [**GetTeams**](docs/TeamsApi.md#getteams) | **Get** /api/v2/teams | List teams
*TeamsApi* | [**PatchTeam**](docs/TeamsApi.md#patchteam) | **Patch** /api/v2/teams/{teamKey} | Update team
*TeamsApi* | [**PostTeam**](docs/TeamsApi.md#postteam) | **Post** /api/v2/teams | Create team
*TeamsApi* | [**PostTeamMembers**](docs/TeamsApi.md#postteammembers) | **Post** /api/v2/teams/{teamKey}/members | Add multiple members to team
*TeamsBetaApi* | [**PatchTeams**](docs/TeamsBetaApi.md#patchteams) | **Patch** /api/v2/teams | Update teams
*UserSettingsApi* | [**GetExpiringFlagsForUser**](docs/UserSettingsApi.md#getexpiringflagsforuser) | **Get** /api/v2/users/{projectKey}/{userKey}/expiring-user-targets/{environmentKey} | Get expiring dates on flags for user
*UserSettingsApi* | [**GetUserFlagSetting**](docs/UserSettingsApi.md#getuserflagsetting) | **Get** /api/v2/users/{projectKey}/{environmentKey}/{userKey}/flags/{featureFlagKey} | Get flag setting for user
*UserSettingsApi* | [**GetUserFlagSettings**](docs/UserSettingsApi.md#getuserflagsettings) | **Get** /api/v2/users/{projectKey}/{environmentKey}/{userKey}/flags | List flag settings for user
*UserSettingsApi* | [**PatchExpiringFlagsForUser**](docs/UserSettingsApi.md#patchexpiringflagsforuser) | **Patch** /api/v2/users/{projectKey}/{userKey}/expiring-user-targets/{environmentKey} | Update expiring user target for flags
*UserSettingsApi* | [**PutFlagSetting**](docs/UserSettingsApi.md#putflagsetting) | **Put** /api/v2/users/{projectKey}/{environmentKey}/{userKey}/flags/{featureFlagKey} | Update flag settings for user
*UsersApi* | [**DeleteUser**](docs/UsersApi.md#deleteuser) | **Delete** /api/v2/users/{projectKey}/{environmentKey}/{userKey} | Delete user
*UsersApi* | [**GetSearchUsers**](docs/UsersApi.md#getsearchusers) | **Get** /api/v2/user-search/{projectKey}/{environmentKey} | Find users
*UsersApi* | [**GetUser**](docs/UsersApi.md#getuser) | **Get** /api/v2/users/{projectKey}/{environmentKey}/{userKey} | Get user
*UsersApi* | [**GetUsers**](docs/UsersApi.md#getusers) | **Get** /api/v2/users/{projectKey}/{environmentKey} | List users
*UsersBetaApi* | [**GetUserAttributeNames**](docs/UsersBetaApi.md#getuserattributenames) | **Get** /api/v2/user-attributes/{projectKey}/{environmentKey} | Get user attribute names
*WebhooksApi* | [**DeleteWebhook**](docs/WebhooksApi.md#deletewebhook) | **Delete** /api/v2/webhooks/{id} | Delete webhook
*WebhooksApi* | [**GetAllWebhooks**](docs/WebhooksApi.md#getallwebhooks) | **Get** /api/v2/webhooks | List webhooks
*WebhooksApi* | [**GetWebhook**](docs/WebhooksApi.md#getwebhook) | **Get** /api/v2/webhooks/{id} | Get webhook
*WebhooksApi* | [**PatchWebhook**](docs/WebhooksApi.md#patchwebhook) | **Patch** /api/v2/webhooks/{id} | Update webhook
*WebhooksApi* | [**PostWebhook**](docs/WebhooksApi.md#postwebhook) | **Post** /api/v2/webhooks | Creates a webhook
*WorkflowTemplatesApi* | [**CreateWorkflowTemplate**](docs/WorkflowTemplatesApi.md#createworkflowtemplate) | **Post** /api/v2/templates | Create workflow template
*WorkflowTemplatesApi* | [**DeleteWorkflowTemplate**](docs/WorkflowTemplatesApi.md#deleteworkflowtemplate) | **Delete** /api/v2/templates/{templateKey} | Delete workflow template
*WorkflowTemplatesApi* | [**GetWorkflowTemplates**](docs/WorkflowTemplatesApi.md#getworkflowtemplates) | **Get** /api/v2/templates | Get workflow templates
*WorkflowsApi* | [**DeleteWorkflow**](docs/WorkflowsApi.md#deleteworkflow) | **Delete** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/workflows/{workflowId} | Delete workflow
*WorkflowsApi* | [**GetCustomWorkflow**](docs/WorkflowsApi.md#getcustomworkflow) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/workflows/{workflowId} | Get custom workflow
*WorkflowsApi* | [**GetWorkflows**](docs/WorkflowsApi.md#getworkflows) | **Get** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/workflows | Get workflows
*WorkflowsApi* | [**PostWorkflow**](docs/WorkflowsApi.md#postworkflow) | **Post** /api/v2/projects/{projectKey}/flags/{featureFlagKey}/environments/{environmentKey}/workflows | Create workflow


## Documentation For Models

 - [Access](docs/Access.md)
 - [AccessAllowedReason](docs/AccessAllowedReason.md)
 - [AccessAllowedRep](docs/AccessAllowedRep.md)
 - [AccessDenied](docs/AccessDenied.md)
 - [AccessDeniedReason](docs/AccessDeniedReason.md)
 - [AccessTokenPost](docs/AccessTokenPost.md)
 - [ActionInput](docs/ActionInput.md)
 - [ActionOutput](docs/ActionOutput.md)
 - [ApplicationCollectionRep](docs/ApplicationCollectionRep.md)
 - [ApplicationExpandableFields](docs/ApplicationExpandableFields.md)
 - [ApplicationFlagCollectionRep](docs/ApplicationFlagCollectionRep.md)
 - [ApplicationRep](docs/ApplicationRep.md)
 - [ApplicationVersionRep](docs/ApplicationVersionRep.md)
 - [ApplicationVersionsCollectionRep](docs/ApplicationVersionsCollectionRep.md)
 - [ApprovalConditionInput](docs/ApprovalConditionInput.md)
 - [ApprovalConditionOutput](docs/ApprovalConditionOutput.md)
 - [ApprovalRequestResponse](docs/ApprovalRequestResponse.md)
 - [ApprovalSettings](docs/ApprovalSettings.md)
 - [Audience](docs/Audience.md)
 - [AudiencePost](docs/AudiencePost.md)
 - [AuditLogEntryListingRep](docs/AuditLogEntryListingRep.md)
 - [AuditLogEntryListingRepCollection](docs/AuditLogEntryListingRepCollection.md)
 - [AuditLogEntryRep](docs/AuditLogEntryRep.md)
 - [AuthorizedAppDataRep](docs/AuthorizedAppDataRep.md)
 - [BigSegmentTarget](docs/BigSegmentTarget.md)
 - [BooleanDefaults](docs/BooleanDefaults.md)
 - [BooleanFlagDefaults](docs/BooleanFlagDefaults.md)
 - [BranchCollectionRep](docs/BranchCollectionRep.md)
 - [BranchRep](docs/BranchRep.md)
 - [BulkEditMembersRep](docs/BulkEditMembersRep.md)
 - [BulkEditTeamsRep](docs/BulkEditTeamsRep.md)
 - [Clause](docs/Clause.md)
 - [Client](docs/Client.md)
 - [ClientCollection](docs/ClientCollection.md)
 - [ClientSideAvailability](docs/ClientSideAvailability.md)
 - [ClientSideAvailabilityPost](docs/ClientSideAvailabilityPost.md)
 - [CompletedBy](docs/CompletedBy.md)
 - [ConditionBaseOutput](docs/ConditionBaseOutput.md)
 - [ConditionInput](docs/ConditionInput.md)
 - [ConditionOutput](docs/ConditionOutput.md)
 - [ConfidenceIntervalRep](docs/ConfidenceIntervalRep.md)
 - [Conflict](docs/Conflict.md)
 - [ConflictOutput](docs/ConflictOutput.md)
 - [ContextAttributeName](docs/ContextAttributeName.md)
 - [ContextAttributeNames](docs/ContextAttributeNames.md)
 - [ContextAttributeNamesCollection](docs/ContextAttributeNamesCollection.md)
 - [ContextAttributeValue](docs/ContextAttributeValue.md)
 - [ContextAttributeValues](docs/ContextAttributeValues.md)
 - [ContextAttributeValuesCollection](docs/ContextAttributeValuesCollection.md)
 - [ContextInstanceEvaluation](docs/ContextInstanceEvaluation.md)
 - [ContextInstanceEvaluationReason](docs/ContextInstanceEvaluationReason.md)
 - [ContextInstanceEvaluations](docs/ContextInstanceEvaluations.md)
 - [ContextInstanceRecord](docs/ContextInstanceRecord.md)
 - [ContextInstanceSearch](docs/ContextInstanceSearch.md)
 - [ContextInstanceSegmentMembership](docs/ContextInstanceSegmentMembership.md)
 - [ContextInstanceSegmentMemberships](docs/ContextInstanceSegmentMemberships.md)
 - [ContextInstances](docs/ContextInstances.md)
 - [ContextKind](docs/ContextKind.md)
 - [ContextKindRep](docs/ContextKindRep.md)
 - [ContextKindsCollectionRep](docs/ContextKindsCollectionRep.md)
 - [ContextRecord](docs/ContextRecord.md)
 - [ContextSearch](docs/ContextSearch.md)
 - [Contexts](docs/Contexts.md)
 - [CopiedFromEnv](docs/CopiedFromEnv.md)
 - [CreateApplicationInput](docs/CreateApplicationInput.md)
 - [CreateApplicationVersionInput](docs/CreateApplicationVersionInput.md)
 - [CreateApprovalRequestRequest](docs/CreateApprovalRequestRequest.md)
 - [CreateCopyFlagConfigApprovalRequestRequest](docs/CreateCopyFlagConfigApprovalRequestRequest.md)
 - [CreateFlagConfigApprovalRequestRequest](docs/CreateFlagConfigApprovalRequestRequest.md)
 - [CreatePhaseInput](docs/CreatePhaseInput.md)
 - [CreateReleasePipelineInput](docs/CreateReleasePipelineInput.md)
 - [CreateWorkflowTemplateInput](docs/CreateWorkflowTemplateInput.md)
 - [CredibleIntervalRep](docs/CredibleIntervalRep.md)
 - [CustomProperty](docs/CustomProperty.md)
 - [CustomRole](docs/CustomRole.md)
 - [CustomRolePost](docs/CustomRolePost.md)
 - [CustomRolePostData](docs/CustomRolePostData.md)
 - [CustomRoleSummary](docs/CustomRoleSummary.md)
 - [CustomRoles](docs/CustomRoles.md)
 - [CustomWorkflowInput](docs/CustomWorkflowInput.md)
 - [CustomWorkflowMeta](docs/CustomWorkflowMeta.md)
 - [CustomWorkflowOutput](docs/CustomWorkflowOutput.md)
 - [CustomWorkflowStageMeta](docs/CustomWorkflowStageMeta.md)
 - [CustomWorkflowsListingOutput](docs/CustomWorkflowsListingOutput.md)
 - [Database](docs/Database.md)
 - [DefaultClientSideAvailability](docs/DefaultClientSideAvailability.md)
 - [DefaultClientSideAvailabilityPost](docs/DefaultClientSideAvailabilityPost.md)
 - [Defaults](docs/Defaults.md)
 - [DependentExperimentRep](docs/DependentExperimentRep.md)
 - [DependentFlag](docs/DependentFlag.md)
 - [DependentFlagEnvironment](docs/DependentFlagEnvironment.md)
 - [DependentFlagsByEnvironment](docs/DependentFlagsByEnvironment.md)
 - [DependentMetricOrMetricGroupRep](docs/DependentMetricOrMetricGroupRep.md)
 - [DesignExpandableProperties](docs/DesignExpandableProperties.md)
 - [DesignRep](docs/DesignRep.md)
 - [Destination](docs/Destination.md)
 - [DestinationPost](docs/DestinationPost.md)
 - [Destinations](docs/Destinations.md)
 - [Distribution](docs/Distribution.md)
 - [Environment](docs/Environment.md)
 - [EnvironmentPost](docs/EnvironmentPost.md)
 - [EnvironmentSummary](docs/EnvironmentSummary.md)
 - [Environments](docs/Environments.md)
 - [EvaluationReason](docs/EvaluationReason.md)
 - [ExecutionOutput](docs/ExecutionOutput.md)
 - [ExpandableApprovalRequestResponse](docs/ExpandableApprovalRequestResponse.md)
 - [ExpandableApprovalRequestsResponse](docs/ExpandableApprovalRequestsResponse.md)
 - [ExpandedFlagRep](docs/ExpandedFlagRep.md)
 - [Experiment](docs/Experiment.md)
 - [ExperimentAllocationRep](docs/ExperimentAllocationRep.md)
 - [ExperimentBayesianResultsRep](docs/ExperimentBayesianResultsRep.md)
 - [ExperimentCollectionRep](docs/ExperimentCollectionRep.md)
 - [ExperimentEnabledPeriodRep](docs/ExperimentEnabledPeriodRep.md)
 - [ExperimentEnvironmentSettingRep](docs/ExperimentEnvironmentSettingRep.md)
 - [ExperimentExpandableProperties](docs/ExperimentExpandableProperties.md)
 - [ExperimentInfoRep](docs/ExperimentInfoRep.md)
 - [ExperimentMetadataRep](docs/ExperimentMetadataRep.md)
 - [ExperimentPatchInput](docs/ExperimentPatchInput.md)
 - [ExperimentPost](docs/ExperimentPost.md)
 - [ExperimentResults](docs/ExperimentResults.md)
 - [ExperimentStatsRep](docs/ExperimentStatsRep.md)
 - [ExperimentTimeSeriesSlice](docs/ExperimentTimeSeriesSlice.md)
 - [ExperimentTimeSeriesVariationSlice](docs/ExperimentTimeSeriesVariationSlice.md)
 - [ExperimentTotalsRep](docs/ExperimentTotalsRep.md)
 - [ExperimentationSettingsPut](docs/ExperimentationSettingsPut.md)
 - [ExperimentationSettingsRep](docs/ExperimentationSettingsRep.md)
 - [ExpiringTarget](docs/ExpiringTarget.md)
 - [ExpiringTargetError](docs/ExpiringTargetError.md)
 - [ExpiringTargetGetResponse](docs/ExpiringTargetGetResponse.md)
 - [ExpiringTargetPatchResponse](docs/ExpiringTargetPatchResponse.md)
 - [ExpiringUserTargetGetResponse](docs/ExpiringUserTargetGetResponse.md)
 - [ExpiringUserTargetItem](docs/ExpiringUserTargetItem.md)
 - [ExpiringUserTargetPatchResponse](docs/ExpiringUserTargetPatchResponse.md)
 - [Export](docs/Export.md)
 - [Extinction](docs/Extinction.md)
 - [ExtinctionCollectionRep](docs/ExtinctionCollectionRep.md)
 - [FeatureFlag](docs/FeatureFlag.md)
 - [FeatureFlagBody](docs/FeatureFlagBody.md)
 - [FeatureFlagConfig](docs/FeatureFlagConfig.md)
 - [FeatureFlagScheduledChange](docs/FeatureFlagScheduledChange.md)
 - [FeatureFlagScheduledChanges](docs/FeatureFlagScheduledChanges.md)
 - [FeatureFlagStatus](docs/FeatureFlagStatus.md)
 - [FeatureFlagStatusAcrossEnvironments](docs/FeatureFlagStatusAcrossEnvironments.md)
 - [FeatureFlagStatuses](docs/FeatureFlagStatuses.md)
 - [FeatureFlags](docs/FeatureFlags.md)
 - [FileRep](docs/FileRep.md)
 - [FlagConfigApprovalRequestResponse](docs/FlagConfigApprovalRequestResponse.md)
 - [FlagConfigApprovalRequestsResponse](docs/FlagConfigApprovalRequestsResponse.md)
 - [FlagConfigEvaluation](docs/FlagConfigEvaluation.md)
 - [FlagConfigMigrationSettingsRep](docs/FlagConfigMigrationSettingsRep.md)
 - [FlagCopyConfigEnvironment](docs/FlagCopyConfigEnvironment.md)
 - [FlagCopyConfigPost](docs/FlagCopyConfigPost.md)
 - [FlagDefaults](docs/FlagDefaults.md)
 - [FlagDefaultsApiBaseRep](docs/FlagDefaultsApiBaseRep.md)
 - [FlagDefaultsRep](docs/FlagDefaultsRep.md)
 - [FlagFollowersByProjEnvGetRep](docs/FlagFollowersByProjEnvGetRep.md)
 - [FlagFollowersGetRep](docs/FlagFollowersGetRep.md)
 - [FlagGlobalAttributesRep](docs/FlagGlobalAttributesRep.md)
 - [FlagInput](docs/FlagInput.md)
 - [FlagLinkCollectionRep](docs/FlagLinkCollectionRep.md)
 - [FlagLinkMember](docs/FlagLinkMember.md)
 - [FlagLinkPost](docs/FlagLinkPost.md)
 - [FlagLinkRep](docs/FlagLinkRep.md)
 - [FlagListingRep](docs/FlagListingRep.md)
 - [FlagMigrationSettingsRep](docs/FlagMigrationSettingsRep.md)
 - [FlagRep](docs/FlagRep.md)
 - [FlagScheduledChangesInput](docs/FlagScheduledChangesInput.md)
 - [FlagSempatch](docs/FlagSempatch.md)
 - [FlagStatusRep](docs/FlagStatusRep.md)
 - [FlagSummary](docs/FlagSummary.md)
 - [FlagTriggerInput](docs/FlagTriggerInput.md)
 - [FollowFlagMember](docs/FollowFlagMember.md)
 - [FollowersPerFlag](docs/FollowersPerFlag.md)
 - [ForbiddenErrorRep](docs/ForbiddenErrorRep.md)
 - [HunkRep](docs/HunkRep.md)
 - [Import](docs/Import.md)
 - [InitiatorRep](docs/InitiatorRep.md)
 - [InstructionUserRequest](docs/InstructionUserRequest.md)
 - [Integration](docs/Integration.md)
 - [IntegrationDeliveryConfiguration](docs/IntegrationDeliveryConfiguration.md)
 - [IntegrationDeliveryConfigurationCollection](docs/IntegrationDeliveryConfigurationCollection.md)
 - [IntegrationDeliveryConfigurationCollectionLinks](docs/IntegrationDeliveryConfigurationCollectionLinks.md)
 - [IntegrationDeliveryConfigurationLinks](docs/IntegrationDeliveryConfigurationLinks.md)
 - [IntegrationDeliveryConfigurationPost](docs/IntegrationDeliveryConfigurationPost.md)
 - [IntegrationDeliveryConfigurationResponse](docs/IntegrationDeliveryConfigurationResponse.md)
 - [IntegrationMetadata](docs/IntegrationMetadata.md)
 - [IntegrationStatus](docs/IntegrationStatus.md)
 - [IntegrationStatusRep](docs/IntegrationStatusRep.md)
 - [IntegrationSubscriptionStatusRep](docs/IntegrationSubscriptionStatusRep.md)
 - [Integrations](docs/Integrations.md)
 - [InvalidRequestErrorRep](docs/InvalidRequestErrorRep.md)
 - [IpList](docs/IpList.md)
 - [IterationInput](docs/IterationInput.md)
 - [IterationRep](docs/IterationRep.md)
 - [LastSeenMetadata](docs/LastSeenMetadata.md)
 - [LegacyExperimentRep](docs/LegacyExperimentRep.md)
 - [Link](docs/Link.md)
 - [MaintainerInput](docs/MaintainerInput.md)
 - [MaintainerRep](docs/MaintainerRep.md)
 - [MaintainerTeam](docs/MaintainerTeam.md)
 - [Member](docs/Member.md)
 - [MemberDataRep](docs/MemberDataRep.md)
 - [MemberImportItem](docs/MemberImportItem.md)
 - [MemberInput](docs/MemberInput.md)
 - [MemberPermissionGrantSummaryRep](docs/MemberPermissionGrantSummaryRep.md)
 - [MemberSummary](docs/MemberSummary.md)
 - [MemberTeamSummaryRep](docs/MemberTeamSummaryRep.md)
 - [MemberTeamsPostInput](docs/MemberTeamsPostInput.md)
 - [Members](docs/Members.md)
 - [MembersPatchInput](docs/MembersPatchInput.md)
 - [MethodNotAllowedErrorRep](docs/MethodNotAllowedErrorRep.md)
 - [MetricCollectionRep](docs/MetricCollectionRep.md)
 - [MetricEventDefaultRep](docs/MetricEventDefaultRep.md)
 - [MetricGroupCollectionRep](docs/MetricGroupCollectionRep.md)
 - [MetricGroupPost](docs/MetricGroupPost.md)
 - [MetricGroupRep](docs/MetricGroupRep.md)
 - [MetricGroupRepExpandableProperties](docs/MetricGroupRepExpandableProperties.md)
 - [MetricGroupResultsRep](docs/MetricGroupResultsRep.md)
 - [MetricInGroupRep](docs/MetricInGroupRep.md)
 - [MetricInGroupResultsRep](docs/MetricInGroupResultsRep.md)
 - [MetricInMetricGroupInput](docs/MetricInMetricGroupInput.md)
 - [MetricInput](docs/MetricInput.md)
 - [MetricListingRep](docs/MetricListingRep.md)
 - [MetricListingRepExpandableProperties](docs/MetricListingRepExpandableProperties.md)
 - [MetricPost](docs/MetricPost.md)
 - [MetricRep](docs/MetricRep.md)
 - [MetricRepExpandableProperties](docs/MetricRepExpandableProperties.md)
 - [MetricSeen](docs/MetricSeen.md)
 - [MetricV2Rep](docs/MetricV2Rep.md)
 - [MigrationSafetyIssueRep](docs/MigrationSafetyIssueRep.md)
 - [MigrationSettingsPost](docs/MigrationSettingsPost.md)
 - [Modification](docs/Modification.md)
 - [MultiEnvironmentDependentFlag](docs/MultiEnvironmentDependentFlag.md)
 - [MultiEnvironmentDependentFlags](docs/MultiEnvironmentDependentFlags.md)
 - [NewMemberForm](docs/NewMemberForm.md)
 - [NotFoundErrorRep](docs/NotFoundErrorRep.md)
 - [OauthClientPost](docs/OauthClientPost.md)
 - [ParameterDefault](docs/ParameterDefault.md)
 - [ParameterDefaultInput](docs/ParameterDefaultInput.md)
 - [ParameterRep](docs/ParameterRep.md)
 - [ParentResourceRep](docs/ParentResourceRep.md)
 - [PatchFailedErrorRep](docs/PatchFailedErrorRep.md)
 - [PatchFlagsRequest](docs/PatchFlagsRequest.md)
 - [PatchOperation](docs/PatchOperation.md)
 - [PatchSegmentExpiringTargetInputRep](docs/PatchSegmentExpiringTargetInputRep.md)
 - [PatchSegmentExpiringTargetInstruction](docs/PatchSegmentExpiringTargetInstruction.md)
 - [PatchSegmentInstruction](docs/PatchSegmentInstruction.md)
 - [PatchSegmentRequest](docs/PatchSegmentRequest.md)
 - [PatchUsersRequest](docs/PatchUsersRequest.md)
 - [PatchWithComment](docs/PatchWithComment.md)
 - [PermissionGrantInput](docs/PermissionGrantInput.md)
 - [Phase](docs/Phase.md)
 - [PostApprovalRequestApplyRequest](docs/PostApprovalRequestApplyRequest.md)
 - [PostApprovalRequestReviewRequest](docs/PostApprovalRequestReviewRequest.md)
 - [PostFlagScheduledChangesInput](docs/PostFlagScheduledChangesInput.md)
 - [Prerequisite](docs/Prerequisite.md)
 - [Project](docs/Project.md)
 - [ProjectListingRep](docs/ProjectListingRep.md)
 - [ProjectPost](docs/ProjectPost.md)
 - [ProjectRep](docs/ProjectRep.md)
 - [ProjectSummary](docs/ProjectSummary.md)
 - [Projects](docs/Projects.md)
 - [PubNubDetailRep](docs/PubNubDetailRep.md)
 - [PutBranch](docs/PutBranch.md)
 - [RandomizationUnitInput](docs/RandomizationUnitInput.md)
 - [RandomizationUnitRep](docs/RandomizationUnitRep.md)
 - [RateLimitedErrorRep](docs/RateLimitedErrorRep.md)
 - [RecentTriggerBody](docs/RecentTriggerBody.md)
 - [ReferenceRep](docs/ReferenceRep.md)
 - [RelativeDifferenceRep](docs/RelativeDifferenceRep.md)
 - [RelayAutoConfigCollectionRep](docs/RelayAutoConfigCollectionRep.md)
 - [RelayAutoConfigPost](docs/RelayAutoConfigPost.md)
 - [RelayAutoConfigRep](docs/RelayAutoConfigRep.md)
 - [Release](docs/Release.md)
 - [ReleasePhase](docs/ReleasePhase.md)
 - [ReleasePipeline](docs/ReleasePipeline.md)
 - [ReleasePipelineCollection](docs/ReleasePipelineCollection.md)
 - [RepositoryCollectionRep](docs/RepositoryCollectionRep.md)
 - [RepositoryPost](docs/RepositoryPost.md)
 - [RepositoryRep](docs/RepositoryRep.md)
 - [ResolvedContext](docs/ResolvedContext.md)
 - [ResolvedImage](docs/ResolvedImage.md)
 - [ResolvedTitle](docs/ResolvedTitle.md)
 - [ResolvedUIBlockElement](docs/ResolvedUIBlockElement.md)
 - [ResolvedUIBlocks](docs/ResolvedUIBlocks.md)
 - [ResourceAccess](docs/ResourceAccess.md)
 - [ResourceIDResponse](docs/ResourceIDResponse.md)
 - [ResourceId](docs/ResourceId.md)
 - [ReviewOutput](docs/ReviewOutput.md)
 - [ReviewResponse](docs/ReviewResponse.md)
 - [Rollout](docs/Rollout.md)
 - [RootResponse](docs/RootResponse.md)
 - [Rule](docs/Rule.md)
 - [RuleClause](docs/RuleClause.md)
 - [ScheduleConditionInput](docs/ScheduleConditionInput.md)
 - [ScheduleConditionOutput](docs/ScheduleConditionOutput.md)
 - [SdkListRep](docs/SdkListRep.md)
 - [SdkVersionListRep](docs/SdkVersionListRep.md)
 - [SdkVersionRep](docs/SdkVersionRep.md)
 - [SegmentBody](docs/SegmentBody.md)
 - [SegmentMetadata](docs/SegmentMetadata.md)
 - [SegmentTarget](docs/SegmentTarget.md)
 - [SegmentUserList](docs/SegmentUserList.md)
 - [SegmentUserState](docs/SegmentUserState.md)
 - [Series](docs/Series.md)
 - [SeriesIntervalsRep](docs/SeriesIntervalsRep.md)
 - [SeriesListRep](docs/SeriesListRep.md)
 - [SlicedResultsRep](docs/SlicedResultsRep.md)
 - [SourceEnv](docs/SourceEnv.md)
 - [SourceFlag](docs/SourceFlag.md)
 - [StageInput](docs/StageInput.md)
 - [StageOutput](docs/StageOutput.md)
 - [Statement](docs/Statement.md)
 - [StatementPost](docs/StatementPost.md)
 - [StatementPostData](docs/StatementPostData.md)
 - [StatisticCollectionRep](docs/StatisticCollectionRep.md)
 - [StatisticRep](docs/StatisticRep.md)
 - [StatisticsRep](docs/StatisticsRep.md)
 - [StatisticsRoot](docs/StatisticsRoot.md)
 - [StatusConflictErrorRep](docs/StatusConflictErrorRep.md)
 - [StatusServiceUnavailable](docs/StatusServiceUnavailable.md)
 - [SubjectDataRep](docs/SubjectDataRep.md)
 - [SubscriptionPost](docs/SubscriptionPost.md)
 - [TagCollection](docs/TagCollection.md)
 - [Target](docs/Target.md)
 - [TargetResourceRep](docs/TargetResourceRep.md)
 - [Team](docs/Team.md)
 - [TeamCustomRole](docs/TeamCustomRole.md)
 - [TeamCustomRoles](docs/TeamCustomRoles.md)
 - [TeamImportsRep](docs/TeamImportsRep.md)
 - [TeamInput](docs/TeamInput.md)
 - [TeamMaintainers](docs/TeamMaintainers.md)
 - [TeamMembers](docs/TeamMembers.md)
 - [TeamPatchInput](docs/TeamPatchInput.md)
 - [TeamPostInput](docs/TeamPostInput.md)
 - [TeamProjects](docs/TeamProjects.md)
 - [TeamRepExpandableProperties](docs/TeamRepExpandableProperties.md)
 - [Teams](docs/Teams.md)
 - [TeamsPatchInput](docs/TeamsPatchInput.md)
 - [TimestampRep](docs/TimestampRep.md)
 - [TitleRep](docs/TitleRep.md)
 - [Token](docs/Token.md)
 - [TokenSummary](docs/TokenSummary.md)
 - [Tokens](docs/Tokens.md)
 - [TreatmentInput](docs/TreatmentInput.md)
 - [TreatmentParameterInput](docs/TreatmentParameterInput.md)
 - [TreatmentRep](docs/TreatmentRep.md)
 - [TreatmentResultRep](docs/TreatmentResultRep.md)
 - [TriggerPost](docs/TriggerPost.md)
 - [TriggerWorkflowCollectionRep](docs/TriggerWorkflowCollectionRep.md)
 - [TriggerWorkflowRep](docs/TriggerWorkflowRep.md)
 - [UnauthorizedErrorRep](docs/UnauthorizedErrorRep.md)
 - [UpsertContextKindPayload](docs/UpsertContextKindPayload.md)
 - [UpsertFlagDefaultsPayload](docs/UpsertFlagDefaultsPayload.md)
 - [UpsertPayloadRep](docs/UpsertPayloadRep.md)
 - [UpsertResponseRep](docs/UpsertResponseRep.md)
 - [UrlPost](docs/UrlPost.md)
 - [User](docs/User.md)
 - [UserAttributeNamesRep](docs/UserAttributeNamesRep.md)
 - [UserFlagSetting](docs/UserFlagSetting.md)
 - [UserFlagSettings](docs/UserFlagSettings.md)
 - [UserRecord](docs/UserRecord.md)
 - [UserRecordRep](docs/UserRecordRep.md)
 - [UserSegment](docs/UserSegment.md)
 - [UserSegmentRule](docs/UserSegmentRule.md)
 - [UserSegments](docs/UserSegments.md)
 - [Users](docs/Users.md)
 - [UsersRep](docs/UsersRep.md)
 - [ValuePut](docs/ValuePut.md)
 - [Variate](docs/Variate.md)
 - [Variation](docs/Variation.md)
 - [VariationOrRolloutRep](docs/VariationOrRolloutRep.md)
 - [VariationSummary](docs/VariationSummary.md)
 - [VersionsRep](docs/VersionsRep.md)
 - [Webhook](docs/Webhook.md)
 - [WebhookPost](docs/WebhookPost.md)
 - [Webhooks](docs/Webhooks.md)
 - [WeightedVariation](docs/WeightedVariation.md)
 - [WorkflowTemplateMetadata](docs/WorkflowTemplateMetadata.md)
 - [WorkflowTemplateOutput](docs/WorkflowTemplateOutput.md)
 - [WorkflowTemplateParameter](docs/WorkflowTemplateParameter.md)
 - [WorkflowTemplateParameterInput](docs/WorkflowTemplateParameterInput.md)
 - [WorkflowTemplatesListingOutputRep](docs/WorkflowTemplatesListingOutputRep.md)


## Documentation For Authorization



### ApiKey

- **Type**: API key
- **API key parameter name**: Authorization
- **Location**: HTTP header

Note, each API key must be added to a map of `map[string]APIKey` where the key is: Authorization and passed in as the auth context for each request.


## Documentation for Utility Methods

Due to the fact that model structure members are all pointers, this package contains
a number of utility functions to easily obtain pointers to values of basic types.
Each of these functions takes a value of the given basic type and returns a pointer to it:

* `PtrBool`
* `PtrInt`
* `PtrInt32`
* `PtrInt64`
* `PtrFloat`
* `PtrFloat32`
* `PtrFloat64`
* `PtrString`
* `PtrTime`

## Author

support@launchdarkly.com

## Sample Code

```go
package main

import (
	"context"
	"fmt"
	"os"

	ldapi "github.com/launchdarkly/api-client-go"
)

func main() {
	apiKey := os.Getenv("LD_API_KEY")
	if apiKey == "" {
		panic("LD_API_KEY env var was empty!")
	}
	client := ldapi.NewAPIClient(ldapi.NewConfiguration())

	auth := make(map[string]ldapi.APIKey)
	auth["ApiKey"] = ldapi.APIKey{
		Key: apiKey,
	}

	ctx := context.WithValue(context.Background(), ldapi.ContextAPIKeys, auth)

	flagName := "Test Flag Go"
	flagKey := "test-go"
	// Create a multi-variate feature flag
	valOneVal := []int{1, 2}
	valOne := map[string]interface{}{"one": valOneVal}
	valTwoVal := []int{4, 5}
	valTwo := map[string]interface{}{"two": valTwoVal}

	body := ldapi.FeatureFlagBody{
		Name: flagName,
		Key:  flagKey,
		Variations: []ldapi.Variation{
			{Value: &valOne},
			{Value: &valTwo},
		},
	}
	flag, resp, err := client.FeatureFlagsApi.PostFeatureFlag(ctx, "openapi").FeatureFlagBody(body).Execute()
	if err != nil {
		if resp.StatusCode != 409 {
			panic(fmt.Errorf("create failed: %s", err))
		} else {
			if _, err := client.FeatureFlagsApi.DeleteFeatureFlag(ctx, "openapi", body.Key).Execute(); err != nil {
				panic(fmt.Errorf("delete failed: %s", err))
			}
			flag, resp, err = client.FeatureFlagsApi.PostFeatureFlag(ctx, "openapi").FeatureFlagBody(body).Execute()
			if err != nil {
				panic(fmt.Errorf("create failed: %s", err))
			}
		}
	}
	fmt.Printf("Created flag: %+v\n", flag)
	// Clean up new flag
	defer func() {
		if _, err := client.FeatureFlagsApi.DeleteFeatureFlag(ctx, "openapi", body.Key).Execute(); err != nil {
			panic(fmt.Errorf("delete failed: %s", err))
		}
	}()
}

func intfPtr(i interface{}) *interface{} {
	return &i
}
```
