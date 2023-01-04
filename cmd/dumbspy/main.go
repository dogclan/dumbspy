package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"cetteup/dumbspy/internal"
	"cetteup/dumbspy/pkg/packet"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	port    = 29900
	network = "tcp"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	var err error
	config, err = internal.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	zerolog.SetGlobalLevel(config.LogLevel)
}

var (
	config *internal.Config
)

func main() {
	listen, err := net.Listen(network, fmt.Sprintf("%s:%d", config.Host, port))
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to start listener")
	}

	log.Info().
		Str("host", config.Host).
		Int("port", port).
		Msg("Listening for connections")

	// close listener
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close listener")
		}
	}(listen)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error().Err(err).Msg("Failed to accept new connection")
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close connection")
		}
	}(conn)

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		log.Error().Err(err).Msg("Failed to set read deadline")
		return
	}

	challenge := internal.RandString(10)
	prompt := new(packet.GamespyPacket)
	prompt.Write("lc", "1")
	prompt.Write("challenge", challenge)
	prompt.Write("id", "1")

	log.Debug().
		Bytes("data", prompt.Bytes()).
		Msg("Sending challenge prompt")
	if _, err := conn.Write(prompt.Bytes()); err != nil {
		log.Error().Err(err).Msg("Failed to send challenge")
		return
	}

	log.Debug().
		Msg("Reading login request")
	buffer := make([]byte, 512)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read login request")
		return
	}

	log.Debug().
		Bytes("response", buffer[:n]).
		Msg("Received login request")
	req, err := packet.FromString(string(buffer[:n]))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse login request packet")
		return
	}

	res := new(packet.GamespyPacket)
	login := internal.NewGamespyLoginRequest(req)
	if err := login.Validate(); err != nil {
		log.Error().Err(err).Msg("Received invalid login request")

		res.Write("error", "")
		res.Write("err", "0")
		res.Write("fatal", "")
		res.Write("errmsg", "Invalid Query!")
		res.Write("id", "1")

		log.Debug().
			Bytes("data", res.Bytes()).
			Msg("Sending error response")
	} else {
		playerID := internal.GetPlayerID(login.UniqueNick, login.Response)
		res.Write("lc", "2")
		res.Write("sesskey", internal.ComputeChecksum(login.UniqueNick))
		res.Write("proof", internal.GenerateProof(
			login.UniqueNick,
			login.Response,
			challenge,
			login.Challenge,
		))
		res.Write("userid", strconv.Itoa(playerID))
		res.Write("profileid", strconv.Itoa(playerID))
		res.Write("uniquenick", login.UniqueNick)
		res.Write("lt", fmt.Sprintf("%s__", internal.RandString(22)))
		res.Write("id", "1")

		log.Debug().
			Bytes("data", res.Bytes()).
			Msg("Sending login response")
	}

	if _, err := conn.Write(res.Bytes()); err != nil {
		log.Error().Err(err).Msg("Failed to send response")
	}
}
