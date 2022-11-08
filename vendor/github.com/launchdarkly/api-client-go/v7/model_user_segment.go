/*
LaunchDarkly REST API

# Overview  ## Authentication  All REST API resources are authenticated with either [personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens), or session cookies. Other authentication mechanisms are not supported. You can manage personal access tokens on your [Account settings](https://app.launchdarkly.com/settings/tokens) page.  LaunchDarkly also has SDK keys, mobile keys, and client-side IDs that are used by our server-side SDKs, mobile SDKs, and client-side SDKs, respectively. **These keys cannot be used to access our REST API**. These keys are environment-specific, and can only perform read-only operations (fetching feature flag settings).  | Auth mechanism                                                                                  | Allowed resources                                                                                     | Use cases                                          | | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- | | [Personal access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens) | Can be customized on a per-token basis                                                                | Building scripts, custom integrations, data export | | SDK keys                                                                                        | Can only access read-only SDK-specific resources and the firehose, restricted to a single environment | Server-side SDKs, Firehose API                     | | Mobile keys                                                                                     | Can only access read-only mobile SDK-specific resources, restricted to a single environment           | Mobile SDKs                                        | | Client-side ID                                                                                  | Single environment, only flags marked available to client-side                                        | Client-side JavaScript                             |  > #### Keep your access tokens and SDK keys private > > Access tokens should _never_ be exposed in untrusted contexts. Never put an access token in client-side JavaScript, or embed it in a mobile application. LaunchDarkly has special mobile keys that you can embed in mobile apps. If you accidentally expose an access token or SDK key, you can reset it from your [Account Settings](https://app.launchdarkly.com/settings#/tokens) page. > > The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.  ### Via request header  The preferred way to authenticate with the API is by adding an `Authorization` header containing your access token to your requests. The value of the `Authorization` header must be your access token.  Manage personal access tokens from the [Account Settings](https://app.launchdarkly.com/settings/tokens) page.  ### Via session cookie  For testing purposes, you can make API calls directly from your web browser. If you're logged in to the application, the API will use your existing session to authenticate calls.  If you have a [role](https://docs.launchdarkly.com/home/team/built-in-roles) other than Admin, or have a [custom role](https://docs.launchdarkly.com/home/team/custom-roles) defined, you may not have permission to perform some API calls. You will receive a `401` response code in that case.  > ### Modifying the Origin header causes an error > > LaunchDarkly validates that the Origin header for any API request authenticated by a session cookie matches the expected Origin header. The expected Origin header is `https://app.launchdarkly.com`. > > If the Origin header does not match what's expected, LaunchDarkly returns an error. This error can prevent the LaunchDarkly app from working correctly. > > Any browser extension that intentionally changes the Origin header can cause this problem. For example, the `Allow-Control-Allow-Origin: *` Chrome extension changes the Origin header to `http://evil.com` and causes the app to fail. > > To prevent this error, do not modify your Origin header. > > LaunchDarkly does not require origin matching when authenticating with an access token, so this issue does not affect normal API usage.  ## Representations  All resources expect and return JSON response bodies. Error responses will also send a JSON body. Read [Errors](#section/Errors) for a more detailed description of the error format used by the API.  In practice this means that you always get a response with a `Content-Type` header set to `application/json`.  In addition, request bodies for `PUT`, `POST`, `REPORT` and `PATCH` requests must be encoded as JSON with a `Content-Type` header set to `application/json`.  ### Summary and detailed representations  When you fetch a list of resources, the response includes only the most important attributes of each resource. This is a _summary representation_ of the resource. When you fetch an individual resource (for example, a single feature flag), you receive a _detailed representation_ containing all of the attributes of the resource.  The best way to find a detailed representation is to follow links. Every summary representation includes a link to its detailed representation.  ### Links and addressability  The best way to navigate the API is by following links. These are attributes in representations that link to other resources. The API always uses the same format for links:  - Links to other resources within the API are encapsulated in a `_links` object. - If the resource has a corresponding link to HTML content on the site, it is stored in a special `_site` link.  Each link has two attributes: an href (the URL) and a type (the content type). For example, a feature resource might return the following:  ```json {   \"_links\": {     \"parent\": {       \"href\": \"/api/features\",       \"type\": \"application/json\"     },     \"self\": {       \"href\": \"/api/features/sort.order\",       \"type\": \"application/json\"     }   },   \"_site\": {     \"href\": \"/features/sort.order\",     \"type\": \"text/html\"   } } ```  From this, you can navigate to the parent collection of features by following the `parent` link, or navigate to the site page for the feature by following the `_site` link.  Collections are always represented as a JSON object with an `items` attribute containing an array of representations. Like all other representations, collections have `_links` defined at the top level.  Paginated collections include `first`, `last`, `next`, and `prev` links containing a URL with the respective set of elements in the collection.  ## Updates  Resources that accept partial updates use the `PATCH` verb, and support the [JSON Patch](https://datatracker.ietf.org/doc/html/rfc6902) format. Some resources also support the [JSON Merge Patch](https://datatracker.ietf.org/doc/html/rfc7386) format. In addition, some resources support optional comments that can be submitted with updates. Comments appear in outgoing webhooks, the audit log, and other integrations.  ### Updates via JSON Patch  [JSON Patch](https://datatracker.ietf.org/doc/html/rfc6902) is a way to specify the modifications to perform on a resource. For example, in this feature flag representation:  ```json {     \"name\": \"New recommendations engine\",     \"key\": \"engine.enable\",     \"description\": \"This is the description\",     ... } ```  You can change the feature flag's description with the following patch document:  ```json [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"This is the new description\" }] ```  JSON Patch documents are always arrays. You can specify multiple modifications to perform in a single request. You can also test that certain preconditions are met before applying the patch:  ```json [   { \"op\": \"test\", \"path\": \"/version\", \"value\": 10 },   { \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" } ] ```  The above patch request tests whether the feature flag's `version` is `10`, and if so, changes the feature flag's description.  Attributes that aren't editable, like a resource's `_links`, have names that start with an underscore.  ### Updates via JSON Merge Patch  The API also supports the [JSON Merge Patch](https://datatracker.ietf.org/doc/html/rfc7386) format, as well as the [Update feature flag](/tag/Feature-flags#operation/patchFeatureFlag) resource.  JSON Merge Patch is less expressive than JSON Patch but in many cases, it is simpler to construct a merge patch document. For example, you can change a feature flag's description with the following merge patch document:  ```json {   \"description\": \"New flag description\" } ```  ### Updates with comments  You can submit optional comments with `PATCH` changes. The [Update feature flag](/tag/Feature-flags#operation/patchFeatureFlag) resource supports comments.  To submit a comment along with a JSON Patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"patch\": [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }] } ```  To submit a comment along with a JSON Merge Patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"merge\": { \"description\": \"New flag description\" } } ```  ### Updates via semantic patches  The API also supports the Semantic patch format. A semantic `PATCH` is a way to specify the modifications to perform on a resource as a set of executable instructions.  JSON Patch uses paths and a limited set of operations to describe how to transform the current state of the resource into a new state. Semantic patch allows you to be explicit about intent using precise, custom instructions. In many cases, semantic patch instructions can also be defined independently of the current state of the resource. This can be useful when defining a change that may be applied at a future date.  For example, in this feature flag configuration in environment Production:  ```json {     \"name\": \"Alternate sort order\",     \"kind\": \"boolean\",     \"key\": \"sort.order\",    ...     \"environments\": {         \"production\": {             \"on\": true,             \"archived\": false,             \"salt\": \"c29ydC5vcmRlcg==\",             \"sel\": \"8de1085cb7354b0ab41c0e778376dfd3\",             \"lastModified\": 1469131558260,             \"version\": 81,             \"targets\": [                 {                     \"values\": [                         \"Gerhard.Little@yahoo.com\"                     ],                     \"variation\": 0                 },                 {                     \"values\": [                         \"1461797806429-33-861961230\",                         \"438580d8-02ee-418d-9eec-0085cab2bdf0\"                     ],                     \"variation\": 1                 }             ],             \"rules\": [],             \"fallthrough\": {                 \"variation\": 0             },             \"offVariation\": 1,             \"prerequisites\": [],             \"_site\": {                 \"href\": \"/default/production/features/sort.order\",                 \"type\": \"text/html\"             }        }     } } ```  You can add a date you want a user to be removed from the feature flag's user targets. For example, “remove user 1461797806429-33-861961230 from the user target for variation 0 on the Alternate sort order flag in the production environment on Wed Jul 08 2020 at 15:27:41 pm”. This is done using the following:  ```json {   \"comment\": \"update expiring user targets\",   \"instructions\": [     {       \"kind\": \"removeExpireUserTargetDate\",       \"userKey\": \"userKey\",       \"variationId\": \"978d53f9-7fe3-4a63-992d-97bcb4535dc8\"     },     {       \"kind\": \"updateExpireUserTargetDate\",       \"userKey\": \"userKey2\",       \"variationId\": \"978d53f9-7fe3-4a63-992d-97bcb4535dc8\",       \"value\": 1587582000000     },     {       \"kind\": \"addExpireUserTargetDate\",       \"userKey\": \"userKey3\",       \"variationId\": \"978d53f9-7fe3-4a63-992d-97bcb4535dc8\",       \"value\": 1594247266386     }   ] } ```  Here is another example. In this feature flag configuration:  ```json {   \"name\": \"New recommendations engine\",   \"key\": \"engine.enable\",   \"environments\": {     \"test\": {       \"on\": true     }   } } ```  You can change the feature flag's description with the following patch document as a set of executable instructions. For example, “add user X to targets for variation Y and remove user A from targets for variation B for test flag”:  ```json {   \"comment\": \"\",   \"instructions\": [     {       \"kind\": \"removeUserTargets\",       \"values\": [\"438580d8-02ee-418d-9eec-0085cab2bdf0\"],       \"variationId\": \"852cb784-54ff-46b9-8c35-5498d2e4f270\"     },     {       \"kind\": \"addUserTargets\",       \"values\": [\"438580d8-02ee-418d-9eec-0085cab2bdf0\"],       \"variationId\": \"1bb18465-33b6-49aa-a3bd-eeb6650b33ad\"     }   ] } ```  > ### Supported semantic patch API endpoints > > - [Update feature flag](/tag/Feature-flags#operation/patchFeatureFlag) > - [Update expiring user targets on feature flag](/tag/Feature-flags#operation/patchExpiringUserTargets) > - [Update expiring user target for flags](/tag/User-settings#operation/patchExpiringFlagsForUser) > - [Update expiring user targets on segment](/tag/Segments#operation/patchExpiringUserTargetsForSegment)  ## Errors  The API always returns errors in a common format. Here's an example:  ```json {   \"code\": \"invalid_request\",   \"message\": \"A feature with that key already exists\",   \"id\": \"30ce6058-87da-11e4-b116-123b93f75cba\" } ```  The general class of error is indicated by the `code`. The `message` is a human-readable explanation of what went wrong. The `id` is a unique identifier. Use it when you're working with LaunchDarkly support to debug a problem with a specific API call.  ### HTTP Status - Error Response Codes  | Code | Definition        | Desc.                                                                                       | Possible Solution                                                | | ---- | ----------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | | 400  | Bad Request       | A request that fails may return this HTTP response code.                                    | Ensure JSON syntax in request body is correct.                   | | 401  | Unauthorized      | User doesn't have permission to an API call.                                                | Ensure your SDK key is good.                                     | | 403  | Forbidden         | User does not have permission for operation.                                                | Ensure that the user or access token has proper permissions set. | | 409  | Conflict          | The API request could not be completed because it conflicted with a concurrent API request. | Retry your request.                                              | | 429  | Too many requests | See [Rate limiting](/#section/Rate-limiting).                                               | Wait and try again later.                                        |  ## CORS  The LaunchDarkly API supports Cross Origin Resource Sharing (CORS) for AJAX requests from any origin. If an `Origin` header is given in a request, it will be echoed as an explicitly allowed origin. Otherwise, a wildcard is returned: `Access-Control-Allow-Origin: *`. For more information on CORS, see the [CORS W3C Recommendation](http://www.w3.org/TR/cors). Example CORS headers might look like:  ```http Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization Access-Control-Allow-Methods: OPTIONS, GET, DELETE, PATCH Access-Control-Allow-Origin: * Access-Control-Max-Age: 300 ```  You can make authenticated CORS calls just as you would make same-origin calls, using either [token or session-based authentication](#section/Authentication). If you’re using session auth, you should set the `withCredentials` property for your `xhr` request to `true`. You should never expose your access tokens to untrusted users.  ## Rate limiting  We use several rate limiting strategies to ensure the availability of our APIs. Rate-limited calls to our APIs will return a `429` status code. Calls to our APIs will include headers indicating the current rate limit status. The specific headers returned depend on the API route being called. The limits differ based on the route, authentication mechanism, and other factors. Routes that are not rate limited may not contain any of the headers described below.  > ### Rate limiting and SDKs > > LaunchDarkly SDKs are never rate limited and do not use the API endpoints defined here. LaunchDarkly uses a different set of approaches, including streaming/server-sent events and a global CDN, to ensure availability to the routes used by LaunchDarkly SDKs. > > The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.  ### Global rate limits  Authenticated requests are subject to a global limit. This is the maximum number of calls that can be made to the API per ten seconds. All personal access tokens on the account share this limit, so exceeding the limit with one access token will impact other tokens. Calls that are subject to global rate limits will return the headers below:  | Header name                    | Description                                                                      | | ------------------------------ | -------------------------------------------------------------------------------- | | `X-Ratelimit-Global-Remaining` | The maximum number of requests the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`            | The time at which the current rate limit window resets in epoch milliseconds.    |  We do not publicly document the specific number of calls that can be made globally. This limit may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limit.  ### Route-level rate limits  Some authenticated routes have custom rate limits. These also reset every ten seconds. Any access tokens hitting the same route share this limit, so exceeding the limit with one access token may impact other tokens. Calls that are subject to route-level rate limits will return the headers below:  | Header name                   | Description                                                                                           | | ----------------------------- | ----------------------------------------------------------------------------------------------------- | | `X-Ratelimit-Route-Remaining` | The maximum number of requests to the current route the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`           | The time at which the current rate limit window resets in epoch milliseconds.                         |  A _route_ represents a specific URL pattern and verb. For example, the [Delete environment](/tag/Environments#operation/deleteEnvironment) endpoint is considered a single route, and each call to delete an environment counts against your route-level rate limit for that route.  We do not publicly document the specific number of calls that can be made to each endpoint per ten seconds. These limits may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limits.  ### IP-based rate limiting  We also employ IP-based rate limiting on some API routes. If you hit an IP-based rate limit, your API response will include a `Retry-After` header indicating how long to wait before re-trying the call. Clients must wait at least `Retry-After` seconds before making additional calls to our API, and should employ jitter and backoff strategies to avoid triggering rate limits again.  ## OpenAPI (Swagger)  We have a [complete OpenAPI (Swagger) specification](https://app.launchdarkly.com/api/v2/openapi.json) for our API.  You can use this specification to generate client libraries to interact with our REST API in your language of choice.  This specification is supported by several API-based tools such as Postman and Insomnia. In many cases, you can directly import our specification to ease use in navigating the APIs in the tooling.  ## Client libraries  We auto-generate multiple client libraries based on our OpenAPI specification. To learn more, visit [GitHub](https://github.com/search?q=topic%3Alaunchdarkly-api+org%3Alaunchdarkly&type=Repositories).  ## Method Overriding  Some firewalls and HTTP clients restrict the use of verbs other than `GET` and `POST`. In those environments, our API endpoints that use `PUT`, `PATCH`, and `DELETE` verbs will be inaccessible.  To avoid this issue, our API supports the `X-HTTP-Method-Override` header, allowing clients to \"tunnel\" `PUT`, `PATCH`, and `DELETE` requests via a `POST` request.  For example, if you wish to call one of our `PATCH` resources via a `POST` request, you can include `X-HTTP-Method-Override:PATCH` as a header.  ## Beta resources  We sometimes release new API resources in **beta** status before we release them with general availability.  Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  We try to promote resources into general availability as quickly as possible. This happens after sufficient testing and when we're satisfied that we no longer need to make backwards-incompatible changes.  We mark beta resources with a \"Beta\" callout in our documentation, pictured below:  > ### This feature is in beta > > To use this feature, pass in a header including the `LD-API-Version` key with value set to `beta`. Use this header with each call. To learn more, read [Beta resources](/#section/Beta-resources).  ### Using beta resources  To use a beta resource, you must include a header in the request. If you call a beta resource without this header, you'll receive a `403` response.  Use this header:  ``` LD-API-Version: beta ```  ## Versioning  We try hard to keep our REST API backwards compatible, but we occasionally have to make backwards-incompatible changes in the process of shipping new features. These breaking changes can cause unexpected behavior if you don't prepare for them accordingly.  Updates to our REST API include support for the latest features in LaunchDarkly. We also release a new version of our REST API every time we make a breaking change. We provide simultaneous support for multiple API versions so you can migrate from your current API version to a new version at your own pace.  ### Setting the API version per request  You can set the API version on a specific request by sending an `LD-API-Version` header, as shown in the example below:  ``` LD-API-Version: 20191212 ```  The header value is the version number of the API version you'd like to request. The number for each version corresponds to the date the version was released. In the example above the version `20191212` corresponds to December 12, 2019.  ### Setting the API version per access token  When creating an access token, you must specify a specific version of the API to use. This ensures that integrations using this token cannot be broken by version changes.  Tokens created before versioning was released have their version set to `20160426` (the version of the API that existed before versioning) so that they continue working the same way they did before versioning.  If you would like to upgrade your integration to use a new API version, you can explicitly set the header described above.  > ### Best practice: Set the header for every client or integration > > We recommend that you set the API version header explicitly in any client or integration you build. > > Only rely on the access token API version during manual testing. 

API version: 2.0
Contact: support@launchdarkly.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package ldapi

import (
	"encoding/json"
)

// UserSegment struct for UserSegment
type UserSegment struct {
	// A human-friendly name for the segment
	Name string `json:"name"`
	// A description of the segment's purpose
	Description *string `json:"description,omitempty"`
	// Tags for the segment
	Tags []string `json:"tags"`
	CreationDate int64 `json:"creationDate"`
	// A unique key used to reference the segment
	Key string `json:"key"`
	// Included users are always segment members, regardless of segment rules. For Big Segments this array is either empty or omitted entirely.
	Included *[]string `json:"included,omitempty"`
	// Segment rules bypass excluded users, so they will never be included based on rules. Excluded users may still be included explicitly. This value is omitted for Big Segments.
	Excluded *[]string `json:"excluded,omitempty"`
	Links map[string]Link `json:"_links"`
	Rules []UserSegmentRule `json:"rules"`
	Version int32 `json:"version"`
	Deleted bool `json:"deleted"`
	Access *AccessRep `json:"_access,omitempty"`
	Flags *[]FlagListingRep `json:"_flags,omitempty"`
	Unbounded *bool `json:"unbounded,omitempty"`
	Generation int32 `json:"generation"`
	UnboundedMetadata *SegmentMetadata `json:"_unboundedMetadata,omitempty"`
	External *string `json:"_external,omitempty"`
	ExternalLink *string `json:"_externalLink,omitempty"`
	ImportInProgress *bool `json:"_importInProgress,omitempty"`
}

// NewUserSegment instantiates a new UserSegment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUserSegment(name string, tags []string, creationDate int64, key string, links map[string]Link, rules []UserSegmentRule, version int32, deleted bool, generation int32) *UserSegment {
	this := UserSegment{}
	this.Name = name
	this.Tags = tags
	this.CreationDate = creationDate
	this.Key = key
	this.Links = links
	this.Rules = rules
	this.Version = version
	this.Deleted = deleted
	this.Generation = generation
	return &this
}

// NewUserSegmentWithDefaults instantiates a new UserSegment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUserSegmentWithDefaults() *UserSegment {
	this := UserSegment{}
	return &this
}

// GetName returns the Name field value
func (o *UserSegment) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetNameOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UserSegment) SetName(v string) {
	o.Name = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *UserSegment) GetDescription() string {
	if o == nil || o.Description == nil {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetDescriptionOk() (*string, bool) {
	if o == nil || o.Description == nil {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *UserSegment) HasDescription() bool {
	if o != nil && o.Description != nil {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *UserSegment) SetDescription(v string) {
	o.Description = &v
}

// GetTags returns the Tags field value
func (o *UserSegment) GetTags() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetTagsOk() (*[]string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Tags, true
}

// SetTags sets field value
func (o *UserSegment) SetTags(v []string) {
	o.Tags = v
}

// GetCreationDate returns the CreationDate field value
func (o *UserSegment) GetCreationDate() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.CreationDate
}

// GetCreationDateOk returns a tuple with the CreationDate field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetCreationDateOk() (*int64, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.CreationDate, true
}

// SetCreationDate sets field value
func (o *UserSegment) SetCreationDate(v int64) {
	o.CreationDate = v
}

// GetKey returns the Key field value
func (o *UserSegment) GetKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetKeyOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *UserSegment) SetKey(v string) {
	o.Key = v
}

// GetIncluded returns the Included field value if set, zero value otherwise.
func (o *UserSegment) GetIncluded() []string {
	if o == nil || o.Included == nil {
		var ret []string
		return ret
	}
	return *o.Included
}

// GetIncludedOk returns a tuple with the Included field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetIncludedOk() (*[]string, bool) {
	if o == nil || o.Included == nil {
		return nil, false
	}
	return o.Included, true
}

// HasIncluded returns a boolean if a field has been set.
func (o *UserSegment) HasIncluded() bool {
	if o != nil && o.Included != nil {
		return true
	}

	return false
}

// SetIncluded gets a reference to the given []string and assigns it to the Included field.
func (o *UserSegment) SetIncluded(v []string) {
	o.Included = &v
}

// GetExcluded returns the Excluded field value if set, zero value otherwise.
func (o *UserSegment) GetExcluded() []string {
	if o == nil || o.Excluded == nil {
		var ret []string
		return ret
	}
	return *o.Excluded
}

// GetExcludedOk returns a tuple with the Excluded field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExcludedOk() (*[]string, bool) {
	if o == nil || o.Excluded == nil {
		return nil, false
	}
	return o.Excluded, true
}

// HasExcluded returns a boolean if a field has been set.
func (o *UserSegment) HasExcluded() bool {
	if o != nil && o.Excluded != nil {
		return true
	}

	return false
}

// SetExcluded gets a reference to the given []string and assigns it to the Excluded field.
func (o *UserSegment) SetExcluded(v []string) {
	o.Excluded = &v
}

// GetLinks returns the Links field value
func (o *UserSegment) GetLinks() map[string]Link {
	if o == nil {
		var ret map[string]Link
		return ret
	}

	return o.Links
}

// GetLinksOk returns a tuple with the Links field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetLinksOk() (*map[string]Link, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Links, true
}

// SetLinks sets field value
func (o *UserSegment) SetLinks(v map[string]Link) {
	o.Links = v
}

// GetRules returns the Rules field value
func (o *UserSegment) GetRules() []UserSegmentRule {
	if o == nil {
		var ret []UserSegmentRule
		return ret
	}

	return o.Rules
}

// GetRulesOk returns a tuple with the Rules field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetRulesOk() (*[]UserSegmentRule, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Rules, true
}

// SetRules sets field value
func (o *UserSegment) SetRules(v []UserSegmentRule) {
	o.Rules = v
}

// GetVersion returns the Version field value
func (o *UserSegment) GetVersion() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Version
}

// GetVersionOk returns a tuple with the Version field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetVersionOk() (*int32, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Version, true
}

// SetVersion sets field value
func (o *UserSegment) SetVersion(v int32) {
	o.Version = v
}

// GetDeleted returns the Deleted field value
func (o *UserSegment) GetDeleted() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Deleted
}

// GetDeletedOk returns a tuple with the Deleted field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetDeletedOk() (*bool, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Deleted, true
}

// SetDeleted sets field value
func (o *UserSegment) SetDeleted(v bool) {
	o.Deleted = v
}

// GetAccess returns the Access field value if set, zero value otherwise.
func (o *UserSegment) GetAccess() AccessRep {
	if o == nil || o.Access == nil {
		var ret AccessRep
		return ret
	}
	return *o.Access
}

// GetAccessOk returns a tuple with the Access field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetAccessOk() (*AccessRep, bool) {
	if o == nil || o.Access == nil {
		return nil, false
	}
	return o.Access, true
}

// HasAccess returns a boolean if a field has been set.
func (o *UserSegment) HasAccess() bool {
	if o != nil && o.Access != nil {
		return true
	}

	return false
}

// SetAccess gets a reference to the given AccessRep and assigns it to the Access field.
func (o *UserSegment) SetAccess(v AccessRep) {
	o.Access = &v
}

// GetFlags returns the Flags field value if set, zero value otherwise.
func (o *UserSegment) GetFlags() []FlagListingRep {
	if o == nil || o.Flags == nil {
		var ret []FlagListingRep
		return ret
	}
	return *o.Flags
}

// GetFlagsOk returns a tuple with the Flags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetFlagsOk() (*[]FlagListingRep, bool) {
	if o == nil || o.Flags == nil {
		return nil, false
	}
	return o.Flags, true
}

// HasFlags returns a boolean if a field has been set.
func (o *UserSegment) HasFlags() bool {
	if o != nil && o.Flags != nil {
		return true
	}

	return false
}

// SetFlags gets a reference to the given []FlagListingRep and assigns it to the Flags field.
func (o *UserSegment) SetFlags(v []FlagListingRep) {
	o.Flags = &v
}

// GetUnbounded returns the Unbounded field value if set, zero value otherwise.
func (o *UserSegment) GetUnbounded() bool {
	if o == nil || o.Unbounded == nil {
		var ret bool
		return ret
	}
	return *o.Unbounded
}

// GetUnboundedOk returns a tuple with the Unbounded field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetUnboundedOk() (*bool, bool) {
	if o == nil || o.Unbounded == nil {
		return nil, false
	}
	return o.Unbounded, true
}

// HasUnbounded returns a boolean if a field has been set.
func (o *UserSegment) HasUnbounded() bool {
	if o != nil && o.Unbounded != nil {
		return true
	}

	return false
}

// SetUnbounded gets a reference to the given bool and assigns it to the Unbounded field.
func (o *UserSegment) SetUnbounded(v bool) {
	o.Unbounded = &v
}

// GetGeneration returns the Generation field value
func (o *UserSegment) GetGeneration() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Generation
}

// GetGenerationOk returns a tuple with the Generation field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetGenerationOk() (*int32, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Generation, true
}

// SetGeneration sets field value
func (o *UserSegment) SetGeneration(v int32) {
	o.Generation = v
}

// GetUnboundedMetadata returns the UnboundedMetadata field value if set, zero value otherwise.
func (o *UserSegment) GetUnboundedMetadata() SegmentMetadata {
	if o == nil || o.UnboundedMetadata == nil {
		var ret SegmentMetadata
		return ret
	}
	return *o.UnboundedMetadata
}

// GetUnboundedMetadataOk returns a tuple with the UnboundedMetadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetUnboundedMetadataOk() (*SegmentMetadata, bool) {
	if o == nil || o.UnboundedMetadata == nil {
		return nil, false
	}
	return o.UnboundedMetadata, true
}

// HasUnboundedMetadata returns a boolean if a field has been set.
func (o *UserSegment) HasUnboundedMetadata() bool {
	if o != nil && o.UnboundedMetadata != nil {
		return true
	}

	return false
}

// SetUnboundedMetadata gets a reference to the given SegmentMetadata and assigns it to the UnboundedMetadata field.
func (o *UserSegment) SetUnboundedMetadata(v SegmentMetadata) {
	o.UnboundedMetadata = &v
}

// GetExternal returns the External field value if set, zero value otherwise.
func (o *UserSegment) GetExternal() string {
	if o == nil || o.External == nil {
		var ret string
		return ret
	}
	return *o.External
}

// GetExternalOk returns a tuple with the External field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExternalOk() (*string, bool) {
	if o == nil || o.External == nil {
		return nil, false
	}
	return o.External, true
}

// HasExternal returns a boolean if a field has been set.
func (o *UserSegment) HasExternal() bool {
	if o != nil && o.External != nil {
		return true
	}

	return false
}

// SetExternal gets a reference to the given string and assigns it to the External field.
func (o *UserSegment) SetExternal(v string) {
	o.External = &v
}

// GetExternalLink returns the ExternalLink field value if set, zero value otherwise.
func (o *UserSegment) GetExternalLink() string {
	if o == nil || o.ExternalLink == nil {
		var ret string
		return ret
	}
	return *o.ExternalLink
}

// GetExternalLinkOk returns a tuple with the ExternalLink field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExternalLinkOk() (*string, bool) {
	if o == nil || o.ExternalLink == nil {
		return nil, false
	}
	return o.ExternalLink, true
}

// HasExternalLink returns a boolean if a field has been set.
func (o *UserSegment) HasExternalLink() bool {
	if o != nil && o.ExternalLink != nil {
		return true
	}

	return false
}

// SetExternalLink gets a reference to the given string and assigns it to the ExternalLink field.
func (o *UserSegment) SetExternalLink(v string) {
	o.ExternalLink = &v
}

// GetImportInProgress returns the ImportInProgress field value if set, zero value otherwise.
func (o *UserSegment) GetImportInProgress() bool {
	if o == nil || o.ImportInProgress == nil {
		var ret bool
		return ret
	}
	return *o.ImportInProgress
}

// GetImportInProgressOk returns a tuple with the ImportInProgress field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetImportInProgressOk() (*bool, bool) {
	if o == nil || o.ImportInProgress == nil {
		return nil, false
	}
	return o.ImportInProgress, true
}

// HasImportInProgress returns a boolean if a field has been set.
func (o *UserSegment) HasImportInProgress() bool {
	if o != nil && o.ImportInProgress != nil {
		return true
	}

	return false
}

// SetImportInProgress gets a reference to the given bool and assigns it to the ImportInProgress field.
func (o *UserSegment) SetImportInProgress(v bool) {
	o.ImportInProgress = &v
}

func (o UserSegment) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["name"] = o.Name
	}
	if o.Description != nil {
		toSerialize["description"] = o.Description
	}
	if true {
		toSerialize["tags"] = o.Tags
	}
	if true {
		toSerialize["creationDate"] = o.CreationDate
	}
	if true {
		toSerialize["key"] = o.Key
	}
	if o.Included != nil {
		toSerialize["included"] = o.Included
	}
	if o.Excluded != nil {
		toSerialize["excluded"] = o.Excluded
	}
	if true {
		toSerialize["_links"] = o.Links
	}
	if true {
		toSerialize["rules"] = o.Rules
	}
	if true {
		toSerialize["version"] = o.Version
	}
	if true {
		toSerialize["deleted"] = o.Deleted
	}
	if o.Access != nil {
		toSerialize["_access"] = o.Access
	}
	if o.Flags != nil {
		toSerialize["_flags"] = o.Flags
	}
	if o.Unbounded != nil {
		toSerialize["unbounded"] = o.Unbounded
	}
	if true {
		toSerialize["generation"] = o.Generation
	}
	if o.UnboundedMetadata != nil {
		toSerialize["_unboundedMetadata"] = o.UnboundedMetadata
	}
	if o.External != nil {
		toSerialize["_external"] = o.External
	}
	if o.ExternalLink != nil {
		toSerialize["_externalLink"] = o.ExternalLink
	}
	if o.ImportInProgress != nil {
		toSerialize["_importInProgress"] = o.ImportInProgress
	}
	return json.Marshal(toSerialize)
}

type NullableUserSegment struct {
	value *UserSegment
	isSet bool
}

func (v NullableUserSegment) Get() *UserSegment {
	return v.value
}

func (v *NullableUserSegment) Set(val *UserSegment) {
	v.value = val
	v.isSet = true
}

func (v NullableUserSegment) IsSet() bool {
	return v.isSet
}

func (v *NullableUserSegment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUserSegment(val *UserSegment) *NullableUserSegment {
	return &NullableUserSegment{value: val, isSet: true}
}

func (v NullableUserSegment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUserSegment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


