// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package paging

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"golang.org/x/crypto/nacl/secretbox"
)

// DecodePageToken takes a base64-encoded, secretbox-encrypted token and returns the searchAfter string.
// Returns an error if decoding, decryption, or unmarshaling fails.
func DecodePageToken(ctx context.Context, encoded string, secretKey *[32]byte) (string, error) {

	slog.DebugContext(ctx, "decoding page token",
		"encoded_token", encoded,
	)

	// Base64 decode.
	encrypted, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("corrupted page token")
	}

	// Safety check.
	if len(encrypted) < constants.NonceSize+secretbox.Overhead {
		return "", errors.New("invalid page token length")
	}

	// Extract nonce and decrypt.
	var decryptNonce [constants.NonceSize]byte
	copy(decryptNonce[:], encrypted[:constants.NonceSize])
	decrypted, ok := secretbox.Open(nil, encrypted[constants.NonceSize:], &decryptNonce, secretKey)
	if !ok {
		return "", errors.New("invalid page token signature")
	}

	// JSON re-marshal to normalize structure.
	searchAfterMsg := json.RawMessage(string(decrypted))
	searchAfterData, err := json.Marshal(searchAfterMsg)
	if err != nil {
		return "", errors.New("malformed page token")
	}

	slog.DebugContext(ctx, "decoded page token successfully",
		"search_after", string(searchAfterData),
	)

	return string(searchAfterData), nil
}

// EncodePageToken recebe um valor JSON serializável (ex: []interface{}, map[string]interface{}, etc),
// encripta com secretbox, e retorna um token seguro em base64.
func EncodePageToken(searchAfter any, secretKey *[32]byte) (string, error) {
	// Serializa para JSON
	encodedSearchAfter, err := json.Marshal(searchAfter)
	if err != nil {
		return "", errors.New("unrecoverable pagination error: failed to encode")
	}

	// Gera nonce aleatório
	var nonce [constants.NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", errors.New("unrecoverable pagination error: failed to generate nonce")
	}

	// Encripta com secretbox
	encrypted := secretbox.Seal(nonce[:], encodedSearchAfter, &nonce, secretKey)

	// Codifica para base64
	token := base64.RawURLEncoding.EncodeToString(encrypted)
	return token, nil
}
