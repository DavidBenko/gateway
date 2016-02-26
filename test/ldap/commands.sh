# Generate encrypted password to go in slapd.conf
slappasswd

# Remove all LDAP data
sudo rm -f /private/var/db/openldap/openldap-data/*

# Copy LDAP server config file
cp ~/platform/gateway/test/ldap/slapd.conf /usr/local/etc/openldap/slapd.conf

# Start LDAP server
sudo /usr/libexec/slapd -d 255 -f /usr/local/etc/openldap/slapd.conf

# Search
ldapsearch -x -D "cn=anypresence.com, dc=anypresence, dc=com" -b dc=anypresence,dc=com -W -s sub "(objectclass=*)"

# Setup data
ldapadd -x -D "cn=anypresence.com, dc=anypresence, dc=com" -W -f ~/platform/gateway/test/ldap/setup.ldif

# Modify entry
ldapmodify -x -D "cn=anypresence.com, dc=anypresence, dc=com" -W -f ./platform/gateway/test/ldap/modify.ldif
