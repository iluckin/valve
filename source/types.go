package source

import "net"

type ServerType int

const (
	DefaultMaxPacketSize = 1400
	ServerType_Unknown ServerType = iota
	ServerType_Dedicated
	ServerType_NonDedicated
	ServerType_SourceTV
)

type Address struct {
	Host string `json:"host"`
	Port uint32	`json:"port"`
}

type Querier struct {
	addr Address
	conn net.Conn
	ver int32
}

type SourceTVInfo struct {
	Port uint16 `json:"Port"` // Spectator port number for SourceTV.
	Name string `json:"Name"` 	// Name of the spectator server for SourceTV.
}

type ServerInfo struct {
	Protocol uint8 `json:"protocol"`					// Protocol version used by the server.
	Name string `json:"name"`							// Name of the server.
	Map string `json:"map"`								// Map the server has currently loaded.
	Folder string `json:"folder"`						// Name of the folder containing the game files.
	Game string `json:"game"`							// Full name of the game.
	ID uint16 `json:"appID"` 							// Steam Application ID of game.
	Players uint8 `json:"players"`						// Number of players on the server
	MaxPlayers uint8 `json:"maxPlayers"`				// Maximum number of players the server reports it can hold.
	Bots uint8 `json:"Bots"`							// Number of bots on the server.
	ServerType ServerType `json:"serverType"`			// Rag Doll Kung Fu servers always return 0 for "Server type."
	Platform string `json:"platform"`					// The operating system of the server.
	Locked bool `json:"locked"`							// Indicates whether the server requires a password.
	VAC bool `json:"vac"`								// Specifies whether the server uses VAC
	Version string `json:"version"`						// Version of the game installed on the server.
	SourceTV *SourceTVInfo `json:"sourceTV,omitempty"`
	Address Address `json:"address"`					// The Server address (host or port).
}

func (t ServerType) String() string {
	switch t {
	case ServerType_Dedicated:
		return "Dedicated"
	case ServerType_NonDedicated:
		return "Non-Dedicated"
	case ServerType_SourceTV:
		return "SourceTV"
	default:
		return "Unknown"
	}
}