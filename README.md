# simple webmail for Plan9

This repo contains a simple webmail client for my Plan 9
mail server. Effectively, it serves as a web frontend
for upasfs(4) but may expand to support for other mail backends
or OSes in the future.

It authenticates the login credentials against factotum to validate
that the user has access to the mailbox mounted at /mail/fs/mbox,
and then displays the messages paginated with unread messages
first and otherwise by date. Clicking on a subject displays the
message, with text/plain preferred for multipart/alternative messages.

NB. Sending from the web client is not yet implemented.
 