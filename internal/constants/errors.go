package constants

import (
	"fmt"

	"github.com/BeaverHouse/go-common/errorhandle"
)

// Domain error catalog for ba-torment-data-process.
//
// Each entry is ONE canonical identity per distinct situation. Dynamic detail
// (URL, status code, language, file name, student id) is passed as a parameter
// via fmt.Sprintf — that is not a variant. Generic situations (config missing,
// DB operations, plain internal failures) use go-common errorhandle directly
// and must not be redefined here.
//
// Codes use the DP_ prefix and never embed an HTTP status.

// === KindInternal ===

// ErrDataFetch is a failure to fetch data from an external HTTP source
// (SchaleDB, Supabase, party data, portraits, character images).
func ErrDataFetch(resource string, cause error) error {
	return errorhandle.Wrap(errorhandle.KindInternal, "DP_DATA_FETCH_FAILED",
		fmt.Sprintf("failed to fetch %s", resource), cause)
}

// ErrDataDecode is a failure to decode fetched data (JSON unmarshal, image
// decode) for a given resource.
func ErrDataDecode(resource string, cause error) error {
	return errorhandle.Wrap(errorhandle.KindInternal, "DP_DATA_DECODE_FAILED",
		fmt.Sprintf("failed to decode %s", resource), cause)
}

// ErrDataEncode is a failure to encode data before upload (JSON marshal, image
// encode, multipart build) for a given resource.
func ErrDataEncode(resource string, cause error) error {
	return errorhandle.Wrap(errorhandle.KindInternal, "DP_DATA_ENCODE_FAILED",
		fmt.Sprintf("failed to encode %s", resource), cause)
}

// ErrUpload is a failure to upload a file/result to the File Manager API.
func ErrUpload(resource string, cause error) error {
	return errorhandle.Wrap(errorhandle.KindInternal, "DP_UPLOAD_FAILED",
		fmt.Sprintf("failed to upload %s", resource), cause)
}

// ErrGridGenerate is a failure to generate a student grid image.
func ErrGridGenerate(fileName string, cause error) error {
	return errorhandle.Wrap(errorhandle.KindInternal, "DP_GRID_GENERATE_FAILED",
		fmt.Sprintf("failed to generate grid %s", fileName), cause)
}

// === KindUnavailable ===

// ErrUpstreamBadStatus is an unexpected non-2xx HTTP status from an upstream
// source. detail carries the status code and URL/body.
func ErrUpstreamBadStatus(detail string) error {
	return errorhandle.New(errorhandle.KindUnavailable, "DP_UPSTREAM_BAD_STATUS",
		fmt.Sprintf("unexpected upstream response: %s", detail))
}

// ErrEmptyResponse is an empty body from an upstream source for a URL.
func ErrEmptyResponse(url string) error {
	return errorhandle.New(errorhandle.KindUnavailable, "DP_EMPTY_RESPONSE",
		fmt.Sprintf("empty response body for URL: %s", url))
}

// ErrMaxRetriesExceeded is a retryable request that never succeeded.
func ErrMaxRetriesExceeded(url string) error {
	return errorhandle.New(errorhandle.KindUnavailable, "DP_MAX_RETRIES_EXCEEDED",
		fmt.Sprintf("max retries exceeded for URL: %s", url))
}

// ErrDuckDBUnavailable is a missing or unfetchable DuckDB source file for a
// raid. cause carries the underlying download/stat failure.
func ErrDuckDBUnavailable(cause error) error {
	return errorhandle.Wrap(errorhandle.KindUnavailable, "DP_DUCKDB_UNAVAILABLE",
		"duckdb file not available", cause)
}

// ErrDuckDBTooSmall is a downloaded DuckDB file too small to be valid.
func ErrDuckDBTooSmall(bytes int64) error {
	return errorhandle.New(errorhandle.KindUnavailable, "DP_DUCKDB_TOO_SMALL",
		fmt.Sprintf("downloaded file is too small (%d bytes), likely not a valid DuckDB file", bytes))
}

// === KindNotFound ===

// ErrArmorTypeNoDetails is an armor type that yielded no party details.
func ErrArmorTypeNoDetails(armorType string) error {
	return errorhandle.New(errorhandle.KindNotFound, "DP_ARMOR_TYPE_NO_DETAILS",
		fmt.Sprintf("no details found for armor type: %s", armorType))
}

// === KindInvalid ===

// ErrUnknownBuildValue is a build string with no known weapon-star mapping.
func ErrUnknownBuildValue(build string) error {
	return errorhandle.New(errorhandle.KindInvalid, "DP_UNKNOWN_BUILD_VALUE",
		fmt.Sprintf("unknown build value: %s", build))
}

// === KindFailedPrecondition ===

// ErrNoPointColumns is an elimination raid DuckDB with no point columns.
func ErrNoPointColumns() error {
	return errorhandle.New(errorhandle.KindFailedPrecondition, "DP_NO_POINT_COLUMNS",
		"no point columns found for elimination raid")
}
