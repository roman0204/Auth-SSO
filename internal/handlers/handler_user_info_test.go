package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

type FetchSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *FetchSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	// Set the initial user session.
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = 1
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *FetchSuite) TearDownTest() {
	s.mock.Close()
}

type expectedResponse struct {
	db  model.UserInfo
	api *model.UserInfo
	err error
}

type expectedResponseAlt struct {
	description string

	db      model.UserInfo
	api     *model.UserInfo
	loadErr error
	saveErr error
	config  *schema.Configuration
}

func TestUserInfoEndpoint_SetCorrectMethod(t *testing.T) {
	expectedResponses := []expectedResponse{
		{
			db: model.UserInfo{
				Method: "totp",
			},
			err: nil,
		},
		{
			db: model.UserInfo{
				Method:      "webauthn",
				HasWebauthn: true,
				HasTOTP:     true,
			},
			err: nil,
		},
		{
			db: model.UserInfo{
				Method:      "webauthn",
				HasWebauthn: true,
				HasTOTP:     false,
			},
			err: nil,
		},
		{
			db: model.UserInfo{
				Method:      "mobile_push",
				HasWebauthn: false,
				HasTOTP:     false,
			},
			err: nil,
		},
		{
			db:  model.UserInfo{},
			err: sql.ErrNoRows,
		},
		{
			db:  model.UserInfo{},
			err: errors.New("invalid thing"),
		},
	}

	for _, resp := range expectedResponses {
		if resp.api == nil {
			resp.api = &resp.db
		}

		mock := mocks.NewMockAutheliaCtx(t)

		// Set the initial user session.
		userSession := mock.Ctx.GetSession()
		userSession.Username = testUsername
		userSession.AuthenticationLevel = 1
		err := mock.Ctx.SaveSession(userSession)
		require.NoError(t, err)

		mock.StorageMock.
			EXPECT().
			LoadUserInfo(mock.Ctx, gomock.Eq("john")).
			Return(resp.db, resp.err)

		UserInfoGET(mock.Ctx)

		if resp.err == nil {
			t.Run("expected status code", func(t *testing.T) {
				assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
			})

			actualPreferences := model.UserInfo{}

			mock.GetResponseData(t, &actualPreferences)

			t.Run("expected method", func(t *testing.T) {
				assert.Equal(t, resp.api.Method, actualPreferences.Method)
			})

			t.Run("registered webauthn", func(t *testing.T) {
				assert.Equal(t, resp.api.HasWebauthn, actualPreferences.HasWebauthn)
			})

			t.Run("registered totp", func(t *testing.T) {
				assert.Equal(t, resp.api.HasTOTP, actualPreferences.HasTOTP)
			})

			t.Run("registered duo", func(t *testing.T) {
				assert.Equal(t, resp.api.HasDuo, actualPreferences.HasDuo)
			})
		} else {
			t.Run("expected status code", func(t *testing.T) {
				assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
			})

			errResponse := mock.GetResponseError(t)

			assert.Equal(t, "KO", errResponse.Status)
			assert.Equal(t, "Operation failed.", errResponse.Message)
		}

		mock.Close()
	}
}

func TestUserInfoEndpoint_SetDefaultMethod(t *testing.T) {
	expectedResponses := []expectedResponseAlt{
		{
			description: "should set method to totp by default even when user doesn't have totp configured and no preferred method",
			db: model.UserInfo{
				Method:      "",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      false,
			},
			api: &model.UserInfo{
				Method:      "totp",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      false,
			},
			config:  &schema.Configuration{},
			loadErr: nil,
			saveErr: nil,
		},
		{
			description: "should set method to duo by default when user has duo configured and no preferred method",
			db: model.UserInfo{
				Method:      "",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      true,
			},
			api: &model.UserInfo{
				Method:      "mobile_push",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      true,
			},
			config:  &schema.Configuration{},
			loadErr: nil,
			saveErr: nil,
		},
		{
			description: "should set method to totp by default when user has duo configured and no preferred method but duo is not enabled",
			db: model.UserInfo{
				Method:      "",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      true,
			},
			api: &model.UserInfo{
				Method:      "totp",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      true,
			},
			config:  &schema.Configuration{DuoAPI: schema.DuoAPIConfiguration{Disable: true}},
			loadErr: nil,
			saveErr: nil,
		},
		{
			description: "should set method to duo by default when user has duo configured and no preferred method",
			db: model.UserInfo{
				Method:      "",
				HasTOTP:     true,
				HasWebauthn: true,
				HasDuo:      true,
			},
			api: &model.UserInfo{
				Method:      "webauthn",
				HasTOTP:     true,
				HasWebauthn: true,
				HasDuo:      true,
			},
			config: &schema.Configuration{
				TOTP: schema.TOTPConfiguration{
					Disable: true,
				},
			},
			loadErr: nil,
			saveErr: nil,
		},
		{
			description: "should default new users to totp if all enabled",
			db: model.UserInfo{
				Method:      "",
				HasTOTP:     false,
				HasWebauthn: false,
				HasDuo:      false,
			},
			api: &model.UserInfo{
				Method:      "totp",
				HasTOTP:     true,
				HasWebauthn: true,
				HasDuo:      true,
			},
			config:  &schema.Configuration{},
			loadErr: nil,
			saveErr: errors.New("could not save"),
		},
	}

	for _, resp := range expectedResponses {
		t.Run(resp.description, func(t *testing.T) {
			if resp.api == nil {
				resp.api = &resp.db
			}

			mock := mocks.NewMockAutheliaCtx(t)

			if resp.config != nil {
				mock.Ctx.Configuration = *resp.config
			}

			// Set the initial user session.
			userSession := mock.Ctx.GetSession()
			userSession.Username = testUsername
			userSession.AuthenticationLevel = 1
			err := mock.Ctx.SaveSession(userSession)
			require.NoError(t, err)

			if resp.db.Method == "" {
				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadPreferred2FAMethod(mock.Ctx, gomock.Eq("john")).
						Return("", sql.ErrNoRows),
					mock.StorageMock.
						EXPECT().
						SavePreferred2FAMethod(mock.Ctx, gomock.Eq("john"), gomock.Eq("")).
						Return(resp.saveErr),
					mock.StorageMock.
						EXPECT().
						LoadUserInfo(mock.Ctx, gomock.Eq("john")).
						Return(resp.db, nil),
					mock.StorageMock.EXPECT().
						SavePreferred2FAMethod(mock.Ctx, gomock.Eq("john"), gomock.Eq(resp.api.Method)).
						Return(resp.saveErr),
				)
			} else {
				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadPreferred2FAMethod(mock.Ctx, gomock.Eq("john")).
						Return(resp.db.Method, nil),
					mock.StorageMock.
						EXPECT().
						LoadUserInfo(mock.Ctx, gomock.Eq("john")).
						Return(resp.db, nil),
					mock.StorageMock.EXPECT().
						SavePreferred2FAMethod(mock.Ctx, gomock.Eq("john"), gomock.Eq(resp.api.Method)).
						Return(resp.saveErr),
				)
			}

			UserInfoPOST(mock.Ctx)

			if resp.loadErr == nil && resp.saveErr == nil {
				t.Run(fmt.Sprintf("%s/%s", resp.description, "expected status code"), func(t *testing.T) {
					assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
				})

				actualPreferences := model.UserInfo{}

				mock.GetResponseData(t, &actualPreferences)

				t.Run("expected method", func(t *testing.T) {
					assert.Equal(t, resp.api.Method, actualPreferences.Method)
				})

				t.Run("registered webauthn", func(t *testing.T) {
					assert.Equal(t, resp.api.HasWebauthn, actualPreferences.HasWebauthn)
				})

				t.Run("registered totp", func(t *testing.T) {
					assert.Equal(t, resp.api.HasTOTP, actualPreferences.HasTOTP)
				})

				t.Run("registered duo", func(t *testing.T) {
					assert.Equal(t, resp.api.HasDuo, actualPreferences.HasDuo)
				})
			} else {
				t.Run("expected status code", func(t *testing.T) {
					assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
				})

				errResponse := mock.GetResponseError(t)

				assert.Equal(t, "KO", errResponse.Status)
				assert.Equal(t, "Operation failed.", errResponse.Message)
			}

			mock.Close()
		})
	}
}

func (s *FetchSuite) TestShouldReturnError500WhenStorageFailsToLoad() {
	s.mock.StorageMock.EXPECT().
		LoadUserInfo(s.mock.Ctx, gomock.Eq("john")).
		Return(model.UserInfo{}, fmt.Errorf("failure"))

	UserInfoGET(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to load user information: failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func TestFetchSuite(t *testing.T) {
	suite.Run(t, &FetchSuite{})
}

type SaveSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SaveSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	// Set the initial user session.
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = 1
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *SaveSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SaveSuite) TestShouldReturnError500WhenNoBodyProvided() {
	s.mock.Ctx.Request.SetBody(nil)
	MethodPreferencePOST(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenMalformedBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\""))
	MethodPreferencePOST(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"weird_key\":\"abc\"}"))
	MethodPreferencePOST(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to validate body: method: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadMethodProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\"}"))
	MethodPreferencePOST(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unknown or unavailable method 'abc', it should be one of totp, webauthn, mobile_push", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenDatabaseFailsToSave() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"webauthn\"}"))
	s.mock.StorageMock.EXPECT().
		SavePreferred2FAMethod(s.mock.Ctx, gomock.Eq("john"), gomock.Eq("webauthn")).
		Return(fmt.Errorf("Failure"))

	MethodPreferencePOST(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to save new preferred 2FA method: Failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturn200WhenMethodIsSuccessfullySaved() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"webauthn\"}"))
	s.mock.StorageMock.EXPECT().
		SavePreferred2FAMethod(s.mock.Ctx, gomock.Eq("john"), gomock.Eq("webauthn")).
		Return(nil)

	MethodPreferencePOST(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
}

func TestSaveSuite(t *testing.T) {
	suite.Run(t, &SaveSuite{})
}
