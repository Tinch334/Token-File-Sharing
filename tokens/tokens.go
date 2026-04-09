package tokens


import (
    "crypto/rand"

    "github.com/Tinch334/bmap"
)


type token string


type TokenHandler[T comparable] struct {
    tokenLength uint
    checkDuplicates bool

    tokens *bmap.BMap[string, T]
}

// NewTokenHandler creates a new, empty token handler.
// A recommended token length to avoid collisions with reasonable amounts of tokens is 64.
func NewTokenHandler[T comparable](tLen uint, cDup bool) *TokenHandler[T] {
    return &TokenHandler[T]{
        tokenLength:     tLen,
        checkDuplicates: cDup,
        tokens:          bmap.NewBMap[string, T](),
    }
}

// GenerateToken generates a new token for the given element, the generated token is returned.
// Checks for duplicated tokens if instructed.
func (th *TokenHandler[T]) GenerateToken(elem T) string {
    sb := make([]byte, th.tokenLength)

    for {
        // Read cryptographically secure byte string, note that Read cannot fail.
        rand.Read(sb)
        st := string(sb)    

        if !th.checkDuplicates || !th.tokens.ExistsForward(st){
            th.tokens.InsertForward(st, elem)
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

// TokensMap returns a map with the tokens as the key.
func (th *TokenHandler[T]) TokensMap() map[string]T {
    return th.tokens.GetForwardMap()
}

// ElementsMap returns a map with the elements as the key.
func (th *TokenHandler[T]) ElementsMap() map[T]string {
    return th.tokens.GetBackwardMap()
}