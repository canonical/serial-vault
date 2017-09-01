---
title: "Commands"
table_of_contents: True
---

# Admin tool

Along with the snap comes a CLI administration tool. It allows to execute 
several operations that are also available through the web UI. In other cases
they are only available here and even there is a case when it is needed to be 
executed here a command before accessing the UI.
Here they are:

## serial-vault-server.admin account

The *serial-vault-server.admin account* command allows you to cache accounts 
from the store in the database

Example:

```
serial-vault-server.admin account cache
```

## serial-vault-server.admin client

Use *serial-vault-server.admin client* command to generate a test serial 
assertion request

Example:
```
serial-vault-server.admin client -api=IFUyVnlhV0ZzSUZaaGRXeDB787o -brand=thebrand -model=pc -serial=B2011M -url=https://serial-vault/v1/
```

## serial-vault-server.admin database

The *serial-vault-server.admin database* command creates or updates database tables
and relations. Though this is executed just after service startup, this way can be also
executed on demand

Example:

```
serial-vault-server.admin database 
```

## serial-vault-server.admin user

Use *serial-vault-server.admin user* to manage any operation related with 
Serial Vault users. You can add, list, delete or update users

Some examples:

```
serial-vault-server.admin user list
serial-vault-server.admin user add somenickname -n User -r superuser
serial-vault-server.admin user update somenickname -n NewName
serial-vault-server.admin user delete somenickname
```
