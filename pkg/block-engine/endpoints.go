package block_engine_pkg

// Constants for block engine endpoints
const (
	AMS = "https://amsterdam.mainnet.block-engine.jito.wtf" // Amsterdam endpoint
	FRA = "https://frankfurt.mainnet.block-engine.jito.wtf" // Frankfurt endpoint
	NYC = "https://ny.mainnet.block-engine.jito.wtf"        // New York City endpoint
	TKO = "https://tokyo.mainnet.block-engine.jito.wtf"     // Tokyo endpoint
)

// GetEndpoint returns the endpoint URL for the given location code.
// It returns an empty string if the location code is not recognized.
func GetEndpoint(location string) string {
	switch location {
	case "AMS":
		return AMS
	case "FRA":
		return FRA
	case "NYC":
		return NYC
	case "TKO":
		return TKO
	default:
		return ""
	}
}
