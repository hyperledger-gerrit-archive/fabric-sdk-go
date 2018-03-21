/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

// Helpers for simple check of error code

func isCode(err error, code Code) bool {
	s, ok := FromError(err)
	return ok && s.Code == int32(code)
}

// IsOK returns true if error code is OK
func IsOK(err error) bool {
	return isCode(err, OK)
}

// IsUnknown returns true if error code is Unknown
func IsUnknown(err error) bool {
	return isCode(err, Unknown)
}

// IsConnectionFailed returns true if error code is ConnectionFailed
func IsConnectionFailed(err error) bool {
	return isCode(err, ConnectionFailed)
}

// IsEndorsementMismatch returns true if error code is EndorsementMismatch
func IsEndorsementMismatch(err error) bool {
	return isCode(err, EndorsementMismatch)
}

// IsEmptyCert returns true if error code is EmptyCert
func IsEmptyCert(err error) bool {
	return isCode(err, EmptyCert)
}

// IsTimeout returns true if error code is Timeout
func IsTimeout(err error) bool {
	return isCode(err, Timeout)
}

// IsNoPeersFound returns true if error code is NoPeersFound
func IsNoPeersFound(err error) bool {
	return isCode(err, NoPeersFound)
}

// IsMultipleErrors returns true if error code is MultipleErrors
func IsMultipleErrors(err error) bool {
	return isCode(err, MultipleErrors)
}

// IsSignatureVerificationFailed returns true if error code is SignatureVerificationFailed
func IsSignatureVerificationFailed(err error) bool {
	return isCode(err, SignatureVerificationFailed)
}

// IsMissingEndorsement returns true if error code is MissingEndorsement
func IsMissingEndorsement(err error) bool {
	return isCode(err, MissingEndorsement)
}

// IsChaincodeError returns true if error code is ChaincodeError
func IsChaincodeError(err error) bool {
	return isCode(err, ChaincodeError)
}

// IsNoMatchingCertificateAuthorityEntity returns true if error code is NoMatchingCertificateAuthorityEntity
func IsNoMatchingCertificateAuthorityEntity(err error) bool {
	return isCode(err, NoMatchingCertificateAuthorityEntity)
}

// IsNoMatchingPeerEntity returns true if error code is NoMatchingPeerEntity
func IsNoMatchingPeerEntity(err error) bool {
	return isCode(err, NoMatchingPeerEntity)
}

// IsNoMatchingOrdererEntity returns true if error code is NoMatchingOrdererEntity
func IsNoMatchingOrdererEntity(err error) bool {
	return isCode(err, NoMatchingOrdererEntity)
}
