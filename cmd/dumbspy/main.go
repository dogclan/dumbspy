package main

import (
	"cmp"
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
	errorCodeLoginFailed = "256"
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

	challenge := gamespy.RandString(10)
	prompt := new(gamespy.Packet)
	prompt.Add("lc", "1")
	prompt.Add("challenge", challenge)
	prompt.Add("id", "1")

	log.Debug().
		Bytes(logKeyData, prompt.Bytes()).
		Str(logKeyRemote, remoteAddr).
		Msg("Sending challenge prompt")
	if err := write(conn, prompt); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to send challenge")
		return
	}

	log.Debug().
		Str(logKeyRemote, remoteAddr).
		Msg("Reading login request")
	req, err := read(conn)
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
		Bytes(logKeyData, req.Bytes()).
		Str(logKeyRemote, remoteAddr).
		Msg("Received login request")

	res := new(gamespy.Packet)
	var login internal.GamespyLoginRequest
	if err = cmp.Or(req.Bind(&login), login.Validate()); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Received invalid login request")

		res.Add("error", "")
		res.Add("err", errorCodeLoginFailed)
		res.Add("fatal", "")
		res.Add("errmsg", "There was an error logging in to the GP backend.")
		res.Add("id", "1")

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
		res.Add("lc", "2")
		res.AddInt("sesskey", int(gamespy.ComputeCRC16(login.UniqueNick)))
		res.Add("proof", gamespy.GenerateProof(
			login.UniqueNick,
			login.Response,
			challenge,
			login.Challenge,
		))
		res.AddInt("userid", playerID)
		res.AddInt("profileid", playerID)
		res.Add("uniquenick", login.UniqueNick)
		res.Add("lt", gamespy.RandString(22)+"__")
		res.Add("id", "1")

		log.Debug().
			Bytes(logKeyData, res.Bytes()).
			Str(logKeyRemote, remoteAddr).
			Msg("Sending login response")
	}

	if err = write(conn, res); err != nil {
		log.Error().
			Err(err).
			Str(logKeyRemote, remoteAddr).
			Msg("Failed to send response")
	}
}

func write(conn net.Conn, packet *gamespy.Packet) error {
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := conn.Write(packet.Bytes()); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}
	return nil
}

func read(conn net.Conn) (*gamespy.Packet, error) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	buffer := make([]byte, 512)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	packet, err := gamespy.NewPacketFromBytes(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse packet: %w", err)
	}
	return packet, nil
}
