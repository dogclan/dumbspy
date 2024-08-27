package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"dogclan/dumbspy/cmd/dumbspy/internal/options"
	"dogclan/dumbspy/internal"
	"dogclan/dumbspy/pkg/packet"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	network      = "tcp4"
	logKeyRemote = "remote"
	logKeyData   = "data"
)

var (
	buildVersion = "development"
	buildCommit  = "uncommitted"
	buildTime    = "unknown"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	version := fmt.Sprintf("dumbspy %s (%s) built at %s", buildVersion, buildCommit, buildTime)
	opts := options.Init()

	// Print version and exit
	if opts.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: !opts.ColorizeLogs})
	if opts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	listen, err := net.Listen(network, opts.ListenAddr)
	if err != nil {
		log.Fatal().
			Err(err).
			Msgf("Failed to start listener")
	}

	log.Info().
		Str("address", opts.ListenAddr).
		Msg("Listening for connections")

	// close listener
	defer func(listen net.Listener) {
		err2 := listen.Close()
		if err2 != nil {
			log.Error().
				Err(err2).
				Msg("Failed to close listener")
		}
	}(listen)

	for {
		conn, err2 := listen.Accept()
		if err2 != nil {
			log.Error().
				Err(err2).
				Msg("Failed to accept new connection")
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().
				Err(err).
				Str(logKeyRemote, remoteAddr).
				Msg("Failed to close connection")
		}
	}(conn)

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to set read deadline")
		return
	}

	challenge := internal.RandString(10)
	prompt := new(packet.GamespyPacket)
	prompt.Write("lc", "1")
	prompt.Write("challenge", challenge)
	prompt.Write("id", "1")

	log.Debug().
		Bytes(logKeyData, prompt.Bytes()).
		Str(logKeyRemote, remoteAddr).
		Msg("Sending challenge prompt")
	if _, err := conn.Write(prompt.Bytes()); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to send challenge")
		return
	}

	log.Debug().
		Str(logKeyRemote, remoteAddr).
		Msg("Reading login request")
	buffer := make([]byte, 512)
	n, err := conn.Read(buffer)
	if err != nil {
		// EOF and timeout errors are not of interest => only log to debug
		if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) {
			log.Debug().
				Str(logKeyRemote, remoteAddr).
				Msg("Peer closed/reset connection while reading login request")
		} else if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Debug().
				Str(logKeyRemote, remoteAddr).
				Msg("Timed out reading login request")
		} else {
			log.Error().
				Err(err).
				Str(logKeyRemote, remoteAddr).
				Msg("Failed to read login request")
		}
		return
	}

	log.Debug().
		Bytes(logKeyData, buffer[:n]).
		Str(logKeyRemote, remoteAddr).
		Msg("Received login request")
	req, err := packet.FromString(string(buffer[:n]))
	if err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to parse login request packet")
		return
	}

	res := new(packet.GamespyPacket)
	login := internal.NewGamespyLoginRequest(req)
	if err = login.Validate(); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Received invalid login request")

		res.Write("error", "")
		res.Write("err", "0")
		res.Write("fatal", "")
		res.Write("errmsg", "Invalid Query!")
		res.Write("id", "1")

		log.Debug().
			Bytes("data", res.Bytes()).
			Str(logKeyRemote, remoteAddr).
			Msg("Sending error response")
	} else {
		playerID := internal.GetPlayerID(
			login.UniqueNick,
			login.ProductID,
			login.GameName,
			login.NamespaceID,
			login.SDKRevision,
		)
		res.Write("lc", "2")
		res.Write("sesskey", internal.ComputeCRC16Str(login.UniqueNick))
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
			Bytes(logKeyData, res.Bytes()).
			Str(logKeyRemote, remoteAddr).
			Msg("Sending login response")
	}

	if _, err = conn.Write(res.Bytes()); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to send response")
	}
}
