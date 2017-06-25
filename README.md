# goblog
Code samples for the Go microservice blog series


### Calling a secured service through the EDGE gateway

#### 1. Obtain an OAUTH token
    curl -s https://acme:acmesecret@192.168.99.100:9999/uaa/oauth/token   -d grant_type=password  -d client_id=acme  -d scope=webshop  -d username=user  -d password=password -k | jq .
                                                                                                                                                                 
    {
       "access_token": "4c0f86dc-b10f-4095-bafe-66be016732fc",
       "token_type": "bearer",
       "refresh_token": "4aa5b066-5b3a-4e06-bf12-a6dbd72f9c61",
       "expires_in": 43188,
       "scope": "webshop"
     }
     
#### 2. Put the token into an env var
    
    export TOKEN=4c0f86dc-b10f-4095-bafe-66be016732fc
    
#### 2. Call using the token

    curl 'http://192.168.99.100:8765/api/accounts/10000' -H  "Authorization: Bearer $TOKEN" -s