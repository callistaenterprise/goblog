package service

import (
    "net/http"

    "github.com/graphql-go/graphql"
    "log"
    "encoding/json"
    "github.com/Sirupsen/logrus"
    internalmodel "github.com/callistaenterprise/goblog/accountservice/model"
    "github.com/callistaenterprise/goblog/common/model"
    "strconv"
    "time"
    "fmt"
)

var Schema graphql.Schema
var schemaInitialized = false
// define custom GraphQL ObjectType `todoType` for our Golang struct `Todo`
// Note that
// - the fields in our todoType maps with the json tags for the fields in our struct
// - the field type matches the field type in our struct

var accounts []internalmodel.Account

func init() {
    accounts = make([]internalmodel.Account, 0)
    for a := 0; a < 10; a++ {
        accountId := strconv.Itoa(120 + a)
        quote := internalmodel.Quote{Text: "HEJ", Language: "sv"}
        account := internalmodel.Account{ID: accountId, Name: "Test Testsson " + strconv.Itoa(a), ServedBy: "localhost", Quote: quote}
        account.AccountEvents = make([]model.AccountEvent, 0)
        account.AccountEvents = append(account.AccountEvents, model.AccountEvent{strconv.Itoa(1), accountId, "CREATED", time.Now().Format("2006-01-02T15:04:05")})
        account.AccountEvents = append(account.AccountEvents, model.AccountEvent{strconv.Itoa(2), accountId, "UPDATED", time.Now().Format("2006-01-02T15:04:05")})

        accounts = append(accounts, account)
    }
}

func initQL() {
    if schemaInitialized {
        return
    }
    // Types
    var quoteType = graphql.NewObject(graphql.ObjectConfig{
        Name: "Quote",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.String,
            },
            "quote": &graphql.Field{
                Type: graphql.String,
            },
            "ipAddress": &graphql.Field{
                Type: graphql.String,
            },
            "language": &graphql.Field{
                Type: graphql.String,
            },
        },
    })

    var accountEventType = graphql.NewObject(graphql.ObjectConfig{
        Name: "AccountEvent",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.String,
            },
            "eventName": &graphql.Field{
                Type: graphql.String,
            },
            "created": &graphql.Field{
                Type: graphql.String,
            },
        },
    })

    var accountType = graphql.NewObject(graphql.ObjectConfig{
        Name: "Account",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.String,
            },
            "name": &graphql.Field{
                Type: graphql.String,
            },
            "servedBy": &graphql.Field{
                Type: graphql.String,
            },
            "quote": &graphql.Field{
                Type: quoteType,
                Args: graphql.FieldConfigArgument{
                    "language": &graphql.ArgumentConfig{
                        Type:        graphql.String,
                        Description: "Two letter ISO language code such as en or sv",
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    logrus.Infof("ENTER - resolve function for quote with params %v", p.Args)
                    account := p.Source.(internalmodel.Account)
                    if account.Quote.Language == p.Args["language"] {
                        return account.Quote, nil
                    }
                    return nil, nil
                },
            },
            "events": &graphql.Field{
                Type: graphql.NewList(accountEventType),
                Args: graphql.FieldConfigArgument{
                    "eventName": &graphql.ArgumentConfig{
                        Type: graphql.String,
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    logrus.Infof("ENTER - resolve function for events with params %v", p.Args)
                    account := p.Source.(internalmodel.Account)
                    pred := func(i interface{}) bool {
                        return i.(model.AccountEvent).EventName == p.Args["eventName"]
                    }

                    response := make([]model.AccountEvent, 0)
                    for _, item := range account.AccountEvents {
                        if pred(item) {
                            response = append(response, item)
                        }
                    }
                    return response, nil
                },
            },
        },
    })

    // Schema
    fields := graphql.Fields{
        "Account": &graphql.Field{
            Type: graphql.Type(accountType),
            Args: graphql.FieldConfigArgument{
                "id": &graphql.ArgumentConfig{
                    Type: graphql.String,
                },
                "name": &graphql.ArgumentConfig{
                    Type: graphql.String,
                },
            },
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                logrus.Infof("ENTER - resolve function for Account with params %v", p.Args)
                id, _ := p.Args["id"].(string)
                for _, account := range accounts {
                    if account.ID == id {
                        return account, nil
                    }
                }
                return nil, fmt.Errorf("No account found matching ID %v", id)
            },
        },
        "AllAccounts": &graphql.Field{
            Type: graphql.NewList(accountType),
            Args: graphql.FieldConfigArgument{
                "id": &graphql.ArgumentConfig{
                    Type: graphql.String,
                },
                "name": &graphql.ArgumentConfig{
                    Type: graphql.String,
                },
            },
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                logrus.Infof("ENTER - resolve function for AllAccounts with params %v", p.Args)
                return accounts, nil
            },
        },
    }

    rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
    schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
    var err error
    Schema, err = graphql.NewSchema(schemaConfig)
    if err != nil {
        log.Fatalf("failed to create new schema, error: %v", err)
    }
    logrus.Infoln("Successfully initialized GraphQL")
    schemaInitialized = true
}

func GetAccountQL(resp http.ResponseWriter, req *http.Request) {
    logrus.Infoln("ENTER - GetAccountQL")
    query := req.URL.Query()["query"]

    params := graphql.Params{Schema: Schema, RequestString: query[0]}
    r := graphql.Do(params)
    if len(r.Errors) > 0 {
        logrus.Errorf("failed to execute graphql operation, errors: %+v", r.Errors)
        log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
    }
    rJSON, _ := json.Marshal(r)
    writeJSONResponse(resp, 200, rJSON)
}
