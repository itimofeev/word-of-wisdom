package pow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log/slog"
	"math/rand"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type ProofOfWork struct {
	challengeSize uint
	rand          *rand.Rand
}

func New(challengeSize uint) *ProofOfWork {
	return &ProofOfWork{
		challengeSize: challengeSize,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *ProofOfWork) MakeChallenge() []byte {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var stringBuilder strings.Builder
	stringBuilder.Grow(int(p.challengeSize))

	for i := 0; i < int(p.challengeSize); i++ {
		stringBuilder.WriteByte(charset[p.rand.Intn(len(charset))])
	}

	return []byte(stringBuilder.String())
}

func (p *ProofOfWork) ValidateSolution(data, nonce []byte, difficulty uint) bool {
	hash := p.generateHash(data, nonce)

	prefix := strings.Repeat("0", int(difficulty)) // todo cache such strings
	return strings.HasPrefix(hash, prefix)
}

func (p *ProofOfWork) generateHash(data []byte, nonce []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	hasher.Write(nonce)
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}

func (p *ProofOfWork) SolveChallenge(data []byte, difficulty uint) []byte {
	nThreads := runtime.NumCPU()

	foundNonce := atomic.Uint64{}
	founcCh := make(chan struct{})

	for i := 0; i < nThreads; i++ {
		threadNumber := uint64(i + 1)
		localNonce := uint64(i + 1)

		// launch nThreads goroutine, because it's cpu bound task and more go-routines will not give any profit
		go func() {
			slog.Debug("go-routine started", "number", threadNumber)
			iterCount := 1
			started := time.Now()

			for {
				if iterCount%1_000_000 == 0 {
					slog.Debug("go-routine at iteration", "number", threadNumber, "iterations", iterCount, "elapsed:", time.Since(started))
				}
				if foundNonce.Load() != 0 {
					slog.Debug("go-routine stopped as nonce found by another go-routine", "number", threadNumber, "iterations", iterCount)
					break
				}

				if p.ValidateSolution(data, uint64ToBytes(localNonce), difficulty) {
					foundNonce.Swap(localNonce)
					close(founcCh)
					slog.Debug("go-routine stopped as nonce found by iterations number", "number", threadNumber, "iterations", iterCount)
					break
				}
				localNonce += threadNumber
				iterCount++
			}
		}()
	}

	<-founcCh

	return uint64ToBytes(foundNonce.Load())
}

func uint64ToBytes(i uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}
