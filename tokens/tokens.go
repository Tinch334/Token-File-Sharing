package tokens


import (
    "crypto/rand"
    "sync"
    "time"
    "context"
    "fmt"

    "github.com/Tinch334/bmap"
    "github.com/rs/zerolog/log"
)


type TokenHandler[T comparable] struct {
    tokenLength     uint
    checkDuplicates bool

    // Stores tokens and their associated usernames.
    tokens      *bmap.BMap[string, T]
    // Stores expiration time for tokens.
    tokenTimers map[string]time.Time

    timeMutex   sync.RWMutex

    // USed for a wait function to allow clean termination.
    wg          sync.WaitGroup
}


// NewTokenHandler creates a new, empty token handler.
// "ctx" is a context used to stop the cleaner routine.
// A recommended token length to avoid collisions with reasonable amounts of tokens is 64.
func NewTokenHandler[T comparable](ctx context.Context, tokLen uint, chkDup bool, cleanerInt time.Duration) *TokenHandler[T] {
    th := TokenHandler[T]{
        tokenLength:     tokLen,
        checkDuplicates: chkDup,

        tokens:          bmap.NewBMap[string, T](),
        tokenTimers:     make(map[string]time.Time),
    }

    // Start cleaner thread.
    go th.cleaner(ctx, cleanerInt)

    return &th
}


// validTime checks whether the given token hasn't expired, returns false is the token isn't present or has expired.
func (th *TokenHandler[T]) ValidateTokenTime(tok string) bool {
    th.timeMutex.RLock()
    defer th.timeMutex.RUnlock()

    expTime, prs := th.tokenTimers[tok]

    if !prs {
        return false
    }

    // Check if token has expired.
    return !time.Now().After(expTime)
}

// GenerateToken generates a new token for the given element, the generated token is returned.
// Checks for duplicated tokens if instructed.
func (th *TokenHandler[T]) GenerateToken(elem T, ttl time.Duration) string {
    sb := make([]byte, th.tokenLength)

    for {
        // Read cryptographically secure byte string, note that Read cannot fail.
        rand.Read(sb)
        st := string(sb)    

        if !th.checkDuplicates || !th.tokens.ExistsForward(st){
            th.timeMutex.Lock()
            defer th.timeMutex.Unlock()

            // Add element to bmap and set it's time.
            th.tokens.InsertForward(st, elem)
            th.tokenTimers[st] = time.Now().Add(ttl)

            log.Info().
                Str("value", fmt.Sprintf("%v", elem)).
                Dur("ttl", ttl).
                Msg("Token added")

            return st
        }
    }
}

// TokenCount returns the amount of tokens.
func (th *TokenHandler[T]) TokenCount() int {
    return th.tokens.Size()
}

// TokenLength returns the length of the tokens.
func (th *TokenHandler[T]) TokenLength() uint {
    return th.tokenLength
}

// TokensMap returns a map with the tokens as the key, expensive.
func (th *TokenHandler[T]) TokensMap() map[string]T {
    return th.tokens.GetForwardMap()
}

// ElementsMap returns a map with the elements as the key, expensive.
func (th *TokenHandler[T]) ElementsMap() map[T]string {
    return th.tokens.GetBackwardMap()
}

// cleaner periodically removes all expired tokens.
func (th *TokenHandler[T]) cleaner(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    // Indicates when the token system is finished.
    th.wg.Add(1)
    defer th.wg.Done()

    for {
        select{
        case <-ticker.C:
            log.Info().Msg("Token cleaner running")

            th.timeMutex.Lock()

            // Must be calculated after mutex has been acquired, otherwise it could be out of date when access is granted.
            now := time.Now()

            for tok, time := range th.tokenTimers {
                // If a token has expired it's removed from both the token bi-map and the timer map.
                if now.After(time) {
                    val, _ := th.tokens.GetForward(tok)
                    log.Info().
                        Str("value", fmt.Sprintf("%v", val)).
                        Msg("Token expired")

                    th.tokens.DeleteForward(tok)
                    delete(th.tokenTimers, tok)
                }
            }

            th.timeMutex.Unlock()

        case <-ctx.Done():
            log.Info().Msg("Token cleaner stopping")

            // Stop routine if context is cancelled.
            return
        }
    }
}

// Wait blocks until the token handler is done.
func (th *TokenHandler[T]) Wait() {
    th.wg.Wait()
}