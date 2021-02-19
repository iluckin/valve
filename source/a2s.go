package source

import (
	"encoding/binary"
	"errors"
	"fmt"
	"iluckin.cn/valve/utils/packet"
	"net"
	"strconv"
	"strings"
	"time"
)

func NewQuerier(network string, timeout time.Duration) (*Querier, error) {
	ni := strings.Split(network, ":")
	ip := net.ParseIP(ni[0])
	if ip == nil {
		return nil, fmt.Errorf("Illegal IP address: %s", ni[0])
	}

	p, e := strconv.Atoi(ni[1])
	if e != nil {
		return nil, e
	}

	conn, e := net.DialTimeout("udp", network, timeout)
	if e != nil {
		return nil, e
	}

	return &Querier{addr: Address{Host: ip.String(), Port: uint32(p)}, conn: conn}, nil
}

func (q *Querier) GetAddress() Address { return q.addr}
func (q *Querier) Close() { if q.conn != nil { q.conn.Close() }}

func (q *Querier) GetChallengeCode() ([]byte, bool, error) {
	var pb packet.PacketBuilder
	// FF FF FF FF 55 FF FF FF FF
	pb.WriteBytes([]byte{0xff, 0xff, 0xff, 0xff, 0x55, 0xff, 0xff, 0xff, 0xff})
	_, err := q.conn.Write(pb.Bytes())
	if err != nil {
		return nil, false, err
	}

	buf := make([]byte, DefaultMaxPacketSize)
	len, err := q.conn.Read(buf)
	if err != nil {
		return nil, false, err
	}

	data := make([]byte, len)
	copy(data, buf[:len])
	reader := packet.NewPacketReader(data)

	switch int32(reader.ReadUint32()) {
	case -2: // We received an unexpected full reply
		return data, true, nil
	case -1: // Continue
	default:
		return nil, false, errors.New("err bad packet header")
	}

	switch reader.ReadUint8() {
	case 0x41: // Received a challenge number
		return data[reader.Pos() : reader.Pos()+4], false, nil
	case 0x44: // Received full result
		return data, true, nil
	}

	return nil, false, errors.New("err bad challenge response")
}

// Get server info.
func (q *Querier) GetInfo() (*ServerInfo, error) {
	var pb packet.PacketBuilder
	/*
		(FF FF FF FF) 54 53 6F 75 72 63 65 20 45 6E 67 69   ÿÿÿÿTSource Engi
		6E 65 20 51 75 65 72 79 00                        ne Query.
	*/
	pb.WriteBytes([]byte{0xff, 0xff, 0xff, 0xff, 0x54})
	pb.WriteCString("Source Engine Query")

	_, err := q.conn.Write(pb.Bytes())
	if err != nil {
		return nil, err
	}

	buf := make([]byte, DefaultMaxPacketSize)
	_, err = q.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	pr := packet.NewPacketReader(buf[4:])
	pr.ReadUint8(); pr.ReadString()
	s := &ServerInfo{
		Address: q.addr,
		Name: pr.ReadString(),
		Map: pr.ReadString(),
		Folder: pr.ReadString(),
		Game: pr.ReadString(),
		Players: pr.ReadUint8(),
		MaxPlayers: pr.ReadUint8(),
		Protocol: pr.ReadUint8(),
		ServerType: ParseServerType(pr.ReadUint8()),
		Platform: ParsePlatform(pr.ReadUint8()),
		Locked: pr.ReadUint8() == 0x01,
		VAC: pr.ReadUint8() == 0x01,
		Version: pr.ReadString(),
	}

	return s, nil
}

func ParseServerType(sType uint8) ServerType {
	switch sType {
	case uint8('d'):
		return ServerType_Dedicated
	case uint8('l'):
		return ServerType_NonDedicated
	case uint8('p'):
		return ServerType_SourceTV
	}

	return ServerType_Unknown
}

func ParsePlatform(p uint8) string {
	switch p {
	case uint8('l'):
		return "Linux"
	case uint8('w'):
		return "Windows"
	}

	return "Unknown"
}

type Player struct {
	ID uint8 `json:"id"`
	Name string `json:"name"`
	Score uint32 `json:"score"`
	Duration float32 `json:"duration"`
	TheShip *TheShipPlayer `json:"theShip"`
}

type TheShipPlayer struct {
	Deaths uint32 `json:"deaths"`
	Money uint32 `json:"money"`
}

type PlayerInfo struct {
	Count uint8 `json:"count"`
	Players []*Player `json:"players"`
}

func (q *Querier) GetPlayerInfo() (*PlayerInfo, error) {
	data, success, err := q.GetChallengeCode()
	if err != nil {
		return nil, err
	}

	if !success {
		_, err := q.conn.Write([]byte{0xff, 0xff, 0xff, 0xff, 0x55, data[0], data[1], data[2], data[3]})
		if err != nil {
			return nil, err
		}

		buf := make([]byte, DefaultMaxPacketSize)
		length, err := q.conn.Read(buf)
		if err != nil {
			return nil, err
		}

		data = buf[:length]
	}

	// Read header (long 4 bytes)
	switch int32(binary.LittleEndian.Uint32(data)) {
	case -1:
		return q.parsePlayerInfo(data)
	case -2: // collectMultiplePacketResponse
		return nil, errors.New("Multiple packet response.")
	}

	return nil, errors.New("Failed to get player info.")
}

func (q *Querier) parsePlayerInfo(data []byte) (*PlayerInfo, error) {
	reader := packet.NewPacketReader(data)
	if reader.ReadInt32() != -1 { // Simple response now
		return nil, errors.New("Err bad packet header.")
	}

	if reader.ReadUint8() != 0x44 {
		return nil, errors.New("Err bad players reply.")
	}

	var player *Player
	info := &PlayerInfo{}
	info.Count = reader.ReadUint8()

	for i := 0; i < int(info.Count); i++ {
		player = &Player{}

		player.ID = reader.ReadUint8()
		player.Name = reader.ReadString()
		player.Score = reader.ReadUint32()
		player.Duration = reader.ReadFloat32()

		if q.ver == 2400 { // The Ship additional player info only if client AppID is set to 2400
			player.TheShip = &TheShipPlayer{}

			player.TheShip.Deaths = reader.ReadUint32()
			player.TheShip.Money = reader.ReadUint32()
		}

		info.Players = append(info.Players, player)
	}

	return info, nil
}