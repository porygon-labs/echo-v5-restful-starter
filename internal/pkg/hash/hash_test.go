package hash_test

import (
	"errors"
	"math"
	"slices"
	"testing"

	"go-restful-api/internal/constants"
	"go-restful-api/internal/pkg/hash"

	"github.com/sqids/sqids-go"
)

func TestNewHash(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		alphabet string
		wantErr  bool
	}{
		{
			name:     "success default alphabet",
			alphabet: constants.TEST_SQIDS_ALPHABET,
			wantErr:  false,
		},
		{
			name:     "success minimum 3 chars",
			alphabet: "abc",
			wantErr:  false,
		},
		{
			name:     "failed: repeated text ",
			alphabet: "abcdefga",
			wantErr:  true,
		},
		{
			name:     "failed: alphabet too short",
			alphabet: "a",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := hash.NewHash(tt.alphabet)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewHash() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewHash() succeeded unexpectedly")
			}
		})
	}
}

func TestHash_Encode(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		id      int64
		want    string
		wantErr bool
	}{
		{
			name:    "id 1",
			id:      1,
			want:    "UkLWZ",
			wantErr: false,
		},
		{
			name:    "id 0",
			id:      0,
			want:    "",
			wantErr: true,
		},
		{
			name:    "id -1",
			id:      -1,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := hash.NewHash(constants.TEST_SQIDS_ALPHABET)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := h.Encode(tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Encode() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Encode() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_EncodeBatch(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		ids     []int64
		want    string
		wantErr bool
	}{
		{
			name:    "empty ID",
			want:    "",
			ids:     []int64{},
			wantErr: true,
		},
		{
			name: "success encode > 1 id",
			ids: []int64{
				1, 1,
			},
			want:    "kQKMT",
			wantErr: false,
		},
		{
			name:    "one of the id is invalid",
			wantErr: true,
			ids: []int64{
				1, 0,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := hash.NewHash(constants.TEST_SQIDS_ALPHABET)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := h.EncodeBatch(tt.ids)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("EncodeBatch() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("EncodeBatch() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("EncodeBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_Decode(t *testing.T) {
	sqidsObj, err := sqids.New(sqids.Options{
		Alphabet:  constants.TEST_SQIDS_ALPHABET,
		MinLength: 5,
	})
	if err != nil {
		t.Fatalf("failed to setup test vectors: %v", err)
	}

	validSingleHash, _ := sqidsObj.Encode([]uint64{100})
	zeroValueHash, _ := sqidsObj.Encode([]uint64{0})
	overflowHash, _ := sqidsObj.Encode([]uint64{math.MaxUint64})

	tests := []struct {
		name      string // description of this test case
		alphabet  string // optional custom alphabet override
		encodedID string // input parameter for target function
		want      []int64
		wantErr   bool
	}{
		{
			name:      "succeed - batch IDs",
			encodedID: "kQKMT",
			want:      []int64{1, 1},
			wantErr:   false,
		},
		{
			name:      "succeed - single ID",
			encodedID: validSingleHash,
			want:      []int64{100},
			wantErr:   false,
		},
		{
			name:      "error - empty encoded ID string",
			encodedID: "",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "error - non-canonical / tampered hash",
			encodedID: "kQKMTa",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "error - garbage invalid characters",
			encodedID: "!!!@@@",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "error - decoded value contains 0",
			encodedID: zeroValueHash,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "error - decoded value exceeds int64 max limit",
			encodedID: overflowHash,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alphabet := tt.alphabet
			if alphabet == "" {
				alphabet = constants.TEST_SQIDS_ALPHABET
			}

			h, err := hash.NewHash(alphabet)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := h.Decode(tt.encodedID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Fatalf("Decode() failed unexpectedly: %v", gotErr)
				}
				if !errors.Is(gotErr, hash.ErrInvalid) {
					t.Errorf("Decode() error = %v, wantErr = %v", gotErr, hash.ErrInvalid)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("Decode() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
