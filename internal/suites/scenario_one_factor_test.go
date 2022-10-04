package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorSuite struct {
	*RodSuite
}

func New1FAScenario() *OneFactorSuite {
	return &OneFactorSuite{
		RodSuite: new(RodSuite),
	}
}

func (s *OneFactorSuite) SetupSuite() {
	browser, err := StartRod()

	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *OneFactorSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *OneFactorSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *OneFactorSuite) TestShouldAuthorizeSecretAfterOneFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, targetURL)
	s.verifySecretAuthorized(s.T(), s.Page)
}

func (s *OneFactorSuite) TestShouldRedirectToSecondFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "password", false, targetURL)
	s.verifyIsSecondFactorPage(s.T(), s.Context(ctx))
}

func (s *OneFactorSuite) TestShouldDenyAccessOnBadPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), "john", "bad-password", false, targetURL)
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Incorrect username or password.")
}

func TestRunOneFactor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, New1FAScenario())
}
