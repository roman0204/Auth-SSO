package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoFactorSuite struct {
	*RodSuite
}

func New2FAScenario() *TwoFactorSuite {
	return &TwoFactorSuite{
		RodSuite: new(RodSuite),
	}
}

func (s *TwoFactorSuite) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *TwoFactorSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *TwoFactorSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *TwoFactorSuite) TestShouldAuthorizeSecretAfterTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	// Login and register TOTP, logout and login again with 1FA & 2FA.
	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	_ = s.doRegisterAndLogin2FA(s.T(), s.Context(ctx), username, password, false, targetURL)

	// And check if the user is redirected to the secret.
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	// Leave the secret.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// And try to reload it again to check the session is kept.
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func (s *TwoFactorSuite) TestShouldFailTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	// Register TOTP secret and logout.
	s.doRegisterThenLogout(s.T(), s.Context(ctx), testUsername, testPassword)

	wrongPasscode := "123456"

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, "")
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
	s.doEnterOTP(s.T(), s.Context(ctx), wrongPasscode)
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "The one-time password might be wrong")
}

func TestRunTwoFactor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, New2FAScenario())
}
