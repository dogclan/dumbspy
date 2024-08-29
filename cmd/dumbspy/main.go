package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/dogclan/dumbspy/cmd/dumbspy/internal/options"
	"github.com/dogclan/dumbspy/internal"
	"github.com/dogclan/dumbspy/pkg/gamespy"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	network      = "tcp4"
	logKeyRemote = "remote"
	logKeyData   = "data"

	// Following https://github.com/openspy/openspy-core/blob/5993df54c6b289361228920fa0db7209aed4cfe5/code/SharedTasks/src/OS/GPShared.h#L355
	errorCodeLoginFailed = 0x100
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

	challenge := gamespy.RandString(10)
	prompt := new(gamespy.Packet)
	prompt.SetInt("lc", 1)
	prompt.Set("challenge", challenge)
	prompt.SetInt("id", 1)

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
	req, err := gamespy.NewPacketFromBytes(buffer[:n])
	if err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to parse login request packet")
		return
	}

	res := new(gamespy.Packet)
	login := internal.NewGamespyLoginRequest(req)
	if err = login.Validate(); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Received invalid login request")

		res.Set("error", "")
		res.SetInt("err", errorCodeLoginFailed)
		res.Set("fatal", "")
		res.Set("errmsg", "There was an error logging in to the GP backend.")
		res.SetInt("id", 1)

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
		res.SetInt("lc", 2)
		res.SetInt("sesskey", int(gamespy.ComputeCRC16(login.UniqueNick)))
		res.Set("proof", gamespy.GenerateProof(
			login.UniqueNick,
			login.Response,
			challenge,
			login.Challenge,
		))
		res.SetInt("userid", playerID)
		res.SetInt("profileid", playerID)
		res.Set("uniquenick", login.UniqueNick)
		res.Set("lt", gamespy.RandString(22)+"__")
		res.SetInt("id", 1)

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
