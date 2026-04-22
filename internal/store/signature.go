package store

import (
	"crypto/sha256"
	"encoding/hex"

	"faultline/internal/engine"
	"faultline/internal/model"
	signaturepkg "faultline/internal/signature"
)

type Signature = signaturepkg.ResultSignature

func InputHashForLog(text string) string {
	sum := sha256.Sum256([]byte(engine.CanonicalizeLog(text)))
	return hex.EncodeToString(sum[:])
}

func SignatureForResult(result model.Result) Signature {
	return signaturepkg.ForResult(result)
}

func NormalizeEvidenceLine(line string) string {
	return signaturepkg.NormalizeEvidenceLine(line)
}
