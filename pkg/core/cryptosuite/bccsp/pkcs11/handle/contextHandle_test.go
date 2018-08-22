/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package handle

import (
	"testing"

	"github.com/miekg/pkcs11"
	"github.com/stretchr/testify/assert"
)

const (
	pin   = "98765432"
	label = "ForFabric"
	lib   = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
)

func TestContextHandle(t *testing.T) {

	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	//Test session
	session, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	//Test login
	err = handle.Login(session)
	assert.NoError(t, err)

	//test return/get session
	assert.Equal(t, 0, len(handle.sessions))
	handle.ReturnSession(session)
	assert.Equal(t, 1, len(handle.sessions))
	session = handle.GetSession()
	assert.Equal(t, 0, len(handle.sessions))
	handle.ReturnSession(session)
	assert.Equal(t, 1, len(handle.sessions))

	//add new 2 session to pool, externally
	session1, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	session2, err := handle.OpenSession()
	assert.NoError(t, err)
	assert.True(t, session > 0)

	handle.ReturnSession(session1)
	handle.ReturnSession(session2)

	assert.Equal(t, 3, len(handle.sessions))

	//empty pool
	session1 = handle.GetSession()
	session2 = handle.GetSession()
	session3 := handle.GetSession()

	assert.Equal(t, 0, len(handle.sessions))

	//even if pool is empty should be able to get session
	session4 := handle.GetSession()
	assert.Equal(t, 0, len(handle.sessions))

	//return all sessions to pool
	handle.ReturnSession(session1)
	handle.ReturnSession(session2)
	handle.ReturnSession(session3)
	handle.ReturnSession(session4)
	assert.Equal(t, 4, len(handle.sessions))

}

func TestContextHandleInstance(t *testing.T) {

	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	defer func() {
		//reload pkcs11 context for other tests to succeed
		handle, err := ReloadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		assert.NotNil(t, handle.ctx)
		assert.Equal(t, handle.lib, lib)
		assert.Equal(t, handle.label, label)
		assert.Equal(t, handle.pin, pin)
	}()

	//destroy pkcs11 ctx inside
	handle.ctx.Destroy()

	//test if this impacted other 'LoadPKCS11ContextHandle' calls
	t.Run("test corrupted context handler instance", func(t *testing.T) {

		//get it again
		handle1, err := LoadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handle1)

		//Open session should fail it is destroyed by previous instance
		err = handle1.ctx.CloseAllSessions(handle.slot)
		assert.Error(t, err, pkcs11.CKR_CRYPTOKI_NOT_INITIALIZED)
	})

}

func TestContextHandleOpts(t *testing.T) {

	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin, WithOpenSessionRetry(10), WithSessionCacheSize(2))
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)
	assert.Equal(t, handle.lib, lib)
	assert.Equal(t, handle.label, label)
	assert.Equal(t, handle.pin, pin)

	//get 4 sessions
	session1 := handle.GetSession()
	session2 := handle.GetSession()
	session3 := handle.GetSession()
	session4 := handle.GetSession()

	//return all 4, but pool size is 2, so last 2 will sessions will be closed
	handle.ReturnSession(session1)
	handle.ReturnSession(session2)
	handle.ReturnSession(session3)
	handle.ReturnSession(session4)

	//session1 should be valid
	_, e := handle.ctx.GetSessionInfo(session1)
	assert.NoError(t, e)

	//session2 should be valid
	_, e = handle.ctx.GetSessionInfo(session2)
	assert.NoError(t, e)

	//session3 should be closed
	_, e = handle.ctx.GetSessionInfo(session3)
	assert.Equal(t, pkcs11.Error(pkcs11.CKR_SESSION_HANDLE_INVALID), e)

	//session4 should be closed
	_, e = handle.ctx.GetSessionInfo(session4)
	assert.Equal(t, pkcs11.Error(pkcs11.CKR_SESSION_HANDLE_INVALID), e)
}

func TestContextHandleCommonInstance(t *testing.T) {
	//get context handler
	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	oldCtx := handle.ctx
	for i := 0; i < 20; i++ {
		handleX, err := LoadPKCS11ContextHandle(lib, label, pin)
		assert.NoError(t, err)
		assert.NotNil(t, handleX)
		//Should be same instance, for same set of lib, label, pin
		assert.Equal(t, oldCtx, handleX.ctx)
	}
}

func TestContextRefreshOnInvalidSession(t *testing.T) {

	handle, err := LoadPKCS11ContextHandle(lib, label, pin)
	assert.NoError(t, err)
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.ctx)

	//get session
	session := handle.GetSession()

	//close this session and return it, validation on return session should stop it
	handle.ctx.CloseSession(session)
	handle.ReturnSession(session)
	//session pool unchanged, since returned session was invalid
	assert.Equal(t, 0, len(handle.sessions))

	//just for test manually add it into pool
	handle.sessions <- session
	assert.Equal(t, 1, len(handle.sessions))

	oldCtx := handle.ctx
	assert.Equal(t, oldCtx, handle.ctx)

	//get session again, now ctx should be refreshed
	session = handle.GetSession()
	assert.NotEqual(t, oldCtx, handle.ctx)
	assert.NotNil(t, session)

}
