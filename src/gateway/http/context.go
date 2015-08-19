package http

// ContextKey is the type of Gorilla context keys.
type ContextKey int

const (
	// ContextMatchKey is the key to use to store/retrieve the match data.
	ContextMatchKey ContextKey = iota

	// ContextAPIIDKey is the key to use to store/retrieve the API ID.
	ContextAPIIDKey

	// ContextRequestIDKey is the key to use to store/retrieve the request ID.
	ContextRequestIDKey

	// ContextLogPrefixKey is the key to use to store/retrieve the log prefix.
	ContextLogPrefixKey

	// ContextTest is the key to use for test requests.
	ContextTest

	// ContextAccountIDKey is the key to use to store/retrieve the Account ID.
	ContextAccountIDKey

	// ContextEndpointIDKey is the key to use to store/retrieve the Endpoint ID.
	ContextEndpointIDKey
)
