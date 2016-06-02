## Postfix tcp map server

This service desinged to report postfix to reject or not requested address by tcp_table protocol. Service looks through postlog-sa statistics table.

### Desined for

- Postfix
- Postlog-sa

### Configuration specific

Use score section to set reject trigger. `Interval` value sets period in days since now to past and the `limit` is the probable factor of the spam activity.

```
[score]
  interval = 70
  limit = 0.2
```

### How to use with postfix

Add lookup table rule to the postfix main.cf

```
# Permit all connections except local black listed clients
smtpd_client_restrictions = check_client_access tcp:my.server.address:2000

# It's generally polite to say who the mail is from. Again,
# very few real mail do not have a return address, most who don't are spam.
smtpd_sender_restrictions = reject_unknown_address,
                     check_sender_access tcp:my.server.address:2000
```

### See also

Postfix client/server table lookup protocol (http://www.postfix.org/tcp_table.5.html)

