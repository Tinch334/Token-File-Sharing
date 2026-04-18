package server


import (
    "net"
    "context"
    "errors"
    "time"

    "github.com/rs/zerolog/log"
    "golang.org/x/sync/errgroup"

    "github.com/Tinch334/Token-File-Sharing/internal/tokens"
    "github.com/Tinch334/Token-File-Sharing/internal/workers"
    "github.com/Tinch334/Token-File-Sharing/internal/metrics"
    "github.com/Tinch334/Token-File-Sharing/internal/credentials"
)


// The server is the main component of the TFS system, it creates and handles all other systems.
type Server struct {
    ctx    context.Context
    cancel context.CancelFunc

    port   string

    th     *tokens.TokenHandler[string]
    wp     *workers.WorkerPool[net.Conn, int]
    mtr    *metrics.Metrics
    credDB *credentials.CredentialDB
}


// NewServer creates a new server on the given port.
func NewServer(port string, dbPath string) *Server {
    ctx, cancel := context.WithCancel(context.Background())
    th := tokens.NewTokenHandler[string](ctx, 4, 30*time.Minute, true, 5*time.Second)
    wp := workers.CreateWorkerPool[net.Conn, int](10, 20, 20)
    mtr := metrics.InitMetrics()

    credentialsStmt := `
    CREATE TABLE IF NOT EXISTS users (
        name TEXT NOT NULL PRIMARY KEY UNIQUE,
        password CHAR(60) NOT NULL,
        permissions TIYINT NOT NULL,
        lastLog DATETIME DEFAULT NULL

    )
    `
    credDB := credentials.StartCredentialDB(dbPath, credentialsStmt)

    return &Server{
        ctx:    ctx,
        cancel: cancel,

        port:   port,

        th:     th,
        wp:     wp,
        mtr:    mtr,
        credDB: credDB,
    }
}


// StartServer starts the server on it's given port.
func (s *Server) StartServer() {
    // Ensures all parts of the server finish before exiting.
    wg, ctx := errgroup.WithContext(s.ctx)

    // Start worker pool.
    // A closure is used to give the worker pool a function of the correct type, whilst maintaining struct access.
    s.wp.StartWorkerPool(ctx, func(ctx context.Context, conn net.Conn) (int, error) {
        return s.handleConnection(ctx, conn)
    })

    // Clean worker result channel.
    go func() {
        select {
        case <-s.wp.Results():
            // Ignore results.
        case <-ctx.Done():
            return
        }
    }()

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

    // Start server CLI, this function returns on finish, allowing for context cancellation.
    s.cli(ctx)

    // Cancel context and wait for all parts to finish.
    s.cancel()
    if err := wg.Wait(); err != nil {
        log.Fatal().
            Err(err).
            Msg("Shutdown error")
    }

    // Close the credential database.
    s.credDB.Close()
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
            // Check if the error is because the listener was closed.
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