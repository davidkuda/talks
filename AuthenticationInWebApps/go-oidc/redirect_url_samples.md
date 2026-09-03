### Redirect: App -> IdP:

```
http://localhost:8080/realms/master/protocol/openid-connect/auth
    client_id              authn-webapps
    code_challenge         0DsiApkDY7SoPi0MqafpeHwGgGuz1Iv9rn9oqmG4rSo
    code_challenge_method  S256
    nonce                  h0YZI33SIHASWCOO06PfuQ
    redirect_uri           http%3A%2F%2Flocalhost%3A8404%2Fcallback
    response_type          code
    scope                  openid+profile+email
    state                  9oIt7MAf_4QiPmyItUWAog
```



### Redirect: IdP -> App:

```
http://localhost:8404/callback
    code           e1f3ce7e-077f-3bd3-8976-32edeb5f02c8.jNZAeJw6Yh7Rcov6LJ_QjHym.ba00a250-91d4-4f64-93d9-f0f7d25e402d
    iss            http%3A%2F%2Flocalhost%3A8080%2Frealms%2Fmaster
    session_state  jNZAeJw6Yh7Rcov6LJ_QjHym
    state          9oIt7MAf_4QiPmyItUWAog
```
