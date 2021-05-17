package common

import (
    "errors"
    "net/url"
    "strconv"
    "strings"
    "time"
    "unicode"
)

const (
    BYTE = 1 << (10 * iota)
    KILOBYTE
    MEGABYTE
    GIGABYTE
    TERABYTE
    PETABYTE
    EXABYTE
)

var invalidByteQuantityError = errors.New("byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB")

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//	E: Exabyte
//	P: Petabyte
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes uint64) string {
    unit := ""
    value := float64(bytes)

    switch {
    case bytes >= EXABYTE:
        unit = "E"
        value = value / EXABYTE
    case bytes >= PETABYTE:
        unit = "P"
        value = value / PETABYTE
    case bytes >= TERABYTE:
        unit = "T"
        value = value / TERABYTE
    case bytes >= GIGABYTE:
        unit = "G"
        value = value / GIGABYTE
    case bytes >= MEGABYTE:
        unit = "M"
        value = value / MEGABYTE
    case bytes >= KILOBYTE:
        unit = "K"
        value = value / KILOBYTE
    case bytes >= BYTE:
        unit = "B"
    case bytes == 0:
        return "0B"
    }

    result := strconv.FormatFloat(value, 'f', 1, 64)
    result = strings.TrimSuffix(result, ".0")
    return result + unit
}

// ToMegabytes parses a string formatted by ByteSize as megabytes.
func ToMegabytes(s string) (uint64, error) {
    bytes, err := ToBytes(s)
    if err != nil {
        return 0, err
    }

    return bytes / MEGABYTE, nil
}

// ToBytes parses a string formatted by ByteSize as bytes. Note binary-prefixed and SI prefixed units both mean a base-2 units
// KB = K = KiB	= 1024
// MB = M = MiB = 1024 * K
// GB = G = GiB = 1024 * M
// TB = T = TiB = 1024 * G
// PB = P = PiB = 1024 * T
// EB = E = EiB = 1024 * P
func ToBytes(s string) (uint64, error) {
    s = strings.TrimSpace(s)
    s = strings.ToUpper(s)

    i := strings.IndexFunc(s, unicode.IsLetter)

    if i == -1 {
        return 0, invalidByteQuantityError
    }

    bytesString, multiple := s[:i], s[i:]
    bytes, err := strconv.ParseFloat(bytesString, 64)
    if err != nil || bytes < 0 {
        return 0, invalidByteQuantityError
    }

    switch multiple {
    case "E", "EB", "EIB":
        return uint64(bytes * EXABYTE), nil
    case "P", "PB", "PIB":
        return uint64(bytes * PETABYTE), nil
    case "T", "TB", "TIB":
        return uint64(bytes * TERABYTE), nil
    case "G", "GB", "GIB":
        return uint64(bytes * GIGABYTE), nil
    case "M", "MB", "MIB":
        return uint64(bytes * MEGABYTE), nil
    case "K", "KB", "KIB":
        return uint64(bytes * KILOBYTE), nil
    case "B":
        return uint64(bytes), nil
    default:
        return 0, invalidByteQuantityError
    }
}

func UrlString(u *url.URL, key string, defaultValue string) string {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    return v
}

func UrlDuration(u *url.URL, key string, defaultValue time.Duration) time.Duration {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    if vv, err := strconv.ParseInt(v, 10, 64); err != nil {
        return defaultValue
    } else {
        return time.Duration(vv)
    }
}
func UrlDurationSecond(u *url.URL, key string, defaultValue time.Duration) time.Duration {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    if vv, err := strconv.ParseInt(v, 10, 64); err != nil {
        return defaultValue
    } else {
        return time.Duration(vv) * time.Second
    }
}

func UrlInt(u *url.URL, key string, defaultValue int) int {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    if vv, err := strconv.ParseInt(v, 10, 64); err != nil {
        return defaultValue
    } else {
        return int(vv)
    }
}

func UrlInt64(u url.URL, key string, defaultValue int64) int64 {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    if vv, err := strconv.ParseInt(v, 10, 64); err != nil {
        return defaultValue
    } else {
        return int64(vv)
    }
}
func UrlFloat(u url.URL, key string, defaultValue float64) float64 {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    if vv, err := strconv.ParseFloat(v, 10); err != nil {
        return defaultValue
    } else {
        return vv
    }
}
func UrlBool(u *url.URL, key string, defaultValue bool) bool {
    v := u.Query().Get(key)
    if v == "" {
        return defaultValue
    }
    return strings.ToLower(v) == "true"
}
