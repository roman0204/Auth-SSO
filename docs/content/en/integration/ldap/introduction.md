---
title: "LDAP"
description: "An introduction into integrating Authelia with LDAP."
lead: "An introduction into integrating Authelia with LDAP."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "ldap"
weight: 710
toc: true
---

## UNDER CONSTRUCTION

This section is still a work in progress.

## Configuration

### OpenLDAP

**Tested:**
* Version: [v2.5.13](https://www.openldap.org/software/release/announce_lts.html)
* Container `bitnami/openldap:2.5.13-debian-11-r7`

Create within OpenLDAP, either via CLI or with a GUI management application like
[phpLDAPadmin](http://phpldapadmin.sourceforge.net/wiki/index.php/Main_Page) or [LDAP Admin](http://www.ldapadmin.org/)
a basic user with a complex password.

*Make note of its CN.* You can also create a group to use within Authelia if you would like granular control of who can
login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://OpenLDAP:1389` - servers dns name & port.
  *tip: if you have Authelia on a container network that is routable, you can just use the container name*
* server_name `ldap01.example.com` - servers name
* base_dn `dc=example,dc=com` - common name of domain root.
* groups_filter `dc=example,dc=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml
  ldap:
    implementation: custom
    url: ldap://OpenLDAP:1389
    timeout: 5s
    start_tls: false
    tls:
      server_name: ldap01.example.com
      skip_verify: true
      minimum_version: TLS1.2
    base_dn: dc=example,dc=com
    additional_users_dn: ou=users
    users_filter: (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    username_attribute: uid
    mail_attribute: mail
    display_name_attribute: displayName
    additional_groups_dn: ou=groups
    groups_filter: (&(member=uid={input},ou=users,dc=example,dc=com)(objectclass=groupofnames))
    group_name_attribute: cn
    user: uid=authelia,ou=service accounts,dc=example,dc=com
    password: "SUPER_COMPLEX_PASSWORD"
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

### FreeIPA

**Tested:**
* Version: [v4.9.9](https://www.freeipa.org/page/Releases/4.9.9)
* Container: `freeipa/freeipa-server:fedora-36-4.9.9`

Create within FreeIPA, either via CLI or within its GUI management application `https://server_ip` a basic user with a
complex password.

*Make note of its CN.* You can also create a group to use within Authelia if you would like granular control of who can
login, and reference it within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://ldap` - servers dns name. Port will assume 389 as standard. Specify custom port with `:port` if needed.
* server_name `ldap01.example.com` - servers name
* base_dn `dc=example,dc=com` - common name of domain root.
* groups_filter `dc=example,dc=com` - replace relevant section with your own domain in common name format, same as base_dn.
* user `authelia` - username for Authelia service account
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account

```yaml
 ldap:
    implementation: custom
    url: ldaps://ldap.example.com
    timeout: 5s
    start_tls: false
    tls:
      server_name: ldap.example.com
      skip_verify: true
      minimum_version: TLS1.2
    base_dn: dc=example,dc=com
    username_attribute: uid
    additional_users_dn: cn=users,cn=accounts
    users_filter: (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    additional_groups_dn: ou=groups
    groups_filter: (&(member=uid={input},cn=users,cn=accounts,dc=example,dc=com)(objectclass=groupofnames))
    group_name_attribute: cn
    mail_attribute: mail
    display_name_attribute: displayName
    user: uid=authelia,cn=users,cn=accounts,dc=example,dc=com
    password: "SUPER_COMPLEX_PASSWORD"
```
Following this, restart Authelia, and you should be able to begin using LDAP integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

### lldap

**Tested:**
* Version: [v0.4.0](https://github.com/nitnelave/lldap/releases/tag/v0.4.07)

Create within lldap, a basic user with a complex password, and add to the group "lldap_password_manager"
You can also create a group to use within Authelia if you would like granular control of who can login, and reference it
within the filters below.

### Authelia

In your Authelia configuration you will need to enter and update the following variables -
* url `ldap://OpenLDAP:1389` - servers dns name & port.
  *tip: if you have Authelia on a container network that is routable, you can just use the container name*
* base_dn `dc=example,dc=com` - common name of domain root.
* user `authelia` - username for Authelia service account.
* password `SUPER_COMPLEX_PASSWORD` - password for Authelia service account,

```yaml
ldap:
    implementation: custom
    url: ldap://lldap:3890
    timeout: 5s
    start_tls: false
    base_dn: dc=example,dc=com
    username_attribute: uid
    additional_users_dn: ou=people
    # To allow sign in both with username and email, one can use a filter like
    # (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    users_filter: (&({username_attribute}={input})(objectClass=person))
    additional_groups_dn: ou=groups
    groups_filter: (member={dn})
    group_name_attribute: cn
    mail_attribute: mail
    display_name_attribute: displayName
    # The username and password of the admin or service user.
    user: uid=authelia,ou=people,dc=example,dc=com
    password: "SUPER_COMPLEX_PASSWORD"
```
Following this, restart Authelia, and you should be able to begin using lldap integration for your user logins, with
Authelia taking the email attribute for users straight from the 'mail' attribute within the LDAP object.

## See Also

[Authelia]: https://www.authelia.com
[Bitnami OpenLDAP]: https://hub.docker.com/r/bitnami/openldap/
[FreeIPA]: https://www.freeipa.org/page/Main_Page
[lldap]: https://github.com/nitnelave/lldap