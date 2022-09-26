package validator

import (
	"fmt"
	"net"
	"strings"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return utils.IsStringInSlice(policy, validACLRulePolicies)
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) (isValid bool) {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsNetworkGroupValid check if a network group is valid.
func IsNetworkGroupValid(config schema.AccessControlConfiguration, network string) bool {
	for _, networks := range config.Networks {
		if network != networks.Name {
			continue
		} else {
			return true
		}
	}

	return false
}

// IsNetworkValid check if a network is valid.
func IsNetworkValid(network string) (isValid bool) {
	if net.ParseIP(network) == nil {
		_, _, err := net.ParseCIDR(network)
		return err == nil
	}

	return true
}

func ruleDescriptor(position int, rule schema.ACLRule) string {
	if len(rule.Domains) == 0 {
		return fmt.Sprintf("#%d", position)
	}

	return fmt.Sprintf("#%d (domain '%s')", position, strings.Join(rule.Domains, ","))
}

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(config *schema.Configuration, validator *schema.StructValidator) {
	if config.AccessControl.DefaultPolicy == "" {
		config.AccessControl.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(config.AccessControl.DefaultPolicy) {
		validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyValue, strings.Join(validACLRulePolicies, "', '"), config.AccessControl.DefaultPolicy))
	}

	if config.AccessControl.Networks != nil {
		for _, n := range config.AccessControl.Networks {
			for _, networks := range n.Networks {
				if !IsNetworkValid(networks) {
					validator.Push(fmt.Errorf(errFmtAccessControlNetworkGroupIPCIDRInvalid, n.Name, networks))
				}
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(config *schema.Configuration, validator *schema.StructValidator) {
	if config.AccessControl.Rules == nil || len(config.AccessControl.Rules) == 0 {
		if config.AccessControl.DefaultPolicy != policyOneFactor && config.AccessControl.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyWithoutRules, config.AccessControl.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf(errFmtAccessControlWarnNoRulesDefaultPolicy, config.AccessControl.DefaultPolicy))

		return
	}

	for i, rule := range config.AccessControl.Rules {
		rulePosition := i + 1

		if len(rule.Domains)+len(rule.DomainsRegex) == 0 {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleNoDomains, ruleDescriptor(rulePosition, rule)))
		}

		if !IsPolicyValid(rule.Policy) {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidPolicy, ruleDescriptor(rulePosition, rule), rule.Policy))
		}

		validateNetworks(rulePosition, rule, config.AccessControl, validator)

		validateSubjects(rulePosition, rule, validator)

		validateMethods(rulePosition, rule, validator)

		if rule.Policy == policyBypass {
			validateBypass(rulePosition, rule, validator)
		}
	}
}

func validateBypass(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	if len(rule.Subjects) != 0 {
		validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(rulePosition, rule)))
	}

	for _, pattern := range rule.DomainsRegex {
		if utils.IsStringSliceContainsAny(authorization.IdentitySubexpNames, pattern.SubexpNames()) {
			validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjectsWithGroupDomainRegex, ruleDescriptor(rulePosition, rule)))
			return
		}
	}
}

func validateNetworks(rulePosition int, rule schema.ACLRule, config schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range rule.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(config, network) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleNetworksInvalid, ruleDescriptor(rulePosition, rule), network))
			}
		}
	}
}

func validateSubjects(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleSubjectInvalid, ruleDescriptor(rulePosition, rule), subject))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range rule.Methods {
		if !utils.IsStringInSliceFold(method, validACLHTTPMethodVerbs) {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleMethodInvalid, ruleDescriptor(rulePosition, rule), method, strings.Join(validACLHTTPMethodVerbs, "', '")))
		}
	}
}
