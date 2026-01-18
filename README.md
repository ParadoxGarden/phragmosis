# phragmosis

SSO portal

designed to not need an LDAP server, just allow / deny based on a list

takes the user management lifecycle entirely out of your hands

phragmosis is currently a WIP!

## currently tested features:
forward auth with caddy
oauth with discord

## planned features
oauth with atproto
tailnet bypass

### other notes
by default user sessions expire on phragmosis restart, this is intended behavior.

see ./config.json.example and ./Caddyfile.example for examples of setup 

if you want to persist user sessions, generate block & hash keys and put them in ./config.json like ./config.json.example
