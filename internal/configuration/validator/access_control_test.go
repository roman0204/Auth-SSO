package validator

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type AccessControl struct {
	suite.Suite
	config    *schema.Configuration
	validator *schema.StructValidator
}

func (suite *AccessControl) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = &schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: policyDeny,

			Networks: schema.DefaultACLNetwork,
			Rules:    schema.DefaultACLRule,
		},
	}
}

func (suite *AccessControl) TestShouldValidateCompleteConfiguration() {
	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *AccessControl) TestShouldValidateEitherDomainsOrDomainsRegex() {
	domainsRegex := regexp.MustCompile(`^abc.example.com$`)

	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"abc.example.com"},
			Policy:  "bypass",
		},
		{
			DomainsRegex: []regexp.Regexp{*domainsRegex},
			Policy:       "bypass",
		},
		{
			Policy: "bypass",
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	assert.EqualError(suite.T(), suite.validator.Errors()[0], "access control: rule #3: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidDefaultPolicy() {
	suite.config.AccessControl.DefaultPolicy = testInvalidPolicy

	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: option 'default_policy' must be one of 'bypass', 'one_factor', 'two_factor', 'deny' but it is configured as 'invalid'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetworkGroupNetwork() {
	suite.config.AccessControl.Networks = []schema.ACLNetwork{
		{
			Name:     "internal",
			Networks: []string{"abc.def.ghi.jkl"},
		},
	}

	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: networks: network group 'internal' is invalid: the network 'abc.def.ghi.jkl' is not a valid IP or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorWithNoRulesDefined() {
	suite.config.AccessControl.Rules = []schema.ACLRule{}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: 'default_policy' option 'deny' is invalid: when no rules are specified it must be 'two_factor' or 'one_factor'")
}

func (suite *AccessControl) TestShouldRaiseWarningWithNoRulesDefined() {
	suite.config.AccessControl.Rules = []schema.ACLRule{}

	suite.config.AccessControl.DefaultPolicy = policyTwoFactor

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Errors(), 0)
	suite.Require().Len(suite.validator.Warnings(), 1)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "access control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func (suite *AccessControl) TestShouldRaiseErrorsWithEmptyRules() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{},
		{
			Policy: "wrong",
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 4)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
	suite.Assert().EqualError(suite.validator.Errors()[1], "access control: rule #1: rule 'policy' option '' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "access control: rule #2: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
	suite.Assert().EqualError(suite.validator.Errors()[3], "access control: rule #2: rule 'policy' option 'wrong' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidPolicy() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  testInvalidPolicy,
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): rule 'policy' option 'invalid' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetwork() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains:  []string{"public.example.com"},
			Policy:   "bypass",
			Networks: []string{"abc.def.ghi.jkl/32"},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): the network 'abc.def.ghi.jkl/32' is not a valid Group Name, IP, or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidMethod() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  "bypass",
			Methods: []string{"GET", "HOP"},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'methods' option 'HOP' is invalid: must be one of 'GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'TRACE', 'CONNECT', 'OPTIONS', 'COPY', 'LOCK', 'MKCOL', 'MOVE', 'PROPFIND', 'PROPPATCH', 'UNLOCK'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidSubject() {
	domains := []string{"public.example.com"}
	subjects := [][]string{{"invalid"}}
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains:  domains,
			Policy:   "bypass",
			Subjects: subjects,
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'subject' option 'invalid' is invalid: must start with 'user:' or 'group:'")
	suite.Assert().EqualError(suite.validator.Errors()[1], fmt.Sprintf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(1, suite.config.AccessControl.Rules[0])))
}

func TestAccessControl(t *testing.T) {
	suite.Run(t, new(AccessControl))
}

func TestShouldReturnCorrectResultsForValidNetworkGroups(t *testing.T) {
	config := schema.AccessControlConfiguration{
		Networks: schema.DefaultACLNetwork,
	}

	validNetwork := IsNetworkGroupValid(config, "internal")
	invalidNetwork := IsNetworkGroupValid(config, loopback)

	assert.True(t, validNetwork)
	assert.False(t, invalidNetwork)
}
