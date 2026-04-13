package server

import (
    "net"
    "context"
    "errors"
    "strings"
    "bufio"
    "math/rand/v2"
    "time"
    "fmt"

    "github.com/Tinch334/Token-File-Sharing/tokens"
    "github.com/Tinch334/Token-File-Sharing/workers"
    "github.com/Tinch334/Token-File-Sharing/metrics"

    "github.com/rs/zerolog/log"
    "golang.org/x/sync/errgroup"
)


// The server is the main component of the TFS system, it creates and handles all other systems.
type Server struct {
    ctx    context.Context
    cancel context.CancelFunc

    port   string

    th     *tokens.TokenHandler[int]
    wp     *workers.WorkerPool[net.Conn, int]
    mtr    *metrics.Metrics
}


// NewServer creates a new server on the given port.
func NewServer(port string) *Server {
    ctx, cancel := context.WithCancel(context.Background())
    th := tokens.NewTokenHandler[int](ctx, 8, true, 5*time.Second)
    wp := workers.CreateWorkerPool[net.Conn, int](10, 20, 20)
    mtr := metrics.InitMetrics()

    return &Server{
        ctx:    ctx,
        cancel: cancel,

        port:   port,

        th:     th,
        wp:     wp,
        mtr:    mtr,
    }
}


// StartServer starts the server on it's given port.
func (s *Server) StartServer() {
    // Ensures all parts of the server finish before exiting.
    wg, ctx := errgroup.WithContext(s.ctx)

    // Start worker pool.
    s.wp.StartWorkerPool(ctx, func(ctx context.Context, conn net.Conn) (int, bool) {
        return s.handleConnection(ctx, conn)
    })

    // Generate tokens for testing.
    for i := 0; i < 10; i++ {
        et := time.Duration(10 * (1 + rand.IntN(6)))
        nt := s.th.GenerateToken(i, et*time.Second)
        fmt.Printf("\"%x\", et: %v\n", nt, et*time.Second)
    }


    // Start server and add to wait group.
    wg.Go(func() error {
        return s.serverRunner(ctx)
    })
    // Wait for token handler to finish.
    wg.Go(func() error {
        s.th.Wait()
        return nil
    })
    // Wait for worker pool to finish.
    wg.Go(func() error {
        s.wp.Wait()
        return nil
    })

    // Start server cli, this function returns on finish, allowing for context cancellation.
    s.cli()

    // Cancel context and wait for all parts to finish.
    s.cancel()
    if err := wg.Wait(); err != nil {
        log.Fatal().
            Err(err).
            Msg("Shutdown error")
    }
}


// serverRunner runs the server.
func (s *Server) serverRunner(ctx context.Context) error {
    listener, err := net.Listen("tcp", s.port)

    // When context is cancelled close the listener, forces listener to unblock and return an error.
    cleanup := context.AfterFunc(ctx, func() {
        log.Info().
            Msg("Context cancelled, sever listener closing")

        listener.Close()
    })
    // In case the server function ends for another reason, even though it shouldn't happen.
    defer cleanup()

    // Check for server error.
    if err != nil {
        log.Fatal().
            Err(err).
            Msg("Server listener could not be started")

        return err
    }

    log.Info().
        Msg("Server listener started")

    // Accept connections.
    for {
        conn, err := listener.Accept()

        // An error occurred accepting the connection.
        if err != nil {
            // Check if the error is just because the listener was closed.
            if errors.Is(err, net.ErrClosed) {
                log.Info().
                    Msg("Context cancelled, listener closed gracefully")
                break
            }

            log.Error().
                Err(err).
                Msg("Server encountered an error accepting a connection")

            continue
        }

        if s.wp.Submit(ctx, conn) != nil {
            log.Error().
                Err(err).
                Msg("Server encountered an error submitting a connection to worker pool")

            // Connection cannot be processed close it.
            conn.Close()
            continue
        } else {
            // Add connection to metrics
            s.mtr.AddConn()
        }
    }

    return nil
}


// Handles incoming connections to the server.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) (int, bool) {
    reader := bufio.NewReader(conn)
    message, err := reader.ReadString('\n')
    if err != nil {
        log.Printf("Read error: %v", err)
        return 0, false
    }


    ackMsg := strings.ToUpper(strings.TrimSpace(message))
    response := fmt.Sprintf("ACK: %s\n", ackMsg)
    _, err = conn.Write([]byte(response))
    if err != nil {
        log.Printf("Server write error: %v", err)
    }

    return 0, false
}