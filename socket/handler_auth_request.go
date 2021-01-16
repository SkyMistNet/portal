package socket

import (
	"bytes"
	"fmt"
	"github.com/paroxity/portal/config"
	"github.com/paroxity/portal/server"
	portalpacket "github.com/paroxity/portal/socket/packet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// AuthRequestHandler is responsible for handling the AuthRequest packet sent by servers.
type AuthRequestHandler struct{}

// Handle ...
func (*AuthRequestHandler) Handle(p packet.Packet, c *Client) error {
	pk := p.(*portalpacket.AuthRequest)

	if pk.Secret != config.SocketSecret() {
		return c.WritePacket(&portalpacket.AuthResponse{
			Status: portalpacket.AuthResponseIncorrectSecret,
			Reason: "The provided secret does not match the configured secret",
		})
	}

	data := bytes.NewBuffer(pk.ExtraData)
	r := protocol.NewReader(data, 0)
	switch pk.Type {
	case portalpacket.ClientTypeServer:
		var group, address string
		r.String(&group)
		r.String(&address)

		g, ok := server.GroupFromName(group)
		if !ok {
			return c.WritePacket(&portalpacket.AuthResponse{
				Status: portalpacket.AuthResponseInvalidData,
				Reason: fmt.Sprintf("Group %s not found", group),
			})
		}

		if _, ok := g.Server(pk.Name); ok {
			return c.WritePacket(&portalpacket.AuthResponse{
				Status: portalpacket.AuthResponseInvalidData,
				Reason: "A server with the same name is already authenticated",
			})
		}

		_, _ = server.New(pk.Name, group, address)
		c.name = pk.Name
		c.clientType = pk.Type
		c.extraData["address"] = address
		c.extraData["group"] = g.Name()

		clientsMu.Lock()
		clients[pk.Name] = c
		clientsMu.Unlock()
	default:
		return c.WritePacket(&portalpacket.AuthResponse{
			Status: portalpacket.AuthResponseUnknownType,
			Reason: "Unknown client type provided",
		})
	}

	return c.WritePacket(&portalpacket.AuthResponse{
		Status: portalpacket.AuthResponseSuccess,
		Reason: "Authentication was successful",
	})
}
