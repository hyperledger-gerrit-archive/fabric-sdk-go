/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package handle

import (
	"fmt"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/cachebridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/miekg/pkcs11"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/core")
var ctxCache *lazycache.Cache
var once sync.Once

//LoadPKCS11ContextHandle loads PKCS11 context handler instance from underlying cache
func LoadPKCS11ContextHandle(lib, label, pin string, opts ...Options) (*ContextHandle, error) {
	return getInstance(&pkcs11CtxCacheKey{lib: lib, label: label, pin: pin, opts: getCtxOpts(opts...)}, false)
}

//ReloadPKCS11ContextHandle deletes PKCS11 instance from underlying cache and loads new PKCS11 context  handler in cache
func ReloadPKCS11ContextHandle(lib, label, pin string, opts ...Options) (*ContextHandle, error) {
	return getInstance(&pkcs11CtxCacheKey{lib: lib, label: label, pin: pin, opts: getCtxOpts(opts...)}, true)
}

//ContextHandle encapsulate basic pkcs11.Ctx operations and manages sessions
type ContextHandle struct {
	ctx      *pkcs11.Ctx
	slot     uint
	pin      string
	lib      string
	label    string
	sessions chan pkcs11.SessionHandle
	opts     ctxOpts
}

//Context returns underlying pkcs11.ctx instance
//deprecated
//directly using underlying pkcs11.ctx instance is not recommended, instead ContextHandle features should be used
func (handle *ContextHandle) Context() *pkcs11.Ctx {
	return handle.ctx
}

//OpenSession opens a session between an application and a token.
func (handle *ContextHandle) OpenSession() (pkcs11.SessionHandle, error) {
	var session pkcs11.SessionHandle
	var err error
	for i := 0; i < handle.opts.openSessionRetry; i++ {
		session, err = handle.ctx.OpenSession(handle.slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
		if err != nil {
			logger.Warnf("OpenSession failed, retrying [%s]\n", err)
		} else {
			logger.Debug("OpenSession succeeded")
			break
		}
	}
	return session, err
}

// Login logs a user into a token
func (handle *ContextHandle) Login(session pkcs11.SessionHandle) error {
	if handle.pin == "" {
		return errors.New("No PIN set")
	}
	err := handle.ctx.Login(session, pkcs11.CKU_USER, handle.pin)
	if err != nil && err != pkcs11.Error(pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
		return errors.Errorf("Login failed [%s]", err)
	}
	return nil
}

//ReturnSession returns session back into the session pool
//if pool is pull or session is invalid then discards session
func (handle *ContextHandle) ReturnSession(session pkcs11.SessionHandle) {

	_, e := handle.ctx.GetSessionInfo(session)
	if e != nil {
		logger.Warnf("not returning session [%d], due to error [%s]. Discarding it", session, e)
		e = handle.ctx.CloseSession(session)
		if e != nil {
			logger.Warn("unable to close session:", e)
		}
		return
	}

	select {
	case handle.sessions <- session:
		// returned session back to session cache
	default:
		// have plenty of sessions in cache, dropping
		e = handle.ctx.CloseSession(session)
		if e != nil {
			logger.Warn("unable to close session: ", e)
		}
	}
}

//GetSession returns session from session pool
//if pool is empty or completely in use, creates new session
//if new session is invalid recreates one after reloading ctx and re-login
func (handle *ContextHandle) GetSession() (session pkcs11.SessionHandle) {
	select {
	case session = <-handle.sessions:
		logger.Debugf("Reusing existing pkcs11 session %+v on slot %d\n", session, handle.slot)

	default:

		// cache is empty (or completely in use), create a new session
		s, err := handle.OpenSession()
		if err != nil {
			panic(fmt.Errorf("OpenSession failed [%s]", err))
		}
		logger.Debugf("Created new pkcs11 session %+v on slot %d\n", s, handle.slot)
		session = s
		cachebridge.ClearSession(fmt.Sprintf("%d", session))
	}

	return handle.validateSession(session)
}

//validateSession validates given session
//if session is invalid recreates one after reloading ctx and re-login
func (handle *ContextHandle) validateSession(currentSession pkcs11.SessionHandle) pkcs11.SessionHandle {

	_, e := handle.ctx.GetSessionInfo(currentSession)

	switch e {
	case pkcs11.Error(pkcs11.CKR_OBJECT_HANDLE_INVALID),
		pkcs11.Error(pkcs11.CKR_SESSION_HANDLE_INVALID),
		pkcs11.Error(pkcs11.CKR_SESSION_CLOSED),
		pkcs11.Error(pkcs11.CKR_TOKEN_NOT_PRESENT),
		pkcs11.Error(pkcs11.CKR_DEVICE_ERROR),
		pkcs11.Error(pkcs11.CKR_GENERAL_ERROR),
		pkcs11.Error(pkcs11.CKR_USER_NOT_LOGGED_IN):

		logger.Warnf("Found error condition [%s], attempting to recreate session and re-login....", e)

		handle.disposePKCS11Ctx()

		//create new context
		newCtx := handle.createNewPKCS11Ctx()
		if newCtx == nil {
			logger.Warn("Failed to recreate new context for given library")
			return currentSession
		}

		//find slot
		slot, found := handle.findSlot(newCtx)
		if !found {
			logger.Warnf("Unable to find slot for label :%s", handle.label)
			return currentSession
		}
		logger.Debug("got the slot ", slot)

		//open new session for given slot
		newSession, err := createNewSession(newCtx, slot)
		if err != nil {
			logger.Fatalf("OpenSession [%s]\n", err)
		}
		logger.Debugf("Recreated new pkcs11 session %+v on slot %d\n", newSession, slot)

		//login with new session
		err = newCtx.Login(newSession, pkcs11.CKU_USER, handle.pin)
		if err != nil && err != pkcs11.Error(pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
			logger.Warnf("Unable to login with new session :%s", newSession)
			return currentSession
		}

		handle.ctx = newCtx
		handle.slot = slot
		handle.sessions = make(chan pkcs11.SessionHandle, handle.opts.sessionCacheSize)

		logger.Infof("Able to login with recreated session successfully")
		return newSession

	case pkcs11.Error(pkcs11.CKR_DEVICE_MEMORY),
		pkcs11.Error(pkcs11.CKR_DEVICE_REMOVED):
		panic(fmt.Sprintf("PKCS11 Session failure: [%s]", e))

	default:
		// default should be a valid session or valid error, return session as it is
		return currentSession
	}
}

//disposePKCS11Ctx disposes pkcs11.Ctx object
func (handle *ContextHandle) disposePKCS11Ctx() {
	//ignore error on close all sessions
	err := handle.ctx.CloseAllSessions(handle.slot)
	if err != nil {
		logger.Warnf("Unable to close session", err)
	}

	//clear cache
	cachebridge.ClearAllSession()

	//Destroy context
	handle.ctx.Destroy()
}

//createNewPKCS11Ctx creates new pkcs11.Ctx
func (handle *ContextHandle) createNewPKCS11Ctx() *pkcs11.Ctx {
	newCtx := pkcs11.New(handle.lib)
	if newCtx == nil {
		logger.Warn("Failed to recreate new context for given library")
		return nil
	}

	//initialize new context
	err := newCtx.Initialize()
	if err != nil {
		logger.Warn("Failed to initialize context:", err)
		return nil
	}

	return newCtx
}

//findSlot finds slot for given pkcs11 ctx and label
func (handle *ContextHandle) findSlot(ctx *pkcs11.Ctx) (uint, bool) {

	var found bool
	var slot uint

	//get all slots
	slots, err := ctx.GetSlotList(true)
	if err != nil {
		logger.Warn("Failed to get slot list for recreated context:", err)
		return slot, found
	}

	//find slot matching label
	for _, s := range slots {
		info, err := ctx.GetTokenInfo(s)
		if err != nil {
			continue
		}
		logger.Debugf("Looking for %s, found label %s\n", handle.label, info.Label)
		if handle.label == info.Label {
			found = true
			slot = s
			break
		}
	}

	return slot, found
}

func createNewSession(ctx *pkcs11.Ctx, slot uint) (pkcs11.SessionHandle, error) {
	var newSession pkcs11.SessionHandle
	var err error
	for i := 0; i < 10; i++ {
		newSession, err = ctx.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
		if err != nil {
			logger.Warnf("OpenSession failed, retrying [%s]\n", err)
		} else {
			return newSession, nil
		}
	}
	return newSession, err
}

// pkcs11CtxCacheKey for context handler
type pkcs11CtxCacheKey struct {
	lib   string
	label string
	pin   string
	opts  ctxOpts
}

//String return string value for pkcs11CtxCacheKey
func (key *pkcs11CtxCacheKey) String() string {
	return fmt.Sprintf("%x_%s_%s_%d_%d", key.lib, key.label, key.pin, key.opts.sessionCacheSize, key.opts.openSessionRetry)
}

//getInstance loads ContextHandle instance from cache
//key - cache key
//reload - if true then deletes the existing cache instance and recreates one
func getInstance(key lazycache.Key, reload bool) (*ContextHandle, error) {

	once.Do(func() {
		ctxCache = newCtxCache()
		//anyway, loading first time, no need to reload
		reload = false
	})

	if reload {
		ctxCache.Delete(key)
	}

	ref, err := ctxCache.Get(key)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pkcs11 ctx cache for given key")
	}

	clientRef := ref.(*lazyref.Reference)
	client, err := clientRef.Get()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pkcs11 ctx cache for given key")
	}
	return client.(*ContextHandle), nil
}

//newCtxCache creates new lazycache instance of context handle cache
func newCtxCache() *lazycache.Cache {
	initializer := func(key lazycache.Key) (interface{}, error) {
		ck, ok := key.(*pkcs11CtxCacheKey)
		if !ok {
			return nil, errors.New("unexpected cache key")
		}
		return lazyref.New(
			loadLib(ck.lib, ck.label, ck.pin, ck.opts),
			lazyref.WithFinalizer(finalizer()),
		), nil
	}
	return lazycache.New("Client_Cache", initializer)
}

//finalizer finalizer for context handler cache
func finalizer() lazyref.Finalizer {
	return func(v interface{}) {
		if handle, ok := v.(*ContextHandle); ok {
			err := handle.ctx.CloseAllSessions(handle.slot)
			if err != nil {
				logger.Warnf("unable to close all sessions in finalizer for [%s, %s] : %s", handle.lib, handle.label, err)
			}
			handle.ctx.Destroy()
			cachebridge.ClearAllSession()
		}
	}
}

//loadLib initializer for context handler
func loadLib(lib, label, pin string, opts ctxOpts) lazyref.Initializer {
	return func() (interface{}, error) {

		var slot uint
		logger.Debugf("Loading pkcs11 library [%s]\n", lib)
		if lib == "" {
			return &ContextHandle{}, errors.New("No PKCS11 library default")
		}

		ctx := pkcs11.New(lib)
		if ctx == nil {
			return &ContextHandle{}, errors.Errorf("Instantiate failed [%s]", lib)
		}

		ctx.Initialize()
		//if err != nil {
		//	return &ContextHandle{}, errors.WithMessage(err, "Failed to initialize pkcs11 ctx")
		//}

		slots, err := ctx.GetSlotList(true)
		if err != nil {
			return &ContextHandle{}, errors.WithMessage(err, "Could not get Slot List")
		}
		found := false
		for _, s := range slots {
			info, errToken := ctx.GetTokenInfo(s)
			if errToken != nil {
				continue
			}
			logger.Debugf("Looking for %s, found label %s\n", label, info.Label)
			if label == info.Label {
				found = true
				slot = s
				break
			}
		}
		if !found {
			return &ContextHandle{}, errors.Errorf("Could not find token with label %s", label)
		}
		sessions := make(chan pkcs11.SessionHandle, opts.sessionCacheSize)
		return &ContextHandle{ctx: ctx, slot: slot, pin: pin, lib: lib, label: label, sessions: sessions, opts: opts}, nil
	}
}
