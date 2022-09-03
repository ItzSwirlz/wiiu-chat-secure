package main

import (
	"strconv"

	nex "github.com/PretendoNetwork/nex-go"

	nexproto "github.com/PretendoNetwork/nex-protocols-go"
)

func register(err error, client *nex.Client, callID uint32, stationUrls []*nex.StationURL) {
	localStation := stationUrls[0]
	localStationURL := localStation.EncodeToString()
	connectionId := uint32(secureServer.ConnectionIDCounter.Increment())
	client.SetConnectionId(connectionId)
	client.SetLocalStationUrl(localStationURL)

	address := client.Address().IP.String()
	port := strconv.Itoa(client.Address().Port)
	natf := "0"
	natm := "0"
	type_ := "3"
	pid := strconv.Itoa(int(client.PID()))

	localStation.SetAddress(&address)
	localStation.SetPort(&port)
	localStation.SetNatf(&natf)
	localStation.SetNatm(&natm)
	localStation.SetType(&type_)
	localStation.SetPid(&pid)

	globalStationURL := localStation.EncodeToString()

	if !doesSessionExist(client.PID()) {
		addPlayerSession(client.PID(), []string{localStationURL, globalStationURL}, address, port)
	} else {
		updatePlayerSessionAll(client.PID(), []string{localStationURL, globalStationURL}, address, port)
	}

	rmcResponseStream := nex.NewStreamOut(nexServer)

	rmcResponseStream.WriteUInt32LE(0x10001) // Success
	rmcResponseStream.WriteUInt32LE(uint32(secureServer.ConnectionIDCounter.Increment()))
	rmcResponseStream.WriteString(localStationURL)

	rmcResponseBody := rmcResponseStream.Bytes()

	// Build response packet
	rmcResponse := nex.NewRMCResponse(nexproto.SecureProtocolID, callID)
	rmcResponse.SetSuccess(nexproto.SecureMethodRegister, rmcResponseBody)

	rmcResponseBytes := rmcResponse.Bytes()

	responsePacket, _ := nex.NewPacketV1(client, nil)

	responsePacket.SetVersion(1)
	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	nexServer.Send(responsePacket)
}
