/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

// options collector
type options struct {
	secret   string
	orgName  string
	caID     string
	caName   string
	profile  string
	label    string
	typ      string
	attrReqs []*AttributeRequest
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*options) error

// WithOrg option
func WithOrg(orgName string) ClientOption {
	return func(o *options) error {
		o.orgName = orgName
		return nil
	}
}

// WithCAInstance option
func WithCAInstance(caID string) ClientOption {
	return func(o *options) error {
		o.caID = caID
		return nil
	}
}

// RequestOption func for each Opts argument
type RequestOption func(*options) error

// WithCA allows for specifying optional CA name (within the CA server instance)
func WithCA(caName string) RequestOption {
	return func(o *options) error {
		o.caName = caName
		return nil
	}
}

// EnrollmentOption describes a functional parameter for Enroll
type EnrollmentOption func(*options) error

// WithSecret enrollment option
func WithSecret(secret string) EnrollmentOption {
	return func(o *options) error {
		o.secret = secret
		return nil
	}
}

// WithCAName enrollment option
func WithCAName(caName string) EnrollmentOption {
	return func(o *options) error {
		o.caName = caName
		return nil
	}
}

// WithProfile enrollment option
func WithProfile(profile string) EnrollmentOption {
	return func(o *options) error {
		o.profile = profile
		return nil
	}
}

// WithType enrollment option
func WithType(typ string) EnrollmentOption {
	return func(o *options) error {
		o.typ = typ
		return nil
	}
}

// WithLabel enrollment option
func WithLabel(label string) EnrollmentOption {
	return func(o *options) error {
		o.label = label
		return nil
	}
}

// WithAttributeRequests enrollment option
func WithAttributeRequests(attrReqs []*AttributeRequest) EnrollmentOption {
	return func(o *options) error {
		o.attrReqs = attrReqs
		return nil
	}
}
