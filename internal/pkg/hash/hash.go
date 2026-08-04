package hash

import (
	"errors"
	"fmt"
	"math"

	"github.com/sqids/sqids-go"
)

const minIDLength = 5

// ErrInvalid is returned when a hash string or ID cannot be processed.
var ErrInvalid = errors.New("invalid hash")

type Hash struct {
	sqids *sqids.Sqids
}

func NewHash(alphabet string) (*Hash, error) {
	s, err := sqids.New(sqids.Options{
		Alphabet:  alphabet,
		MinLength: minIDLength,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Hash: %w", err)
	}

	return &Hash{sqids: s}, nil
}

func (h *Hash) Encode(id int64) (string, error) {
	if id <= 0 {
		return "", ErrInvalid
	}

	return h.sqids.Encode([]uint64{uint64(id)})
}

func (h *Hash) EncodeBatch(ids []int64) (string, error) {
	if len(ids) == 0 {
		return "", ErrInvalid
	}

	uint64s := make([]uint64, len(ids))
	for i, id := range ids {
		if id <= 0 {
			return "", ErrInvalid
		}
		uint64s[i] = uint64(id)
	}

	return h.sqids.Encode(uint64s)
}

func (h *Hash) Decode(encodedID string) ([]int64, error) {
	result := h.sqids.Decode(encodedID)
	if len(result) == 0 {
		return nil, ErrInvalid
	}

	canonical, err := h.sqids.Encode(result)
	if err != nil || canonical != encodedID {
		return nil, ErrInvalid
	}

	int64s := make([]int64, len(result))
	for i, v := range result {
		// Guard against 0 or overflow when converting uint64 to int64
		if v == 0 || v > math.MaxInt64 {
			return nil, ErrInvalid
		}
		int64s[i] = int64(v)
	}

	return int64s, nil
}
