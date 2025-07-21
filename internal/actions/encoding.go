package actions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// base64EncodeAction encodes data to base64
// Args: [data] - data to encode
func base64EncodeAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("base64_encode", 1, len(args))
	}

	data := fmt.Sprintf("%v", args[0])
	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   encoded,
	}
}

// base64DecodeAction decodes base64 data
// Args: [encoded_data] - base64 encoded data to decode
func base64DecodeAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("base64_decode", 1, len(args))
	}

	encodedData := fmt.Sprintf("%v", args[0])
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_BASE64").
			WithTemplate("Invalid base64 data for decoding").
			WithContext("encoded_data", encodedData).
			WithContext("error", err.Error()).
			WithSuggestion("Ensure the input is valid base64 encoded data").
			Build(fmt.Sprintf("base64 decode error: %s", err.Error()))
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   string(decoded),
	}
}

// urlEncodeAction URL encodes data
// Args: [data] - data to URL encode
func urlEncodeAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("url_encode", 1, len(args))
	}

	data := fmt.Sprintf("%v", args[0])
	encoded := url.QueryEscape(data)

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   encoded,
	}
}

// urlDecodeAction URL decodes data
// Args: [encoded_data] - URL encoded data to decode
func urlDecodeAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("url_decode", 1, len(args))
	}

	encodedData := fmt.Sprintf("%v", args[0])
	decoded, err := url.QueryUnescape(encodedData)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_URL_ENCODING").
			WithTemplate("Invalid URL encoded data for decoding").
			WithContext("encoded_data", encodedData).
			WithContext("error", err.Error()).
			WithSuggestion("Ensure the input is valid URL encoded data").
			Build(fmt.Sprintf("URL decode error: %s", err.Error()))
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   decoded,
	}
}

// hashAction generates hash of data using specified algorithm
// Args: [data, algorithm] - data to hash and algorithm (md5, sha1, sha256, sha512)
func hashAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.MissingArgsError("hash", 2, len(args))
	}

	data := fmt.Sprintf("%v", args[0])
	algorithm := strings.ToLower(fmt.Sprintf("%v", args[1]))

	var hash string

	switch algorithm {
	case "md5":
		hash = fmt.Sprintf("%x", md5.Sum([]byte(data)))
	case "sha1":
		hash = fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	case "sha256":
		hash = fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
	case "sha512":
		hash = fmt.Sprintf("%x", sha512.Sum512([]byte(data)))
	default:
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "UNSUPPORTED_HASH_ALGORITHM").
			WithTemplate("Unsupported hash algorithm").
			WithContext("algorithm", algorithm).
			WithContext("supported_algorithms", "md5, sha1, sha256, sha512").
			WithSuggestion("Use one of the supported hash algorithms: md5, sha1, sha256, sha512").
			Build(fmt.Sprintf("unsupported hash algorithm: %s", algorithm))
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"hash":      hash,
			"algorithm": algorithm,
			"input":     data,
		},
	}
}
