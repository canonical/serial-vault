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

## serial-vault.admin account

The *serial-vault.admin account* command allows you to cache accounts 
from the store in the database

Example:

```
serial-vault.admin account cache
```

## serial-vault.admin client

Use *serial-vault.admin client* command to generate a test serial 
assertion request

Example:
```
serial-vault.admin client -api=IFUyVnlhV0ZzSUZaaGRXeDB787o -brand=thebrand -model=pc -serial=B2011M -url=https://serial-vault/v1/
```

## serial-vault.admin database

The *serial-vault.admin database* command creates or updates database tables
and relations. Though this is executed just after service startup, this way can be also
executed on demand

Example:

```
serial-vault.admin database 
```

## serial-vault.admin user

Use *serial-vault.admin user* to manage any operation related with 
Serial Vault users. You can add, list, delete or update users

Some examples:

```
serial-vault.admin user list
serial-vault.admin user add somenickname -n User -r superuser
serial-vault.admin user update somenickname -n NewName
serial-vault.admin user delete somenickname
```
