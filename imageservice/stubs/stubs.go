package stubs

import "gopkg.in/h2non/gock.v1"

func GetAccount() []*gock.Response {
    resp := make([]*gock.Response, 0)
    req1 := gock.New("http://imageservice:7777").
        Get("/accounts/10000").
        Reply(200).
        BodyString(`{"imageUrl":"http://test.path"}`)
    resp = append(resp, req1)

    gock.New("http://imageservice:7777").
        Get("/accounts").
        Reply(404).
        BodyString(`{"msg":"Not found"}`)

    gock.New("http://imageservice:7777").
        Get("/accounts/1234").
        Reply(500).
        BodyString(`{"msg":"Invalid range"}`)
   }
