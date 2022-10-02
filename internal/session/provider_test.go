package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestShouldInitializerSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, NewDefaultUserSession(), session)
}

func TestShouldUpdateSession(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}

	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, _ := provider.GetSession(ctx)

	session.Username = testUsername
	session.AuthenticationLevel = authentication.TwoFactor

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:            testUsername,
		AuthenticationLevel: authentication.TwoFactor,
	}, session)
}

func TestShouldSetSessionAuthenticationLevels(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}

	timeOneFactor := time.Unix(1625048140, 0)
	timeTwoFactor := time.Unix(1625048150, 0)
	timeZeroFactor := time.Unix(0, 0)

	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, _ := provider.GetSession(ctx)

	session.SetOneFactor(timeOneFactor, &authentication.UserDetails{Username: testUsername}, false)

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	authAt, err := session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeZeroFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	assert.Equal(t, UserSession{
		Username:                  testUsername,
		AuthenticationLevel:       authentication.OneFactor,
		LastActivity:              timeOneFactor.Unix(),
		FirstFactorAuthnTimestamp: timeOneFactor.Unix(),
		AuthenticationMethodRefs:  oidc.AuthenticationMethodsReferences{UsernameAndPassword: true},
	}, session)

	session.SetTwoFactorDuo(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, UserSession{
		Username:                   testUsername,
		AuthenticationLevel:        authentication.TwoFactor,
		LastActivity:               timeTwoFactor.Unix(),
		FirstFactorAuthnTimestamp:  timeOneFactor.Unix(),
		SecondFactorAuthnTimestamp: timeTwoFactor.Unix(),
		AuthenticationMethodRefs:   oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Duo: true},
	}, session)

	authAt, err = session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeTwoFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)
}

func TestShouldSetSessionAuthenticationLevelsAMR(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}

	timeOneFactor := time.Unix(1625048140, 0)
	timeTwoFactor := time.Unix(1625048150, 0)
	timeZeroFactor := time.Unix(0, 0)

	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, _ := provider.GetSession(ctx)

	session.SetOneFactor(timeOneFactor, &authentication.UserDetails{Username: testUsername}, false)

	err := provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	authAt, err := session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeZeroFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	assert.Equal(t, UserSession{
		Username:                  testUsername,
		AuthenticationLevel:       authentication.OneFactor,
		LastActivity:              timeOneFactor.Unix(),
		FirstFactorAuthnTimestamp: timeOneFactor.Unix(),
		AuthenticationMethodRefs:  oidc.AuthenticationMethodsReferences{UsernameAndPassword: true},
	}, session)

	session.SetTwoFactorWebauthn(timeTwoFactor, false, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t, oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true}, session.AuthenticationMethodRefs)
	assert.True(t, session.AuthenticationMethodRefs.MultiFactorAuthentication())

	authAt, err = session.AuthenticatedTime(authorization.OneFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeOneFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.TwoFactor)
	assert.NoError(t, err)
	assert.Equal(t, timeTwoFactor, authAt)

	authAt, err = session.AuthenticatedTime(authorization.Denied)
	assert.EqualError(t, err, "invalid authorization level")
	assert.Equal(t, timeZeroFactor, authAt)

	session.SetTwoFactorWebauthn(timeTwoFactor, false, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebauthn(timeTwoFactor, false, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebauthn(timeTwoFactor, true, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true, WebauthnUserPresence: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebauthn(timeTwoFactor, true, false)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true, WebauthnUserPresence: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebauthn(timeTwoFactor, false, true)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true, WebauthnUserVerified: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorWebauthn(timeTwoFactor, false, true)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, Webauthn: true, WebauthnUserVerified: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorTOTP(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, TOTP: true, Webauthn: true, WebauthnUserVerified: true},
		session.AuthenticationMethodRefs)

	session.SetTwoFactorTOTP(timeTwoFactor)

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	session, err = provider.GetSession(ctx)
	require.NoError(t, err)

	assert.Equal(t,
		oidc.AuthenticationMethodsReferences{UsernameAndPassword: true, TOTP: true, Webauthn: true, WebauthnUserVerified: true},
		session.AuthenticationMethodRefs)
}

func TestShouldDestroySessionAndWipeSessionData(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	provider := NewProvider(configuration, nil)
	session, err := provider.GetSession(ctx)
	require.NoError(t, err)

	session.Username = testUsername
	session.AuthenticationLevel = authentication.TwoFactor

	err = provider.SaveSession(ctx, session)
	require.NoError(t, err)

	newUserSession, err := provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUsername, newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	err = provider.DestroySession(ctx)
	require.NoError(t, err)

	newUserSession, err = provider.GetSession(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}
